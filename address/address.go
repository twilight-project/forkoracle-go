package address

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
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
	"github.com/btcsuite/btcd/txscript"
	"github.com/cosmos/btcutil"
	"github.com/spf13/viper"
	comms "github.com/twilight-project/forkoracle-go/comms"
	db "github.com/twilight-project/forkoracle-go/db"
	btcOracleTypes "github.com/twilight-project/forkoracle-go/types"
	utils "github.com/twilight-project/forkoracle-go/utils"
	bridgetypes "github.com/twilight-project/nyks/x/bridge/types"

	"golang.org/x/crypto/ripemd160"
)

func hash160(data []byte) []byte {
	hash := sha256.Sum256(data)
	ripemdHash := ripemd160.New()
	_, err := ripemdHash.Write(hash[:])
	if err != nil {
		fmt.Println("Error computing hash:", err)
		return nil
	}
	hash160 := ripemdHash.Sum(nil)
	return hash160
}

func Preimage() ([]byte, error) {
	preimage := make([]byte, 32)
	_, err := rand.Read(preimage)
	if err != nil {
		fmt.Println("Error generating preimage:", err)
		return nil, err
	}
	return preimage, nil
}

func buildScript(preimage []byte, unlockHeight int64, judgeAddr string) ([]byte, error) {
	var fragment btcOracleTypes.Fragment
	fragments := comms.GetAllFragments()
	for _, f := range fragments.Fragments {
		if f.JudgeAddress == judgeAddr {
			fragment = f
		}
	}
	signers := fragment.Signers

	number := fmt.Sprintf("%v", viper.Get("csv_delay"))
	delayPeriod, _ := strconv.Atoi(number)
	payment_hash := hash160(preimage)
	builder := txscript.NewScriptBuilder()

	// adding multisig check

	builder.AddInt64(unlockHeight - 1)
	builder.AddOp(txscript.OP_CHECKLOCKTIMEVERIFY)
	builder.AddOp(txscript.OP_DROP)

	required := int64(len(signers) * 2 / 3)

	if required == 0 {
		required = 1
	}

	builder.AddInt64(required)

	for _, signer := range signers {
		pubKeyBytes, err := hex.DecodeString(signer.SignerBtcPublicKey)
		if err != nil {
			panic(err)
		}

		// Deserialize the public key bytes into a *btcec.PublicKey
		pubKey, err := btcec.ParsePubKey(pubKeyBytes, btcec.S256())
		if err != nil {
			panic(err)
		}

		builder.AddData(pubKey.SerializeCompressed())

	}
	builder.AddInt64(int64(len(signers)))
	builder.AddOp(txscript.OP_CHECKMULTISIGVERIFY)

	// adding preimage check if multisig passes
	builder.AddOp(txscript.OP_SIZE)
	builder.AddInt64(32)
	builder.AddOp(txscript.OP_EQUALVERIFY)
	builder.AddOp(txscript.OP_HASH160)
	builder.AddData(payment_hash)
	builder.AddOp(txscript.OP_EQUAL)

	builder.AddOp(txscript.OP_NOTIF)
	builder.AddInt64(unlockHeight + int64(delayPeriod))
	builder.AddOp(txscript.OP_CHECKSEQUENCEVERIFY)
	builder.AddOp(txscript.OP_DROP)
	builder.AddOp(txscript.OP_ENDIF)

	redeemScript, err := builder.Script()
	if err != nil {
		fmt.Println(err)
	}
	return redeemScript, nil
}

func buildWitnessScript(redeemScript []byte) []byte {
	WitnessScript := sha256.Sum256(redeemScript)
	return WitnessScript[:]
}

func GenerateAddress(unlock_height int64, oldReserveAddress string, judgeAddr string, dbconn *sql.DB) (string, []byte) {
	preimage, err := Preimage()
	if err != nil {
		fmt.Println(err)
	}
	redeemScript, err := buildScript(preimage, unlock_height, judgeAddr)
	if err != nil {
		fmt.Println(err)
	}
	WitnessScript := buildWitnessScript(redeemScript)
	address, err := btcutil.NewAddressWitnessScriptHash(WitnessScript, &chaincfg.MainNetParams)
	if err != nil {
		fmt.Println("Error:", err)
		return "", nil
	}

	addressStr := address.String()
	fmt.Println("new address generated : ", addressStr)

	db.InsertSweepAddress(dbconn, addressStr, redeemScript, preimage, int64(unlock_height), oldReserveAddress, true)

	return addressStr, redeemScript
}

func proposeAddress(accountName string, reserveId uint64, roundId uint64, oldAddress string, judgeAddr string, dbconn *sql.DB) {
	number := fmt.Sprintf("%v", viper.Get("unlocking_time"))
	unlockingTime, _ := strconv.Atoi(number)

	var lastSweepAddress btcOracleTypes.SweepAddress
	addresses := db.QuerySweepAddressesOrderByHeight(dbconn, 1)
	if len(addresses) == 0 {
		fmt.Println("address proposer : no Sweep address found")
		return
	}

	lastSweepAddress = addresses[0]

	unlockHeight := lastSweepAddress.Unlock_height + int64(unlockingTime)
	newReserveAddress, hexScript := GenerateAndRegisterNewProposedAddress(dbconn, accountName, unlockHeight, oldAddress, judgeAddr)

	cosmos_client := comms.GetCosmosClient()
	msg := &bridgetypes.MsgProposeSweepAddress{
		BtcScript:    hexScript,
		BtcAddress:   newReserveAddress,
		JudgeAddress: judgeAddr,
		ReserveId:    reserveId,
		RoundId:      roundId,
	}

	comms.SendTransactionSweepAddressProposal(accountName, cosmos_client, msg)
	db.InsertProposedAddress(dbconn, oldAddress, newReserveAddress, unlockHeight, int64(roundId), int64(reserveId))

	fmt.Println("finishing propose Address after proposing")
}

func ProcessProposeAddress(accountName string, judgeAddr string, dbconn *sql.DB) {
	fmt.Println("Process propose address started")
	number := fmt.Sprintf("%v", viper.Get("sweep_preblock"))
	sweepInitateBlockHeight, _ := strconv.Atoi(number)

	for {

		resp := comms.GetAttestations("20")
		if len(resp.Attestations) <= 0 {
			time.Sleep(1 * time.Minute)
			fmt.Println("no attestaions (start judge)")
			continue
		}

		btcReserves := comms.GetBtcReserves()

		for _, attestation := range resp.Attestations {
			height, _ := strconv.Atoi(attestation.Proposal.Height)
			if !attestation.Observed {
				continue
			}

			addresses := db.QuerySweepAddressesByHeight(dbconn, uint64(height+sweepInitateBlockHeight), true)
			if len(addresses) <= 0 {
				continue
			}

			var reserve btcOracleTypes.BtcReserve
			for _, r := range btcReserves.BtcReserves {
				if r.ReserveAddress == addresses[0].Address {
					reserve = r
				}
			}
			roundId, _ := strconv.Atoi(reserve.RoundId)
			reserveId, _ := strconv.Atoi(reserve.ReserveId)

			proposed := db.CheckIfAddressIsProposed(dbconn, int64(roundId+1), uint64(reserveId))
			if proposed {
				break
			}
			fmt.Println("Sweep Address found, proposing address for reserve : {}, round : {}", reserveId, roundId+1)
			proposeAddress(accountName, uint64(reserveId), uint64(roundId+1), reserve.ReserveAddress, judgeAddr, dbconn)
			break
		}

	}
}

func GenerateAndRegisterNewProposedAddress(dbconn *sql.DB, accountName string, height int64, oldReserveAddress string, oracleAddr string) (string, string) {
	newSweepAddress, script := GenerateAddress(height, oldReserveAddress, oracleAddr, dbconn)
	registerAddressOnForkscanner(newSweepAddress)
	hexScript := hex.EncodeToString(script)
	return newSweepAddress, hexScript
}

func GenerateAndRegisterNewBtcReserveAddress(dbconn *sql.DB, accountName string, height int64, judgeAddr string, fragmentId int) string {
	newSweepAddress, reserveScript := GenerateAddress(height, "", judgeAddr, dbconn)
	registerReserveAddressOnNyks(accountName, newSweepAddress, reserveScript, judgeAddr, fragmentId)
	registerAddressOnForkscanner(newSweepAddress)

	return newSweepAddress
}

func registerReserveAddressOnNyks(accountName string, address string, script []byte, judgeAddr string, fragmentId int) {

	cosmos := comms.GetCosmosClient()
	reserveScript := hex.EncodeToString(script)

	msg := &bridgetypes.MsgRegisterReserveAddress{
		FragmentId:     uint64(fragmentId),
		ReserveScript:  reserveScript,
		ReserveAddress: address,
		JudgeAddress:   judgeAddr,
	}

	// store response in txResp
	txResp, err := comms.SendTransactionRegisterReserveAddress(accountName, cosmos, msg)
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

	fmt.Println("registered address on forkscanner : ", address)

}

func UnRegisterAddressOnForkscanner(address string) {
	dt := time.Now().UTC()
	dt = dt.AddDate(1, 0, 0)

	request_body := map[string]interface{}{
		"method":  "remove_watched_addresses",
		"id":      1,
		"jsonrpc": "2.0",
		"params": map[string]interface{}{
			"remove": []string{address},
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

	fmt.Println("registered address on forkscanner : ", address)

}

func RegisterAddressOnValidators(dbconn *sql.DB) {
	// {add check to see if the address already exists}
	fmt.Println("registering address on validators")
	savedAddress := db.QueryAllAddressOnly(dbconn)
	respReserve := comms.GetReserveAddresses()
	if len(respReserve.Addresses) > 0 {
		for _, address := range respReserve.Addresses {
			if !utils.StringInSlice(address.ReserveAddress, savedAddress) {
				registerAddressOnForkscanner(address.ReserveAddress)
				decodedScript := utils.DecodeBtcScript(address.ReserveScript)
				height := utils.GetHeightFromScript(decodedScript)
				reserveScript, _ := hex.DecodeString(address.ReserveScript)
				db.InsertSweepAddress(dbconn, address.ReserveAddress, reserveScript, nil, height+1, "", false)
			}
		}
	}
	respProposed := comms.GetProposedAddresses()
	if len(respProposed.ProposeSweepAddressMsgs) > 0 {
		for _, address := range respProposed.ProposeSweepAddressMsgs {
			if !utils.StringInSlice(address.BtcAddress, savedAddress) {
				registerAddressOnForkscanner(address.BtcAddress)
				decodedScript := utils.DecodeBtcScript(address.BtcScript)
				height := utils.GetHeightFromScript(decodedScript)
				reserveScript, _ := hex.DecodeString(address.BtcScript)
				db.InsertSweepAddress(dbconn, address.BtcAddress, reserveScript, nil, height+1, "", false)
			}
		}
	}
}

func RegisterAddressOnSigners(dbconn *sql.DB) {
	// {add check to see if the address already exists}
	fmt.Println("registering address on validators")
	savedAddress := db.QueryAllAddressOnly(dbconn)
	respReserve := comms.GetReserveAddresses()
	if len(respReserve.Addresses) > 0 {
		for _, address := range respReserve.Addresses {
			if !utils.StringInSlice(address.ReserveAddress, savedAddress) {
				decodedScript := utils.DecodeBtcScript(address.ReserveScript)
				height := utils.GetHeightFromScript(decodedScript)
				reserveScript, _ := hex.DecodeString(address.ReserveScript)
				db.InsertSweepAddress(dbconn, address.ReserveAddress, reserveScript, nil, height+1, "", false)
			}
		}
	}
	respProposed := comms.GetProposedAddresses()
	if len(respProposed.ProposeSweepAddressMsgs) > 0 {
		for _, address := range respProposed.ProposeSweepAddressMsgs {
			if !utils.StringInSlice(address.BtcAddress, savedAddress) {
				decodedScript := utils.DecodeBtcScript(address.BtcScript)
				height := utils.GetHeightFromScript(decodedScript)
				reserveScript, _ := hex.DecodeString(address.BtcScript)
				db.InsertSweepAddress(dbconn, address.BtcAddress, reserveScript, nil, height+1, "", false)
			}
		}
	}
}
