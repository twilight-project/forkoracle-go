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

	reserveScript := hex.EncodeToString(script)

	msg := &bridgetypes.MsgRegisterReserveAddress{
		ReserveScript:  reserveScript,
		ReserveAddress: address,
		JudgeAddress:   oracleAddr,
	}

	// store response in txResp
	txResp, err := cosmos.BroadcastTx(accountName, msg)
	if err != nil {
		fmt.Println("error in registering reserve address : ", err)
	}

	// print response from broadcasting a transaction
	fmt.Println("MsgRegisterReserveAddress : ")
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

func removeAddressOnForkscanner(address string) {
	dt := time.Now().UTC()
	dt = dt.AddDate(1, 0, 0)

	request_body := map[string]interface{}{
		"method":  "remove_watched_addresses",
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

func generateAndRegisterNewAddress(accountName string, height int) string {
	newSweepAddress, reserveScript := generateAddress(height)
	registerReserveAddressOnNyks(accountName, newSweepAddress, reserveScript)
	registerAddressOnForkscanner(newSweepAddress)
	return newSweepAddress
}

func generateSweepTx(sweepAddress SweepAddress, accountName string, height int) (string, []BtcWithdrawRequest, uint64, error) {
	number := fmt.Sprintf("%v", viper.Get("no_of_Multisigs"))
	noOfMultisigs, _ := strconv.Atoi(number)

	number = fmt.Sprintf("%v", viper.Get("unlocking_time"))
	unlockingTimeInBlocks, _ := strconv.Atoi(number)

	utxos := queryUtxo(sweepAddress.Address)
	if len(utxos) <= 0 {
		// addr := generateAndRegisterNewAddress(accountName, height+noOfMultisigs)
		fmt.Println("INFO : No funds in address : ", sweepAddress.Address, " generating new address : ")
		return "", nil, 0, nil
	}
	withdrawals := getBtcWithdrawRequest()
	totalAmountTxIn := uint64(0)
	totalAmountTxOut := uint64(0)

	redeemTx := wire.NewMsgTx(wire.TxVersion)
	for _, utxo := range utxos {
		txIn, err := CreateTxIn(utxo)
		if err != nil {
			fmt.Println("error while add tx in : ", err)
			return "", nil, 0, err
		}
		totalAmountTxIn = totalAmountTxIn + utxo.Amount
		txIn.Sequence = wire.MaxTxInSequenceNum - 10
		redeemTx.AddTxIn(txIn)
	}

	withdrawRequests := make([]BtcWithdrawRequest, 0)
	for _, withdrawal := range withdrawals.WithdrawRequest {
		if withdrawal.ReserveAddress == sweepAddress.Address {
			amount, err := strconv.Atoi(withdrawal.WithdrawAmount)
			txOut, err := CreateTxOut(withdrawal.WithdrawAddress, int64(amount))
			if err != nil {
				fmt.Println("error while txout : ", err)
				return "", nil, 0, err
			}
			totalAmountTxOut = totalAmountTxOut + uint64(amount)
			redeemTx.AddTxOut(txOut)
			withdrawRequests = append(withdrawRequests, withdrawal)
		}
	}

	fee := 5000

	newSweepAddress := generateAndRegisterNewAddress(accountName, height+(noOfMultisigs*unlockingTimeInBlocks))
	// newSweepAddress := "bc1qeplu0p23jyu3vkp7wrn0dka00qsg7uacxkslp39m6tcqfg759vasr03hzp"
	// updateAddressUnlockHeight(sweepAddress.Address, height+(noOfMultisigs*unlockingTimeInBlocks))

	txOut, err := CreateTxOut(newSweepAddress, int64(totalAmountTxIn-totalAmountTxOut-uint64(fee)))
	if err != nil {
		log.Println("error with txout", err)
		return "", nil, 0, err
	}
	redeemTx.AddTxOut(txOut)

	var signedTx bytes.Buffer
	redeemTx.Serialize(&signedTx)
	hexTx := hex.EncodeToString(signedTx.Bytes())
	fmt.Println("transaction UnSigned: ", hexTx)

	return hexTx, withdrawRequests, totalAmountTxIn, nil
}

func createAndSendSweepProposal(tx string, address string, withdrawals []BtcWithdrawRequest, accountName string, total uint64) {

	fmt.Println("inside sending sweep proposal")

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

	msg := &bridgetypes.MsgSweepProposal{

		ReserveId:                        1,
		ReserveAddress:                   address,
		JudgeAddress:                     oracleAddr,
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
	msg := &bridgetypes.MsgSignSweep{
		ReserveAddress:   address,
		SignerAddress:    address, // no idea what this is
		SweepSignature:   hexSignatures,
		BtcOracleAddress: oracleAddr,
	}

	sendTransactionSignSweep(accountName, cosmos, msg)
}

func broadcastSweeptxNYKS(sweepTxHex string, refundTxHex string, accountName string) {
	cosmos := getCosmosClient()
	msg := &bridgetypes.MsgBroadcastTxSweep{
		SignedRefundTx: refundTxHex,
		SignedSweepTx:  sweepTxHex,
		JudgeAddress:   oracleAddr,
	}

	sendTransactionBroadcastSweeptx(accountName, cosmos, msg)
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

	number := fmt.Sprintf("%v", viper.Get("no_of_validators"))
	noOfValidators, _ := strconv.Atoi(number)
	for {
		time.Sleep(30 * time.Second)
		receiveSweepSignatures := getSignSweep()
		filteredSweepSignatures := filterSignSweep(receiveSweepSignatures, address)

		if len(filteredSweepSignatures) <= 0 {
			continue
		}

		minSignsRequired := noOfValidators * 2 / 3

		// fmt.Println("INFO: ", "noOfValidators", noOfValidators)
		// fmt.Println("INFO: ", "minSignsRequired", minSignsRequired)
		// fmt.Println("INFO: ", "len(filteredSweepSignatures)", len(filteredSweepSignatures))

		if len(filteredSweepSignatures)/minSignsRequired < 1 {
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
			witness := wire.TxWitness{}
			// dummy := []byte{0}
			// witness = append(witness, dummy)
			for j := 0; j < len(dataSig); j++ {
				witness = append(witness, dataSig[j])
			}
			witness = append(witness, preimage)
			witness = append(witness, script)
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

func reverseArray(arr []MsgSignSweep) []MsgSignSweep {
	for i, j := 0, len(arr)-1; i < j; i, j = i+1, j-1 {
		arr[i], arr[j] = arr[j], arr[i]
	}
	return arr
}

func filterSignSweep(sweepSignatures MsgSignSweepResp, address string) []MsgSignSweep {
	signSweep := make([]MsgSignSweep, 0)

	for _, sig := range sweepSignatures.SignSweepMsg {
		if sig.ReserveAddress == address {
			signSweep = append(signSweep, sig)
		}
	}

	delegateAddresses := getDelegateAddresses()
	orderedSignSweep := make([]MsgSignSweep, 0)

	for _, oracleAddr := range delegateAddresses.Addresses {
		for _, sweepSig := range signSweep {
			if oracleAddr.BtcOracleAddress == sweepSig.BtcOracleAddress {
				orderedSignSweep = append(orderedSignSweep, sweepSig)
			}
		}
	}

	fmt.Println("ordered Signatures Sweep : ", reverseArray(orderedSignSweep))

	return reverseArray(orderedSignSweep)
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

	privkeybytes, err := masterPrivateKey.Serialize()
	if err != nil {
		fmt.Println("Error: converting private key to bytes : ", err)
	}

	privkey, _ := btcec.PrivKeyFromBytes(btcec.S256(), privkeybytes)

	signature, err := txscript.RawTxInWitnessSignature(tx, sighashes, 0, int64(amount), script, txscript.SigHashAll|txscript.SigHashAnyOneCanPay, privkey)
	if err != nil {
		fmt.Println("Error:", err)
	}

	return signature
}

func registerJudge(accountName string) {
	cosmos := getCosmosClient()
	msg := &bridgetypes.MsgRegisterJudge{
		Creator:          oracleAddr,
		JudgeAddress:     oracleAddr,
		ValidatorAddress: valAddr,
	}

	sendTransactionRegisterJudge(accountName, cosmos, msg)
	fmt.Println("registered Judge")
}

func initJudge(accountName string) {
	fmt.Println("init judge")
	height := 0
	number := fmt.Sprintf("%v", viper.Get("no_of_Multisigs"))
	noOfMultisigs, _ := strconv.Atoi(number)

	number = fmt.Sprintf("%v", viper.Get("unlocking_time"))
	unlockingTimeInBlocks, _ := strconv.Atoi(number)

	registerJudge(accountName)
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

		for i := 1; i <= noOfMultisigs; i++ {
			_ = generateAndRegisterNewAddress(accountName, height+(noOfMultisigs*unlockingTimeInBlocks))
			height = height + 1
		}
	}

	fmt.Println("judge initialized")
}

func startJudge(accountName string) {
	fmt.Println("starting judge")
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
				address := addresses[0]

				tx, withdrawals, total, err := generateSweepTx(address, accountName, height)
				if err != nil {
					fmt.Println("Error in generating a Sweep transaction: ", err)
					continue
				}
				if tx == "" {
					fmt.Println("INFO: ", "no sweep tx generated because no funds in current address")
					time.Sleep(1 * time.Minute)
					continue
				}

				createAndSendSweepProposal(tx, address.Address, withdrawals, accountName, total)

				time.Sleep(1 * time.Minute)

				sweeptx, err := createTxFromHex(tx)
				if err != nil {
					fmt.Println("error decoding sweep tx : inside judge")
					fmt.Println(err)
				}
				signedTx, err := generate_signed_tx(address.Address, accountName, sweeptx)
				if err != nil {
					fmt.Println(err)
				}

				signedTxHex := hex.EncodeToString(signedTx)
				fmt.Println("Signed P2WSH transaction with preimage:", signedTxHex)

				broadcastSweeptxNYKS(signedTxHex, signedTxHex, accountName)
				markProcessedSweepAddress(address.Address)
			}

		}

	}
}
