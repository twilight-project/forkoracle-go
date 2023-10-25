package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/btcsuite/btcd/wire"
	"github.com/spf13/viper"
	bridgetypes "github.com/twilight-project/nyks/x/bridge/types"
)

func generateSweepTx(sweepAddress SweepAddress, newSweepAddress string, accountName string, withdrawals []BtcWithdrawRequest, unlockHeight int64) (string, string, uint64, error) {
	number := fmt.Sprintf("%v", viper.Get("sweep_preblock"))
	sweepPreblock, _ := strconv.Atoi(number)
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
		txIn.Sequence = uint32(unlockHeight)
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

	if int64(totalAmountTxIn-totalAmountTxOut) > 0 {
		txOut, err := CreateTxOut(newSweepAddress, int64(totalAmountTxIn-totalAmountTxOut))
		if err != nil {
			log.Println("error with txout", err)
			return "", "", 0, err
		}
		sweepTx.AddTxOut(txOut)
	}

	feeRate := getBtcFeeRate()
	baseSize := sweepTx.SerializeSizeStripped()
	totalSize := sweepTx.SerializeSize()
	weight := (baseSize * 3) + totalSize
	vsize := (weight + 3) / 4

	// Calculate the required fee
	requiredFee := vsize * feeRate.Priority

	lastOutput := sweepTx.TxOut[len(sweepTx.TxOut)-1]
	if lastOutput.Value < int64(requiredFee) {
		log.Fatalf("Change output is smaller than required fee")
	}

	// Deduct the fee from the change output
	lastOutput.Value = lastOutput.Value - int64(requiredFee)
	sweepTx.TxOut[len(sweepTx.TxOut)-1] = lastOutput

	script := querySweepAddressScript(sweepAddress.Address)
	witness := wire.TxWitness{}
	witness = append(witness, script)
	sweepTx.TxIn[0].Witness = witness
	sweepTx.LockTime = uint32(unlockHeight + int64(sweepPreblock))

	var UnsignedTx bytes.Buffer
	sweepTx.Serialize(&UnsignedTx)
	hexTx := hex.EncodeToString(UnsignedTx.Bytes())
	fmt.Println("transaction UnSigned Sweep: ", hexTx)
	txid := sweepTx.TxHash().String()

	return hexTx, txid, totalAmountTxIn, nil
}

func generateRefundTx(txHex string, script string, reserveId uint64) (string, error) {
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

	clearingAccounts := getClearingAccounts(uint64(reserveId))
	for _, cAccount := range clearingAccounts.ReserveClearingAccountsAll {
		for _, balance := range cAccount.ReserveAccountBalances {
			amount, _ := strconv.Atoi(balance.Amount)
			txout, err := CreateTxOut(cAccount.BtcWithdrawAddress, int64(amount))
			if err != nil {
				fmt.Println("error while add tx out : ", err)
				return "", err
			}

			refundTx.AddTxOut(txout)
		}
	}

	feeRate := getBtcFeeRate()
	baseSize := sweepTx.SerializeSizeStripped()
	totalSize := sweepTx.SerializeSize()
	weight := (baseSize * 3) + totalSize
	vsize := (weight + 3) / 4

	// Calculate the required fee
	requiredFee := vsize * feeRate.Priority

	validators := getDelegateAddresses()

	feeAdjustment := requiredFee / len(validators.Addresses)

	for i, output := range refundTx.TxOut {
		refundTx.TxOut[i].Value = output.Value - int64(feeAdjustment)
	}
	// need to be decided
	// txout, err := CreateTxOut("bc1q49kzd05aqxs8q7r4rnnxc35cdk6783sf0khepr", 5000)
	// if err != nil {
	// 	fmt.Println("error while add tx out : ", err)
	// 	return "", err
	// }

	// refundTx.AddTxOut(txout)
	locktime := uint32(5)
	refundTx.LockTime = locktime

	scriptbytes, _ := hex.DecodeString(script)
	witness := wire.TxWitness{}
	witness = append(witness, scriptbytes)
	refundTx.TxIn[0].Witness = witness

	var UnsignedTx bytes.Buffer
	refundTx.Serialize(&UnsignedTx)
	hexTx := hex.EncodeToString(UnsignedTx.Bytes())
	fmt.Println("transaction UnSigned Refund: ", hexTx)

	return hexTx, nil
}

func generateSignedSweepTx(accountName string, sweepTx *wire.MsgTx, reserveId uint64, roundId uint64, currentReserveAddress SweepAddress) []byte {

	number := fmt.Sprintf("%v", viper.Get("no_of_validators"))
	noOfValidators, _ := strconv.Atoi(number)
	for {
		time.Sleep(30 * time.Second)
		receiveSweepSignatures := getSignSweep(reserveId, roundId)

		filteredSweepSignatures := orderSignSweep(receiveSweepSignatures)

		if len(filteredSweepSignatures) <= 0 {
			continue
		}

		minSignsRequired := noOfValidators * 2 / 3

		if len(filteredSweepSignatures)/minSignsRequired < 1 {
			fmt.Println("INFO: ", "not enough sweep signatures")
			continue
		}

		script := currentReserveAddress.Script
		preimage := currentReserveAddress.Preimage

		for i, _ := range sweepTx.TxIn {
			dataSig := make([][]byte, 0)
			for _, sig := range filteredSweepSignatures {
				sig, _ := hex.DecodeString(sig.SweepSignature[i])
				dataSig = append(dataSig, sig)
			}
			witness := wire.TxWitness{}
			witness = append(witness, preimage)
			dummy := []byte{}
			witness = append(witness, dummy)
			for j := 0; j < minSignsRequired; j++ {
				witness = append(witness, dataSig[j])
			}

			// buf := make([]byte, 8)
			// binary.BigEndian.PutUint64(buf, uint64(currentReserveAddress.Unlock_height))

			// witness = append(witness, buf)
			witness = append(witness, script)
			sweepTx.TxIn[i].Witness = witness
		}
		var signedSweepTx bytes.Buffer
		err := sweepTx.Serialize(&signedSweepTx)
		if err != nil {
			log.Fatal(err)
		}

		return signedSweepTx.Bytes()
	}
}

func generateSignedRefundTx(accountName string, refundTx *wire.MsgTx, reserveId uint64, roundId uint64) ([]byte, error, SweepAddress) {

	number := fmt.Sprintf("%v", viper.Get("no_of_validators"))
	noOfValidators, _ := strconv.Atoi(number)

	addrs := getProposedSweepAddress(reserveId, roundId)
	if addrs.ProposeSweepAddressMsg.BtcAddress == "" {
		return nil, nil, SweepAddress{}
	}

	addresses := querySweepAddress(addrs.ProposeSweepAddressMsg.BtcAddress)
	if len(addresses) <= 0 {
		fmt.Println("address not found in DB")
		return nil, nil, SweepAddress{}
	}
	newReserveAddress := addresses[0]

	for {
		time.Sleep(30 * time.Second)
		receiveRefundSignatures := getSignRefund(reserveId, roundId)
		filteredRefundSignatures, JudgeSign := OrderSignRefund(receiveRefundSignatures, newReserveAddress.Address)

		minSignsRequired := noOfValidators * 2 / 3

		if len(filteredRefundSignatures) <= 0 {
			continue
		}

		if len(filteredRefundSignatures)/minSignsRequired < 1 {
			fmt.Println("INFO: ", "not enough refund signatures")
			continue
		}

		dataSig := make([][]byte, 0)

		for _, sig := range filteredRefundSignatures {
			sig, _ := hex.DecodeString(sig.RefundSignature)
			dataSig = append(dataSig, sig)
		}

		script := newReserveAddress.Script
		preimageFalse, _ := preimage()

		for i := 0; i < len(refundTx.TxIn); i++ {

			witness := wire.TxWitness{}
			witness = append(witness, []byte(JudgeSign.RefundSignature))
			witness = append(witness, preimageFalse)
			dummy := []byte{}
			witness = append(witness, dummy)
			for j := 0; j < minSignsRequired; j++ {
				witness = append(witness, dataSig[j])
			}
			witness = append(witness, script)
			refundTx.TxIn[i].Witness = witness
		}

		var signedRefundTx bytes.Buffer
		err := refundTx.Serialize(&signedRefundTx)
		if err != nil {
			log.Fatal(err)
		}

		return signedRefundTx.Bytes(), nil, newReserveAddress
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

	number := fmt.Sprintf("%v", viper.Get("unlocking_time"))
	unlockingTimeInBlocks, _ := strconv.Atoi(number)

	judges := getRegisteredJudges()
	if len(judges.Judges) == 0 {
		registerJudge(accountName)
	} else {
		registered := false
		for _, judge := range judges.Judges {
			if judge.JudgeAddress == oracleAddr {
				registered = true
			}
		}
		if registered == false {
			registerJudge(accountName)
		}
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

	_ = generateAndRegisterNewBtcReserveAddress(accountName, int64(height+unlockingTimeInBlocks))

	fmt.Println("judge initialized")
}

func addressProposer(accountName string) {
	fmt.Println("starting address proposer")
	number := fmt.Sprintf("%v", viper.Get("unlocking_time"))
	unlockingTimeInBlocks, _ := strconv.Atoi(number)

	// temporary till staking is implemented
	number = fmt.Sprintf("%v", viper.Get("height_diff_between_judges"))
	heightDiffBetweenJudges, _ := strconv.Atoi(number)

	for {
		time.Sleep(5 * time.Minute)

		var latestProposedAddress SweepAddress
		addresses := querySweepAddressesOrderByHeight(1)
		if len(addresses) == 0 {
			fmt.Println("no Sweep address found")
			continue
		}

		latestProposedAddress = addresses[0]

		var currentJudgeReserves []BtcReserve
		btcReserves := getBtcReserves()
		for _, reserve := range btcReserves.BtcReserves {
			if reserve.JudgeAddress == oracleAddr {
				currentJudgeReserves = append(currentJudgeReserves, reserve)
				break
			}
		}

		if len(currentJudgeReserves) != 1 {
			fmt.Println("current judge address not found")
			continue
		}

		currentJudgeReserve := currentJudgeReserves[0]

		reserveIdForProposal, _ := strconv.Atoi(currentJudgeReserve.ReserveId)
		if reserveIdForProposal == 1 {
			reserveIdForProposal = len(btcReserves.BtcReserves)
		} else {
			reserveIdForProposal = reserveIdForProposal - 1
		}

		var reserveToBeUpdated BtcReserve
		for _, reserve := range btcReserves.BtcReserves {
			tempId, _ := strconv.Atoi(reserve.ReserveId)
			if tempId == reserveIdForProposal {
				reserveToBeUpdated = reserve
				break
			}
		}

		RoundId, _ := strconv.Atoi(reserveToBeUpdated.RoundId)
		proposed := checkIfAddressIsProposed(int64(RoundId+1), int64(reserveIdForProposal))
		if proposed == true {
			continue
		}

		unlockHeight := latestProposedAddress.Unlock_height + int64(unlockingTimeInBlocks) + int64(heightDiffBetweenJudges)
		newReserveAddress, hexScript := generateAndRegisterNewProposedAddress(accountName, unlockHeight, currentJudgeReserve.ReserveAddress)

		cosmos_client := getCosmosClient()
		msg := &bridgetypes.MsgProposeSweepAddress{
			BtcScript:    hexScript,
			BtcAddress:   newReserveAddress,
			JudgeAddress: oracleAddr,
			ReserveId:    uint64(reserveIdForProposal),
			RoundId:      uint64(RoundId + 1),
		}

		sendTransactionSweepAddressProposal(accountName, cosmos_client, msg)

		insertProposedAddress(reserveToBeUpdated.ReserveAddress, newReserveAddress, unlockHeight, int64(RoundId+1), int64(reserveIdForProposal))
	}
}

func processSweep(accountName string) {
	fmt.Println("Process Sweep unsigned started")
	number := fmt.Sprintf("%v", viper.Get("sweep_preblock"))
	sweepInitateBlockHeight, _ := strconv.Atoi(number)

	for {
		resp := getAttestations("20")
		if len(resp.Attestations) <= 0 {
			time.Sleep(1 * time.Minute)
			fmt.Println("no attestaions (start judge)")
			continue
		}

		var currentReservesForThisJudge []BtcReserve
		reserves := getBtcReserves()
		for _, reserve := range reserves.BtcReserves {
			if reserve.JudgeAddress == oracleAddr {
				currentReservesForThisJudge = append(currentReservesForThisJudge, reserve)
			}
		}

		if len(currentReservesForThisJudge) == 0 {
			time.Sleep(2 * time.Minute)
			continue
		}

		for _, attestation := range resp.Attestations {
			height, _ := strconv.Atoi(attestation.Proposal.Height)

			if attestation.Observed == false {
				continue
			}

			addresses := querySweepAddressesByHeight(uint64(height + sweepInitateBlockHeight))
			if len(addresses) <= 0 {
				continue
			}

			fmt.Println("sweep address found")

			currentSweepAddress := addresses[0]
			var newSweepAddress *string
			var reserveTobeProcessed BtcReserve

			minRoundId := 500
			// Iterate through the array and find the minimum roundId
			for _, reserve := range currentReservesForThisJudge {
				tempRoundId, _ := strconv.Atoi(reserve.RoundId)
				if tempRoundId < minRoundId {
					minRoundId = tempRoundId
					reserveTobeProcessed = reserve
				}
			}

			currentRoundId, _ := strconv.Atoi(reserveTobeProcessed.RoundId)
			currentReserveId, _ := strconv.Atoi(reserveTobeProcessed.ReserveId)

			for {
				sweepAddresses := getProposedSweepAddress(uint64(currentReserveId), uint64(currentRoundId+1))
				if sweepAddresses.ProposeSweepAddressMsg.BtcAddress == "" {
					fmt.Println("no proposed sweep address found ")
					time.Sleep(2 * time.Minute)
					continue
				}
				newSweepAddress = &sweepAddresses.ProposeSweepAddressMsg.BtcAddress
				break
			}

			withdrawals := getBtcWithdrawRequestForAddress(currentSweepAddress)
			sweepTxHex, sweepTxId, _, err := generateSweepTx(currentSweepAddress, *newSweepAddress, accountName, withdrawals, int64(height))
			if err != nil {
				fmt.Println("Error in generating a Sweep transaction: ", err)
				return
			}
			if sweepTxHex == "" {
				fmt.Println("INFO: ", "no sweep tx generated because no funds in current address")
				time.Sleep(1 * time.Minute)
				return
			}

			sendUnsignedSweepTx(uint64(currentReserveId), uint64(currentRoundId+1), sweepTxHex, sweepTxId, accountName)
			markAddressArchived(currentSweepAddress.Address)
		}
	}
}

func processRefund(accountName string) {
	fmt.Println("Process Refund unsigned started")
	for {
		reserves := getBtcReserves()
		var currentReservesForThisJudge []BtcReserve
		for _, reserve := range reserves.BtcReserves {
			if reserve.JudgeAddress == oracleAddr {
				currentReservesForThisJudge = append(currentReservesForThisJudge, reserve)
			}
		}

		var reserveTobeProcessed *BtcReserve
		minRoundId := 500
		// Iterate through the array and find the minimum roundId
		for _, reserve := range currentReservesForThisJudge {
			tempRoundId, _ := strconv.Atoi(reserve.RoundId)
			if tempRoundId < minRoundId {
				minRoundId = tempRoundId
				reserveTobeProcessed = &reserve
			}
		}

		var currentRoundId int
		var currentReserveId int
		if reserveTobeProcessed == nil {
			reserveTobeProcessed = &reserves.BtcReserves[len(reserves.BtcReserves)]
			currentReserveId = 1
		} else {
			currentReserveId, _ = strconv.Atoi(reserveTobeProcessed.ReserveId)

		}
		currentRoundId, _ = strconv.Atoi(reserveTobeProcessed.RoundId)

		var reserveIdForRefund int
		if currentReserveId == 1 {
			reserveIdForRefund = len(reserves.BtcReserves)
		} else {
			reserveIdForRefund = currentReserveId - 1
		}

		sweepTxs := getUnsignedSweepTx(uint64(reserveIdForRefund), uint64(currentRoundId+1))
		if len(sweepTxs.UnsignedTxSweepMsg) == 0 {
			continue
		}

		sweeptx := sweepTxs.UnsignedTxSweepMsg[0]

		sweepAddresses := getProposedSweepAddress(uint64(reserveIdForRefund), uint64(currentRoundId+1))
		if sweepAddresses.ProposeSweepAddressMsg.BtcAddress == "" {
			continue
		}

		refundTxHex, _ := generateRefundTx(sweeptx.BtcUnsignedSweepTx, sweepAddresses.ProposeSweepAddressMsg.BtcScript, uint64(reserveIdForRefund))
		sendUnsignedRefundTx(refundTxHex, uint64(reserveIdForRefund), uint64(currentRoundId+1), accountName)
	}
}

func processSweepSigning(accountName string) {
	fmt.Println("Process Sweep signed started")
	for {
		var currentReservesForThisJudge []BtcReserve
		reserves := getBtcReserves()
		for _, reserve := range reserves.BtcReserves {
			if reserve.JudgeAddress == oracleAddr {
				currentReservesForThisJudge = append(currentReservesForThisJudge, reserve)
			}
		}

		var reserveTobeProcessed BtcReserve
		minRoundId := 500
		// Iterate through the array and find the minimum roundId
		for _, reserve := range currentReservesForThisJudge {
			tempRoundId, _ := strconv.Atoi(reserve.RoundId)
			if tempRoundId < minRoundId {
				minRoundId = tempRoundId
				reserveTobeProcessed = reserve
			}
		}

		reserveIdForSweepTx, _ := strconv.Atoi(reserveTobeProcessed.ReserveId)
		roundIdForSweepTx, _ := strconv.Atoi(reserveTobeProcessed.RoundId)
		roundIdForSweepTx = roundIdForSweepTx + 1

		sweepTxs := getUnsignedSweepTx(uint64(reserveIdForSweepTx), uint64(roundIdForSweepTx))

		fmt.Println("==============================")
		fmt.Println("current reserves : ", currentReservesForThisJudge)
		fmt.Println("Unsigned Sweep Txs : ", sweepTxs)
		fmt.Println("==============================")

		continue

		if len(sweepTxs.UnsignedTxSweepMsg) == 0 {
			continue
		}

		fmt.Println("Signed Sweep process : Unsigned Sweep Transaction found")

		unsignedSweepTxHex := sweepTxs.UnsignedTxSweepMsg[0].BtcUnsignedSweepTx
		sweepTx, err := createTxFromHex(unsignedSweepTxHex)
		if err != nil {
			fmt.Println("error decoding sweep tx : inside judge")
			fmt.Println(err)
		}

		reserveAddresses := queryUnsignedSweepAddressByScript(sweepTx.TxIn[0].Witness[0])

		if len(reserveAddresses) == 0 {
			fmt.Println("No address found")
			continue
		}
		currentReserveAddress := reserveAddresses[0]

		if currentReserveAddress.BroadcastSweep == true {
			continue
		}

		signedSweepTx := generateSignedSweepTx(accountName, sweepTx, uint64(reserveIdForSweepTx), uint64(roundIdForSweepTx), currentReserveAddress)

		signedSweepTxHex := hex.EncodeToString(signedSweepTx)
		fmt.Println("Signed P2WSH Sweep transaction with preimage:", signedSweepTxHex)

		broadcastSweeptxNYKS(signedSweepTxHex, accountName, uint64(reserveIdForSweepTx), uint64(roundIdForSweepTx))

		wireTransaction, err := createTxFromHex(signedSweepTxHex)
		if err != nil {
			fmt.Println("error decodeing signed transaction : ", err)
		}
		broadcastBtcTransaction(wireTransaction)
		markAddressBroadcastedSweep(currentReserveAddress.Address)
		time.Sleep(3 * time.Minute)
	}
}

func processRefundSigning(accountName string) {
	fmt.Println("Process Refund signed started")
	for {
		reserves := getBtcReserves()
		var currentReservesForThisJudge []BtcReserve
		for _, reserve := range reserves.BtcReserves {
			if reserve.JudgeAddress == oracleAddr {
				currentReservesForThisJudge = append(currentReservesForThisJudge, reserve)
			}
		}

		var reserveTobeProcessed *BtcReserve
		minRoundId := 500
		// Iterate through the array and find the minimum roundId
		for _, reserve := range currentReservesForThisJudge {
			tempRoundId, _ := strconv.Atoi(reserve.RoundId)
			if tempRoundId < minRoundId {
				minRoundId = tempRoundId
				reserveTobeProcessed = &reserve
			}
		}

		var currentRoundId int
		var currentReserveId int
		if reserveTobeProcessed == nil {
			reserveTobeProcessed = &reserves.BtcReserves[len(reserves.BtcReserves)]
			currentReserveId = 1
		} else {
			currentReserveId, _ = strconv.Atoi(reserveTobeProcessed.ReserveId)

		}
		currentRoundId, _ = strconv.Atoi(reserveTobeProcessed.RoundId)

		var reserveIdForSweep int
		if currentReserveId == 1 {
			reserveIdForSweep = len(reserves.BtcReserves)
		} else {
			reserveIdForSweep = currentReserveId - 1
		}

		refundTxs := getUnsignedRefundTx(int64(reserveIdForSweep), int64(currentRoundId+1))

		if len(refundTxs.UnsignedTxRefundMsg) == 0 {
			continue
		}

		unsignedRefundTxHex := refundTxs.UnsignedTxRefundMsg[0].BtcUnsignedRefundTx
		refundTx, err := createTxFromHex(unsignedRefundTxHex)
		if err != nil {
			fmt.Println("error decoding sweep tx : inside judge")
			fmt.Println(err)
		}

		signedRefundTx, _, newReserveAddress := generateSignedRefundTx(accountName, refundTx, uint64(reserveIdForSweep), uint64(currentRoundId+1))

		signedRefundTxHex := hex.EncodeToString(signedRefundTx)
		fmt.Println("Signed P2WSH Refund transaction with preimage:", signedRefundTxHex)

		broadcastRefundtxNYKS(signedRefundTxHex, accountName, uint64(currentReserveId), uint64(currentRoundId+1))
		markAddressBroadcastedRefund(newReserveAddress.Address)

		// add tapscript inscription here

		time.Sleep(1 * time.Hour)
	}
}

func startJudge(accountName string) {
	fmt.Println("starting judge")
	go addressProposer(accountName)
	go processSweep(accountName)
	go processSweepSigning(accountName)
	go processRefund(accountName)
	go processRefundSigning(accountName)
}
