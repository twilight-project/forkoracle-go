package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/btcsuite/btcd/wire"
	"github.com/spf13/viper"
)

func generateSweepTx(sweepAddress SweepAddress, newSweepAddress string, accountName string, withdrawals []BtcWithdrawRequest) (string, string, uint64, error) {
	utxos := queryUtxo(sweepAddress.Address)
	if len(utxos) <= 0 {
		// need to decide if this needs to be enabled
		// addr := generateAndRegisterNewAddress(accountName, height+noOfMultisigs, sweepAddress.Address)
		fmt.Println("INFO : No funds in address : ", sweepAddress.Address, " generating new address : ")
		markAddressSignedRefund(sweepAddress.Address)
		markAddressSignedSweep(sweepAddress.Address)
		markAddressArchived(sweepAddress.Address)
		return "", "", 0, nil
	}

	totalAmountTxIn := uint64(0)
	totalAmountTxOut := uint64(0)

	sweepTx := wire.NewMsgTx(wire.TxVersion)
	for _, utxo := range utxos {
		txIn, err := CreateTxIn(utxo)
		if err != nil {
			fmt.Println("error while add tx in : ", err)
			return "", "", 0, err
		}
		totalAmountTxIn = totalAmountTxIn + utxo.Amount
		txIn.Sequence = wire.MaxTxInSequenceNum - 10
		sweepTx.AddTxIn(txIn)
	}

	for _, withdrawal := range withdrawals {
		amount, err := strconv.Atoi(withdrawal.WithdrawAmount)
		txOut, err := CreateTxOut(withdrawal.WithdrawAddress, int64(amount))
		if err != nil {
			fmt.Println("error while txout : ", err)
			return "", "", 0, err
		}
		totalAmountTxOut = totalAmountTxOut + uint64(amount)
		sweepTx.AddTxOut(txOut)
	}

	// need to be worked on

	fee := 15000

	if int64(totalAmountTxIn-totalAmountTxOut-uint64(fee)) > 0 {
		txOut, err := CreateTxOut(newSweepAddress, int64(totalAmountTxIn-totalAmountTxOut-uint64(fee)))
		if err != nil {
			log.Println("error with txout", err)
			return "", "", 0, err
		}
		sweepTx.AddTxOut(txOut)
	}

	script := querySweepAddressScript(sweepAddress.Address)
	witness := wire.TxWitness{}
	witness = append(witness, script)
	sweepTx.TxIn[0].Witness = witness

	var UnsignedTx bytes.Buffer
	sweepTx.Serialize(&UnsignedTx)
	hexTx := hex.EncodeToString(UnsignedTx.Bytes())
	fmt.Println("transaction UnSigned Sweep: ", hexTx)
	txid := sweepTx.TxHash().String()

	return hexTx, txid, totalAmountTxIn, nil
}

func generateRefundTx(txHex string, address string) (string, error) {
	sweepTx, err := createTxFromHex(txHex)
	if err != nil {
		fmt.Println("error decoding tx : ", err)
	}

	inputTx := sweepTx.TxHash().String()
	vout := len(sweepTx.TxOut) - 1

	utxo := Utxo{
		inputTx,
		uint32(vout),
		0,
	}
	refundTx := wire.NewMsgTx(wire.TxVersion)
	txIn, err := CreateTxIn(utxo)
	if err != nil {
		fmt.Println("error while add tx in : ", err)
		return "", err
	}
	txIn.Sequence = wire.MaxTxInSequenceNum - 10
	refundTx.AddTxIn(txIn)

	// need to be decided
	txout, err := CreateTxOut("bc1q49kzd05aqxs8q7r4rnnxc35cdk6783sf0khepr", 5000)
	if err != nil {
		fmt.Println("error while add tx out : ", err)
		return "", err
	}

	refundTx.AddTxOut(txout)
	locktime := uint32(5)
	refundTx.LockTime = locktime

	script := querySweepAddressScript(address)
	witness := wire.TxWitness{}
	witness = append(witness, script)
	refundTx.TxIn[0].Witness = witness

	var UnsignedTx bytes.Buffer
	refundTx.Serialize(&UnsignedTx)
	hexTx := hex.EncodeToString(UnsignedTx.Bytes())
	fmt.Println("transaction UnSigned Refund: ", hexTx)

	return hexTx, nil
}

func generateSignedTxs(address string, accountName string, sweepTx *wire.MsgTx, refundTx *wire.MsgTx) ([]byte, []byte, error) {

	number := fmt.Sprintf("%v", viper.Get("no_of_validators"))
	noOfValidators, _ := strconv.Atoi(number)
	for {
		time.Sleep(30 * time.Second)
		receiveSweepSignatures := getSignSweep()
		receiveRefundSignatures := getSignRefund()

		addrs := querySweepAddress(address)
		if len(addrs) <= 0 {
			continue
		}
		currentReserveAddress := addrs[0]

		addrs = querySweepAddressByParentAddress(currentReserveAddress.Address)
		if len(addrs) <= 0 {
			continue
		}
		newReserveAddress := addrs[0]

		filteredSweepSignatures := filterSignSweep(receiveSweepSignatures, currentReserveAddress.Address)
		filteredRefundSignatures := filterSignRefund(receiveRefundSignatures, newReserveAddress.Address)

		if len(filteredSweepSignatures) <= 0 {
			continue
		}

		if len(filteredRefundSignatures) <= 0 {
			continue
		}

		minSignsRequired := noOfValidators * 2 / 3

		// remove this when I redo the chain
		// minSignsRequired = 3

		if len(filteredSweepSignatures)/minSignsRequired < 1 {
			fmt.Println("INFO: ", "not enough sweep signatures")
			continue
		}

		if len(filteredRefundSignatures)/minSignsRequired < 1 {
			fmt.Println("INFO: ", "not enough refund signatures")
			continue
		}

		script := currentReserveAddress.Script
		preimage := currentReserveAddress.Preimage

		for i, _ := range sweepTx.TxIn {
			dataSig := make([][]byte, 0)
			for _, sig := range filteredSweepSignatures {
				signatures := strings.Split(sig.SweepSignature, ",")
				for _, signature := range signatures {
					sig, _ := hex.DecodeString(signature)
					dataSig = append(dataSig, sig)
				}
			}
			witness := wire.TxWitness{}
			witness = append(witness, preimage)
			dummy := []byte{}
			witness = append(witness, dummy)
			for j := 0; j < minSignsRequired; j++ {
				witness = append(witness, dataSig[j])
			}
			witness = append(witness, script)
			sweepTx.TxIn[i].Witness = witness
		}
		var signedSweepTx bytes.Buffer
		err := sweepTx.Serialize(&signedSweepTx)
		if err != nil {
			log.Fatal(err)
		}

		dataSig := make([][]byte, 0)

		for _, sig := range filteredRefundSignatures {
			sig, _ := hex.DecodeString(sig.RefundSignature)
			dataSig = append(dataSig, sig)
		}

		script = newReserveAddress.Script
		judgeSign := signByJudge(refundTx, script)
		judgeSignature, _ := hex.DecodeString(judgeSign[0])

		for i := 0; i < len(refundTx.TxIn); i++ {

			witness := wire.TxWitness{}
			witness = append(witness, judgeSignature)
			dummy := []byte{}
			witness = append(witness, dummy)
			for j := 0; j < minSignsRequired; j++ {
				witness = append(witness, dataSig[j])
			}
			witness = append(witness, script)
			refundTx.TxIn[i].Witness = witness
		}

		var signedRefundTx bytes.Buffer
		err = refundTx.Serialize(&signedRefundTx)
		if err != nil {
			log.Fatal(err)
		}

		return signedSweepTx.Bytes(), signedRefundTx.Bytes(), nil
	}
}

//temp use function judge sign

func signByJudge(tx *wire.MsgTx, script []byte) []string {
	return signTx(tx, script)
}

func initJudge(accountName string) {
	fmt.Println("init judge")
	if judge == true {
		addr := queryAllSweepAddresses()
		if len(addr) <= 0 {
			time.Sleep(2 * time.Minute)
			initReserve(accountName)
		}
	}
}

func initReserve(accountName string) {
	fmt.Println("init reserve")
	height := 0
	// number := fmt.Sprintf("%v", viper.Get("no_of_Multisigs"))
	// noOfMultisigs, _ := strconv.Atoi(number)

	number := fmt.Sprintf("%v", viper.Get("unlocking_time"))
	unlockingTimeInBlocks, _ := strconv.Atoi(number)

	//TODO Need to change this for multi judge setup
	judges := getRegisteredJudges()
	if len(judges.Judges) == 0 {
		registerJudge(accountName)
	}

	for {
		resp := getAttestations("1")
		if len(resp.Attestations) <= 0 {
			time.Sleep(30)
			continue
		} else {
			attestation := resp.Attestations[0]
			btc_height, err := strconv.Atoi(attestation.Proposal.Height)
			if err != nil {
				fmt.Println("Error: converting to int : ", err)
				continue
			}
			height = btc_height
			break
		}
	}

	if height > 0 {

		// for i := 1; i <= noOfMultisigs; i++ {
		_ = generateAndRegisterNewAddress(accountName, height+unlockingTimeInBlocks, "")
		// 	height = height + 1
		// }
	}

	fmt.Println("judge initialized")
}

func startJudge(accountName string) {
	fmt.Println("starting judge")
	// number := fmt.Sprintf("%v", viper.Get("no_of_Multisigs"))
	// noOfMultisigs, _ := strconv.Atoi(number)

	number := fmt.Sprintf("%v", viper.Get("unlocking_time"))
	unlockingTimeInBlocks, _ := strconv.Atoi(number)
	for {
		resp := getAttestations("20")
		if len(resp.Attestations) <= 0 {
			time.Sleep(1 * time.Minute)
			fmt.Println("no attestaions (start judge)")
			continue
		}

		for _, attestation := range resp.Attestations {
			if attestation.Observed == false {
				continue
			}

			height, _ := strconv.Atoi(attestation.Proposal.Height)
			addresses := querySweepAddressesByHeight(uint64(height))

			if len(addresses) <= 0 {
				continue
			} else {
				fmt.Println("INFO: sweep address found for btc height : ", attestation.Proposal.Height)
				sweepAddress := addresses[0]

				newSweepAddress := generateAndRegisterNewAddress(accountName, height+unlockingTimeInBlocks, sweepAddress.Address)
				withdrawals := getBtcWithdrawRequestForAddress(sweepAddress)
				sweepTxHex, sweepTxId, _, err := generateSweepTx(sweepAddress, newSweepAddress, accountName, withdrawals)
				if err != nil {
					fmt.Println("Error in generating a Sweep transaction: ", err)
					continue
				}
				if sweepTxHex == "" {
					fmt.Println("INFO: ", "no sweep tx generated because no funds in current address")
					time.Sleep(1 * time.Minute)
					continue
				}

				reserve := getReserveForAddress(sweepAddress.Address)
				reserveID, err := strconv.Atoi(reserve.ReserveId)
				if err != nil {
					fmt.Println("Error:", err)
					return
				}

				refundTxHex, err := generateRefundTx(sweepTxHex, newSweepAddress)

				sendUnsignedSweepTx(sweepTxHex, sweepTxId, accountName)
				sendUnsignedRefundTx(refundTxHex, uint64(reserveID), accountName)

				// sleep time so that other validators can sign
				time.Sleep(1 * time.Minute)

				sweeptx, err := createTxFromHex(sweepTxHex)
				if err != nil {
					fmt.Println("error decoding sweep tx : inside judge")
					fmt.Println(err)
				}

				refundTx, err := createTxFromHex(refundTxHex)
				if err != nil {
					fmt.Println("error decoding refund tx : inside judge")
					fmt.Println(err)
				}

				signedSweepTx, signedRefundTx, err := generateSignedTxs(sweepAddress.Address, accountName, sweeptx, refundTx)
				if err != nil {
					fmt.Println(err)
				}

				signedSweepTxHex := hex.EncodeToString(signedSweepTx)
				fmt.Println("Signed P2WSH Sweep transaction with preimage:", signedSweepTxHex)

				signedRefundTxHex := hex.EncodeToString(signedRefundTx)
				fmt.Println("Signed P2WSH Refund transaction with preimage:", signedRefundTxHex)

				broadcastSweeptxNYKS(signedSweepTxHex, accountName)

				wireTransaction, err := createTxFromHex(signedSweepTxHex)
				if err != nil {
					fmt.Println("error decodeing signed transaction : ", err)
				}
				broadcastBtcTransaction(wireTransaction)
				markAddressArchived(sweepAddress.Address)

			}

		}

	}
}
