package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/spf13/viper"
	bridgetypes "github.com/twilight-project/nyks/x/bridge/types"
	volttypes "github.com/twilight-project/nyks/x/volt/types"
)

func registerReserveAddressOnNyks(accountName string, address string, script []byte) {

	cosmos := getCosmosClient()

	judge_address, err := cosmos.Address(accountName)
	if err != nil {
		log.Fatal(err)
	}

	reserveScript := hex.EncodeToString(script)

	msg := &bridgetypes.MsgRegisterReserveAddress{
		ReserveScript:  reserveScript,
		ReserveAddress: address,
		JudgeAddress:   judge_address.String(),
	}

	// store response in txResp
	txResp, err := cosmos.BroadcastTx(accountName, msg)
	if err != nil {
		fmt.Println(err)
	}

	// print response from broadcasting a transaction
	fmt.Print("MsgRegisterReserveAddress:\n\n")
	fmt.Println(txResp)
}

func registerAddressOnForkscanner(address string) {
	dt := time.Now().UTC()
	dt = dt.AddDate(1, 0, 0)

	request_body := map[string]interface{}{
		"method":  "add_watched_addresses",
		"id":      1,
		"jsonrpc": "2.0",
		"params": map[string]interface{}{
			"add": []interface{}{
				map[string]string{
					"address":     address,
					"watch_until": dt.Format(time.RFC3339),
				},
			},
		},
	}

	data, err := json.Marshal(request_body)
	if err != nil {
		log.Fatalf("Post: %v", err)
	}
	fmt.Println(string(data))

	resp, err := http.Post("http://0.0.0.0:8339", "application/json", strings.NewReader(string(data)))
	if err != nil {
		log.Fatalf("Post: %v", err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("ReadAll: %v", err)
	}
	result := make(map[string]interface{})
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	log.Println(result)

}

func CreateTxOut(addr string, amount int64) (*wire.TxOut, error) {
	// Decode the Bitcoin address.
	address, err := btcutil.DecodeAddress(addr, &chaincfg.MainNetParams)
	if err != nil {
		fmt.Println("Error decoding address:", err)
		return nil, err
	}

	// Generate the pay-to-address script.
	destinationAddrByte, err := txscript.PayToAddrScript(address)
	if err != nil {
		fmt.Println("Error generating pay-to-address script:", err)
		return nil, err
	}
	TxOut := wire.NewTxOut(amount, destinationAddrByte)
	return TxOut, nil

}

func CreateTxIn(utxo Utxo) (*wire.TxIn, error) {
	utxoHash, err := chainhash.NewHashFromStr(utxo.Txid)
	if err != nil {
		log.Println("error with UTXO")
		return nil, err
	}
	outPoint := wire.NewOutPoint(utxoHash, utxo.Vout)
	txIn := wire.NewTxIn(outPoint, nil, nil)
	return txIn, nil
}

func generateSweepTx(sweepAddress SweepAddress, accountName string, height int) (string, []BtcWithdrawRequest, uint64, error) {
	utxos := queryUtxo(sweepAddress.Address)
	withdrawals := getBtcWithdrawRequest()
	totalAmountTxIn := uint64(0)
	totalAmountTxOut := uint64(0)

	redeemTx := wire.NewMsgTx(wire.TxVersion)
	for _, utxo := range utxos {
		txIn, err := CreateTxIn(utxo)
		if err != nil {
			log.Println("error with txin")
			return "", nil, 0, err
		}
		totalAmountTxIn = totalAmountTxIn + utxo.Amount
		redeemTx.AddTxIn(txIn)
	}

	withdrawRequests := make([]BtcWithdrawRequest, 0)
	for _, withdrawal := range withdrawals.WithdrawRequest {
		if withdrawal.ReserveAddress == sweepAddress.Address {
			amount, err := strconv.Atoi(withdrawal.WithdrawAmount)
			txOut, err := CreateTxOut(withdrawal.WithdrawAddress, int64(amount))
			if err != nil {
				log.Println("error with txout")
				return "", nil, 0, err
			}
			totalAmountTxOut = totalAmountTxOut + uint64(amount)
			redeemTx.AddTxOut(txOut)
			withdrawRequests = append(withdrawRequests, withdrawal)
		}
	}

	fee := 5000
	number := fmt.Sprintf("%v", viper.Get("no_of_Multisigs"))
	noOfMultisigs, _ := strconv.Atoi(number)
	newSweepAddress, reserveScript := generateAddress(height + noOfMultisigs)
	registerReserveAddressOnNyks(accountName, newSweepAddress, reserveScript)
	registerAddressOnForkscanner(newSweepAddress)

	txOut, err := CreateTxOut(newSweepAddress, int64(totalAmountTxIn-totalAmountTxOut-uint64(fee)))
	if err != nil {
		log.Println("error with txout", err)
		return "", nil, 0, err
	}
	redeemTx.AddTxOut(txOut)

	var signedTx bytes.Buffer
	redeemTx.Serialize(&signedTx)
	hexTx := hex.EncodeToString(signedTx.Bytes())
	fmt.Println("transaction : ", hexTx)

	return hexTx, withdrawRequests, totalAmountTxIn, nil
}

func sendSweepProposal(tx string, address string, withdrawals []BtcWithdrawRequest, accountName string, total uint64) {

	twilightIndividualAccounts := make([]*volttypes.IndividualTwilightReserveAccount, 0)
	for _, withdrawal := range withdrawals {
		amount, _ := strconv.Atoi(withdrawal.WithdrawAmount)
		individualAccount := volttypes.IndividualTwilightReserveAccount{
			TwilightAddress: withdrawal.WithdrawAddress,
			BtcValue:        uint64(amount),
		}
		twilightIndividualAccounts = append(twilightIndividualAccounts, &individualAccount)
	}

	cosmos := getCosmosClient()
	judge_address, err := cosmos.Address(accountName)
	if err != nil {
		log.Fatal(err)
	}

	msg := &bridgetypes.MsgSweepProposal{

		ReserveId:                        1,
		ReserveAddress:                   address,
		JudgeAddress:                     judge_address.String(),
		BtcRelayCapacityValue:            0,
		TotalValue:                       total,
		PrivatePoolValue:                 0,
		PublicValue:                      0,
		FeePool:                          0,
		IndividualTwilightReserveAccount: twilightIndividualAccounts,
		BtcRefundTx:                      tx, // change to refund tx
		BtcSweepTx:                       tx,
	}

	sendTransactionSweepProposal(accountName, cosmos, msg)
}

func sendSweepSign(hexSignatures string, address string, accountName string) {
	cosmos := getCosmosClient()
	cosmos_address := getCosmosAddress(accountName, cosmos)
	msg := &bridgetypes.MsgSignSweep{
		ReserveAddress:   address,
		SignerAddress:    address, // no idea what this is
		SweepSignature:   hexSignatures,
		BtcOracleAddress: cosmos_address.String(),
	}

	sendTransactionSignSweep(accountName, cosmos, msg)
}

func broadcastSweeptxNYKS(sweepTxHex string, refundTxHex string, accountName string) {
	cosmos := getCosmosClient()
	cosmos_address := getCosmosAddress(accountName, cosmos)
	msg := &bridgetypes.MsgBroadcastTxSweep{
		SignedRefundTx: refundTxHex,
		SignedSweepTx:  sweepTxHex,
		JudgeAddress:   cosmos_address.String(),
	}

	broadcastSweeptx(accountName, cosmos, msg)
}

func createTxFromHex(txHex string) (*wire.MsgTx, error) {
	// Decode the transaction hex string
	txBytes, err := hex.DecodeString(txHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex string: %v", err)
	}

	// Create a new transaction object
	tx := wire.NewMsgTx(wire.TxVersion)

	// Deserialize the transaction bytes
	err = tx.Deserialize(bytes.NewReader(txBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize transaction: %v", err)
	}

	return tx, nil
}

func generate_signed_tx(address string, accountName string, sweeptx *wire.MsgTx) ([]byte, error) {

	for {
		receiveSweepSignatures := getSignSweep()

		filteredSweepSignatures := filterSignSweep(receiveSweepSignatures, address)

		if len(filteredSweepSignatures)%1 != 0 {
			fmt.Println("INFO: ", "not enough signatures")
			continue
		}
		dataSig := make([][]byte, 0)

		for _, sig := range filteredSweepSignatures {
			sig, _ := hex.DecodeString(sig.SweepSignature)
			dataSig = append(dataSig, sig)
		}

		script := querySweepAddressScript(address)
		preimage := querySweepAddressPreimage(address)

		for i := 0; i < len(sweeptx.TxIn); i++ {
			witness := make([][]byte, 0)
			witness = append(witness, script)
			witness = append(witness, preimage)
			for j := 0; j < 1; j++ {
				witness = append(witness, dataSig[j])
			}
			sweeptx.TxIn[i].Witness = witness
		}

		var signedTx bytes.Buffer
		err := sweeptx.Serialize(&signedTx)
		if err != nil {
			log.Fatal(err)
		}

		return signedTx.Bytes(), nil
	}
}

func filterSignSweep(sweepSignatures MsgSignSweepResp, address string) []MsgSignSweep {
	signSweep := make([]MsgSignSweep, 0)

	for _, sig := range sweepSignatures.SignSweepMsg {
		if sig.ReserveAddress == address {
			signSweep = append(signSweep, sig)
		}
	}
	return signSweep
}

func encodeSignatures(signatures [][]byte) string {
	hexsignatures := make([]string, 0)
	for _, sig := range signatures {
		hexSig := hex.EncodeToString(sig)
		hexsignatures = append(hexsignatures, hexSig)
	}

	allsignatures := strings.Join(hexsignatures, " ")
	return allsignatures
}

func decodeSignatures(signatures string) [][]byte {
	signaturesbyte := make([][]byte, 0)
	hexsignatures := strings.Split(signatures, " ")
	for _, hexSig := range hexsignatures {
		sig, _ := hex.DecodeString(hexSig)
		signaturesbyte = append(signaturesbyte, sig)
	}

	return signaturesbyte
}

func signTx(tx *wire.MsgTx, address string) []byte {
	amount := queryAmount(tx.TxIn[0].PreviousOutPoint.Index, tx.TxIn[0].PreviousOutPoint.Hash.String())
	sighashes := txscript.NewTxSigHashes(tx)
	script := querySweepAddressScript(address)

	fmt.Println("masterkey : ", masterPrivateKey)

	privkeybytes, err := masterPrivateKey.Serialize()
	if err != nil {
		fmt.Println("Error:", err)
	}

	privkey, _ := btcec.PrivKeyFromBytes(btcec.S256(), privkeybytes)
	fmt.Println("btcec key : ", privkey)

	signature, err := txscript.RawTxInWitnessSignature(tx, sighashes, 0, int64(amount), script, txscript.SigHashSingle|txscript.SigHashAnyOneCanPay, privkey)
	if err != nil {
		fmt.Println("Error:", err)
	}

	return signature
}

func initJudge(accountName string) {
	fmt.Println("init judge")
	height := 0
	number := fmt.Sprintf("%v", viper.Get("no_of_Multisigs"))
	noOfMultisigs, _ := strconv.Atoi(number)
	resp := getAttestations()
	if len(resp.Attestations) <= 0 {
		fmt.Println("no attestations attestaions")
		return
	} else {
		attestation := resp.Attestations[0]
		if attestation.Observed == true {
			btc_height, err := strconv.Atoi(attestation.Proposal.Height)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			height = btc_height
		}
	}

	if height > 0 {

		for i := 0; i < noOfMultisigs; i++ {
			address, script := generateAddress(height + noOfMultisigs)
			registerReserveAddressOnNyks(accountName, address, script)
			registerAddressOnForkscanner(address)
			height = height + 1
		}
	}

}

func startJudge(accountName string) {
	var address SweepAddress
	var transaction string
	for {
		if judge == true {
			resp := getAttestations()
			if len(resp.Attestations) <= 0 {
				fmt.Println("INFO: ", "no attestations")
				time.Sleep(1 * time.Minute)
				continue
			}

			attestation := resp.Attestations[0]
			if attestation.Observed == false {
				fmt.Println("Info: ", "latest attestation not observed")
				time.Sleep(1 * time.Minute)
				continue
			}
			height, _ := strconv.Atoi(attestation.Proposal.Height)
			addresses := querySweepAddresses(uint64(height))
			if len(addresses) <= 0 {
				fmt.Println("INFO: ", "no sweep address found")
				time.Sleep(5 * time.Minute)
				continue
			}

			address = addresses[0]
			tx, withdrawals, total, err := generateSweepTx(address, accountName, height)
			transaction = tx
			if err != nil {
				fmt.Println("Error: ", err)
				continue
			}
			sendSweepProposal(tx, address.Address, withdrawals, accountName, total)
		}

		processSweepTx(accountName)

		if judge == true {
			sweeptx, err := createTxFromHex(transaction)
			if err != nil {
				log.Fatal(err)
			}
			signedTx, err := generate_signed_tx(address.Address, accountName, sweeptx)
			if err != nil {
				log.Fatal(err)
			}

			unsignedTxHex := hex.EncodeToString(signedTx)
			fmt.Println("Signed P2WSH transaction with preimage:", unsignedTxHex)

			broadcastSweeptxNYKS(unsignedTxHex, unsignedTxHex, accountName)

		}

	}
}

func processSweepTx(accountName string) {
	SweepProposal := getAttestationsSweepProposal()

	sweeptxHex := SweepProposal.Attestations[0].Proposal.BtcSweepTx
	reserveAddress := SweepProposal.Attestations[0].Proposal.ReserveAddress
	sweeptx, err := createTxFromHex(sweeptxHex)
	if err != nil {
		log.Fatal(err)
	}

	signature := signTx(sweeptx, reserveAddress)
	hexSignature := hex.EncodeToString(signature)
	sendSweepSign(hexSignature, reserveAddress, accountName)
}