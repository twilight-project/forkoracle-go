package utils

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/spf13/viper"
	comms "github.com/twilight-project/forkoracle-go/comms"
	btcOracleTypes "github.com/twilight-project/forkoracle-go/types"
)

// func GetBtcPublicKey() string {
// 	client := getBitcoinRpcClient()
// 	walletName := fmt.Sprintf("%v", viper.Get("wallet_name"))
// 	var address string

// 	rpc := client.ListReceivedByAddressIncludeEmptyAsync(0, true)
// 	addresses, err := rpc.Receive()
// 	if err != nil || len(addresses) == 0 {
// 		fmt.Println("error getting accounts creating a new address")
// 		addr, err := client.GetNewAddress(walletName)
// 		if err != nil {
// 			fmt.Println("error in getting btc pub key : ", err)
// 		}
// 		address = addr.String()
// 	} else {
// 		address = addresses[0].Address
// 	}

// 	info, err := client.GetAddressInfo(address)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	return *info.PubKey
// }

func InitConfigFile() {
	viper.AddConfigPath("./configs")
	viper.SetConfigName("config") // Register config file name (no extension)
	viper.SetConfigType("json")   // Look for specific type
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("Error reading config file: ", err)
	}
}

func SetDelegator(valAddr string, oracleAddr string, btcPublicKey string) {
	accountName := fmt.Sprintf("%v", viper.Get("accountName"))
	command := fmt.Sprintf("nyksd tx nyks set-delegate-addresses %s %s %s %s --from %s --chain-id nyks --keyring-backend test -y", valAddr, oracleAddr, btcPublicKey, oracleAddr, accountName)
	fmt.Println("delegate command : ", command)

	args := strings.Fields(command)
	cmd := exec.Command(args[0], args[1:]...)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		panic(err)
	}

	fmt.Println("Delegate Address Set")
}

func getBitcoinRpcClient() *rpcclient.Client {
	connCfg := &rpcclient.ConnConfig{
		Host:         viper.GetString("btc_node_host"),
		User:         viper.GetString("btc_node_user"),
		Pass:         viper.GetString("btc_node_pass"),
		HTTPPostMode: true,
		DisableTLS:   true,
	}

	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		fmt.Println("Failed to connect to the Bitcoin client : ", err)
	}

	return client
}

// func BroadcastBtcTransaction(tx *wire.MsgTx) {
// 	client := getBitcoinRpcClient()
// 	txHash, err := client.SendRawTransaction(tx, true)
// 	if err != nil {
// 		fmt.Println("Failed to broadcast transaction : ", err)
// 	}

// 	defer client.Shutdown()
// 	fmt.Println("broadcasted btc transaction, txhash : ", txHash)
// }

// func getReserveForAddress(address string) BtcReserve {
// 	btcReserves := getBtcReserves()
// 	for _, reserve := range btcReserves.BtcReserves {
// 		if reserve.ReserveAddress == address {
// 			return reserve
// 		}
// 	}
// 	return BtcReserve{}
// }

func CreateTxFromHex(txHex string) (*wire.MsgTx, error) {
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

func SignTx(tx *wire.MsgTx, script []byte) []string {
	fmt.Println("inside sign tx")
	signatures := []string{}
	witnessInputs := make([]btcjson.RawTxWitnessInput, len(tx.TxIn))
	client := getBitcoinRpcClient()
	fmt.Println("inside sign tx 1")
	for i, input := range tx.TxIn {
		tx, err := client.GetRawTransactionVerbose(&input.PreviousOutPoint.Hash)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("inside sign tx 2 loop")
		witnessInputs[i] = btcjson.RawTxWitnessInput{
			Txid:         input.PreviousOutPoint.Hash.String(),
			Vout:         input.PreviousOutPoint.Index,
			ScriptPubKey: tx.Vout[input.PreviousOutPoint.Index].ScriptPubKey.Hex,
			Amount:       &tx.Vout[input.PreviousOutPoint.Index].Value,
		}
	}
	fmt.Println("inside sign tx 3")
	signedTx, _, err := client.SignRawTransactionWithWallet3(tx, witnessInputs, rpcclient.SigHashAllAnyoneCanPay)
	if err != nil {
		fmt.Println("Error in signing btc tx:", err)
	}

	for _, input := range signedTx.TxIn {
		signatures = append(signatures, hex.EncodeToString(input.Witness[0]))
	}
	return signatures
}

// func signTx(tx *wire.MsgTx, address string) []byte {
// 	amount := queryAmount(tx.TxIn[0].PreviousOutPoint.Index, tx.TxIn[0].PreviousOutPoint.Hash.String())
// 	sighashes := txscript.NewTxSigHashes(tx)
// 	script := querySweepAddressScriptByParentAddress(address)

// 	privkeybytes, err := masterPrivateKey.Serialize()
// 	if err != nil {
// 		fmt.Println("Error: converting private key to bytes : ", err)
// 	}

// 	privkey, _ := btcec.PrivKeyFromBytes(btcec.S256(), privkeybytes)

// 	signature, err := txscript.RawTxInWitnessSignature(tx, sighashes, 0, int64(amount), script, txscript.SigHashAll|txscript.SigHashAnyOneCanPay, privkey)
// 	if err != nil {
// 		fmt.Println("Error:", err)
// 	}

// 	return signature
// }

// func reverseArray(arr []MsgSignSweep) []MsgSignSweep {
// 	for i, j := 0, len(arr)-1; i < j; i, j = i+1, j-1 {
// 		arr[i], arr[j] = arr[j], arr[i]
// 	}
// 	return arr
// }

func StringInSlice(str string, slice []string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
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

func CreateTxIn(utxo btcOracleTypes.Utxo) (*wire.TxIn, error) {
	utxoHash, err := chainhash.NewHashFromStr(utxo.Txid)
	if err != nil {
		log.Println("error with UTXO")
		return nil, err
	}
	outPoint := wire.NewOutPoint(utxoHash, utxo.Vout)
	txIn := wire.NewTxIn(outPoint, nil, nil)
	return txIn, nil
}

// func SendSweepSign(hexSignatures []string, address string, accountName string, reserveId uint64, roundId uint64, oracleAddr string) {
// 	cosmos := comms.GetCosmosClient()
// 	msg := &bridgetypes.MsgSignSweep{
// 		ReserveId:        reserveId,
// 		RoundId:          roundId,
// 		SignerPublicKey:  getBtcPublicKey(),
// 		SweepSignature:   hexSignatures,
// 		BtcOracleAddress: oracleAddr,
// 	}

// 	comms.SendTransactionSignSweep(accountName, cosmos, msg)
// }

// func sendRefundSign(hexSignatures string, address string, accountName string, reserveId uint64, roundId uint64, oracleAddr string) {
// 	cosmos := comms.GetCosmosClient()
// 	msg := &bridgetypes.MsgSignRefund{
// 		ReserveId:        reserveId,
// 		RoundId:          roundId,
// 		SignerPublicKey:  getBtcPublicKey(),
// 		RefundSignature:  []string{hexSignatures},
// 		BtcOracleAddress: oracleAddr,
// 	}

// 	comms.SendTransactionSignRefund(accountName, cosmos, msg)
// }

// func RegisterJudge(accountName string, oracleAddr string, valAddr string) {
// 	cosmos := comms.GetCosmosClient()
// 	msg := &bridgetypes.MsgRegisterJudge{
// 		Creator:          oracleAddr,
// 		JudgeAddress:     oracleAddr,
// 		ValidatorAddress: valAddr,
// 	}

// 	comms.SendTransactionRegisterJudge(accountName, cosmos, msg)
// 	fmt.Println("registered Judge")
// }

func FilterAndOrderSignSweep(sweepSignatures btcOracleTypes.MsgSignSweepResp, pubkeys []string, judgeAddr string) []btcOracleTypes.MsgSignSweep {
	fmt.Println(sweepSignatures.SignSweepMsg)
	fmt.Println(pubkeys)
	filtereSignSweep := []btcOracleTypes.MsgSignSweep{}
	for _, sweepSig := range sweepSignatures.SignSweepMsg {
		if StringInSlice(sweepSig.SignerPublicKey, pubkeys) {
			filtereSignSweep = append(filtereSignSweep, sweepSig)
		}
	}

	fragments := comms.GetAllFragments()
	var fragment btcOracleTypes.Fragment
	found := false
	for _, f := range fragments.Fragments {
		if f.JudgeAddress == judgeAddr {
			fragment = f
			found = true
			break
		}
	}
	if !found {
		fmt.Println("No fragment found with the specified judge address")
		return nil
	}

	orderedSignSweep := make([]btcOracleTypes.MsgSignSweep, 0)

	for _, signer := range fragment.Signers {
		for _, sweepSig := range filtereSignSweep {
			if signer.SignerAddress == sweepSig.BtcOracleAddress {
				orderedSignSweep = append(orderedSignSweep, sweepSig)
			}
		}
	}

	fmt.Println("Signatures Sweep : ", len(orderedSignSweep))

	return orderedSignSweep
}

func OrderSignRefund(refundSignatures btcOracleTypes.MsgSignRefundResp, address string,
	pubkeys []string, oracleAddr string) ([]btcOracleTypes.MsgSignRefund, btcOracleTypes.MsgSignRefund) {

	delegateAddresses := comms.GetDelegateAddresses()

	filteresSignRefund := make([]btcOracleTypes.MsgSignRefund, 0)
	for _, refundSig := range refundSignatures.SignRefundMsg {
		if StringInSlice(refundSig.SignerPublicKey, pubkeys) {
			filteresSignRefund = append(filteresSignRefund, refundSig)
		}
	}

	orderedSignRefund := make([]btcOracleTypes.MsgSignRefund, 0)
	var judgeSign btcOracleTypes.MsgSignRefund

	for _, oracleAddr := range delegateAddresses.Addresses {
		for _, refundSig := range refundSignatures.SignRefundMsg {
			if oracleAddr.BtcOracleAddress == refundSig.BtcOracleAddress {
				orderedSignRefund = append(orderedSignRefund, refundSig)
			}
		}
	}
	fmt.Println("Signatures refund : ", len(orderedSignRefund))

	return orderedSignRefund, judgeSign
}

// func GetCurrentReserveandRound(oracleAddr string) (btcOracleTypes.BtcReserve, uint64, uint64, error) {

// 	empty := btcOracleTypes.BtcReserve{"", "", "", "", "", "", "", "", "", ""}
// 	var currentReservesForThisJudge []btcOracleTypes.BtcReserve
// 	reserves := comms.GetBtcReserves()
// 	for _, reserve := range reserves.BtcReserves {
// 		if reserve.JudgeAddress == oracleAddr {
// 			currentReservesForThisJudge = append(currentReservesForThisJudge, reserve)
// 		}
// 	}

// 	if len(currentReservesForThisJudge) == 0 {
// 		time.Sleep(2 * time.Minute)
// 		fmt.Println("no judge")
// 		return empty, 0, 0, errors.New("No Judge Found")
// 	}

// 	currentJudgeReserve := currentReservesForThisJudge[0]

// 	reserveIdForProposal, _ := strconv.Atoi(currentJudgeReserve.ReserveId)
// 	if reserveIdForProposal == 1 {
// 		reserveIdForProposal = len(reserves.BtcReserves)
// 	} else {
// 		reserveIdForProposal = reserveIdForProposal - 1
// 	}

// 	var reserveToBeUpdated btcOracleTypes.BtcReserve
// 	for _, reserve := range reserves.BtcReserves {
// 		tempId, _ := strconv.Atoi(reserve.ReserveId)
// 		if tempId == reserveIdForProposal {
// 			reserveToBeUpdated = reserve
// 			break
// 		}
// 	}

// 	RoundId, _ := strconv.Atoi(reserveToBeUpdated.RoundId)
// 	return currentJudgeReserve, uint64(reserveIdForProposal), uint64(RoundId), nil
// }

func btcToSats(btc float64) int64 {
	return int64(btc * 1e8)
}

// func GetFeeFromBtcNode(tx *wire.MsgTx) (int64, error) {
// 	client := getBitcoinRpcClient()
// 	result, err := client.EstimateSmartFee(2, &btcjson.EstimateModeConservative)
// 	if err != nil {
// 		fmt.Println("Failed to get fee from btc node : ", err)
// 		log.Fatal(err)
// 	}
// 	fmt.Printf("Estimated fee per kilobyte for a transaction to be confirmed within 2 blocks: %f BTC\n", *result.FeeRate)
// 	feeRate := btcToSats(*result.FeeRate)
// 	fmt.Printf("Estimated fee per kilobyte for a transaction to be confirmed within 2 blocks: %d Sats\n", feeRate)
// 	baseSize := tx.SerializeSizeStripped()
// 	totalSize := tx.SerializeSize()
// 	weight := (baseSize * 3) + totalSize
// 	vsize := (weight + 3) / 4
// 	fmt.Println("tx size in bytes : ", vsize)
// 	fee := float64(vsize) * float64(feeRate/1024)
// 	return int64(fee), nil
// }

func GetBtcFeeRate() btcOracleTypes.FeeRate {
	resp, err := http.Get("https://api.blockchain.info/mempool/fees")
	if err != nil {
		log.Fatalln(err)
	}
	//We Read the response body on the line below.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	a := btcOracleTypes.FeeRate{}
	err = json.Unmarshal(body, &a)
	if err != nil {
		fmt.Println("Error decoding Fee Rate : ", err)
	}

	return a
}

func DecodeBtcScript(script string) string {
	decoded, err := hex.DecodeString(script)
	if err != nil {
		fmt.Println("Error decoding script Hex : ", err)
	}
	decodedScript, err := txscript.DisasmString(decoded)
	if err != nil {
		fmt.Println("Error decoding script : ", err)
	}

	return decodedScript
}

func GetHeightFromScript(script string) int64 {
	// Split the decoded script into parts
	height := int64(0)
	parts := strings.Split(script, " ")
	if len(parts) == 0 {
		return height
	}
	// Reverse the byte order
	for i, j := 0, len(parts[0])-2; i < j; i, j = i+2, j-2 {
		parts[0] = parts[0][:i] + parts[0][j:j+2] + parts[0][i+2:j] + parts[0][i:i+2] + parts[0][j+2:]
	}
	// Convert the first part from hex to decimal
	height, err := strconv.ParseInt(parts[0], 16, 64)
	if err != nil {
		fmt.Println("Error converting block height from hex to decimal:", err)
	}

	return height
}

func GetMinSignFromScript(script string) int64 {
	// Split the decoded script into parts
	minSignRequired := int64(0)
	parts := strings.Split(script, " ")
	if len(parts) < 4 {
		return minSignRequired
	}
	// Reverse the byte order
	for i, j := 0, len(parts[3])-2; i < j; i, j = i+2, j-2 {
		parts[3] = parts[3][:i] + parts[3][j:j+2] + parts[3][i+2:j] + parts[3][i:i+2] + parts[3][j+2:]
	}
	// Convert the first part from hex to decimal
	minSignRequired, err := strconv.ParseInt(parts[3], 16, 64)
	if err != nil {
		fmt.Println("Error converting block height from hex to decimal:", err)
	}

	return minSignRequired
}

func GetPublicKeysFromScript(script string, limit int) []string {
	// Split the decoded script into parts
	pubkeys := []string{}
	parts := strings.Split(script, " ")
	if len(parts) <= 4+limit {
		return pubkeys
	}
	// Reverse the byte order
	for i, j := 0, len(parts[3])-2; i < j; i, j = i+2, j-2 {
		parts[3] = parts[3][:i] + parts[3][j:j+2] + parts[3][i+2:j] + parts[3][i:i+2] + parts[3][j+2:]
	}
	// Convert the first part from hex to decimal
	pubkeys = append(pubkeys, parts[4:4+limit]...)

	return pubkeys
}
