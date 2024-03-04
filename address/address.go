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

func preimage() ([]byte, error) {
	preimage := make([]byte, 32)
	_, err := rand.Read(preimage)
	if err != nil {
		fmt.Println("Error generating preimage:", err)
		return nil, err
	}
	return preimage, nil
}

func buildScript(preimage []byte, unlockHeight int64, oracleAddr string) ([]byte, error) {
	var judgeBtcPK *btcec.PublicKey
	var refundJudgeAddress string
	judges := comms.GetRegisteredJudges()
	if len(judges.Judges) == 0 {
		fmt.Println("no judge found")
		return nil, nil
	} else if len(judges.Judges) == 1 {
		refundJudgeAddress = judges.Judges[0].JudgeAddress
	} else {
		for _, judge := range judges.Judges {
			if judge.JudgeAddress != oracleAddr {
				refundJudgeAddress = judge.JudgeAddress
			}
		}

	}

	number := fmt.Sprintf("%v", viper.Get("csv_delay"))
	delayPeriod, _ := strconv.Atoi(number)
	delegateAddresses := comms.GetDelegateAddresses()
	payment_hash := hash160(preimage)
	builder := txscript.NewScriptBuilder()

	// adding multisig check

	builder.AddInt64(unlockHeight - 1)
	builder.AddOp(txscript.OP_CHECKLOCKTIMEVERIFY)
	builder.AddOp(txscript.OP_DROP)

	required := int64(len(delegateAddresses.Addresses) * 2 / 3)

	if required == 0 {
		required = 1
	}

	builder.AddInt64(required)

	for _, element := range delegateAddresses.Addresses {
		pubKeyBytes, err := hex.DecodeString(element.BtcPublicKey)
		if err != nil {
			panic(err)
		}

		// Deserialize the public key bytes into a *btcec.PublicKey
		pubKey, err := btcec.ParsePubKey(pubKeyBytes, btcec.S256())
		if err != nil {
			panic(err)
		}

		builder.AddData(pubKey.SerializeCompressed())

		//TODO: might need to change this for multi judge setup
		if element.BtcOracleAddress == refundJudgeAddress {
			judgeBtcPK = pubKey
		}
	}
	builder.AddInt64(int64(len(delegateAddresses.Addresses)))
	builder.AddOp(txscript.OP_CHECKMULTISIGVERIFY)

	// adding preimage check if multisig passes
	builder.AddOp(txscript.OP_SIZE)
	builder.AddInt64(32)
	builder.AddOp(txscript.OP_EQUALVERIFY)
	builder.AddOp(txscript.OP_HASH160)
	builder.AddData(payment_hash)
	builder.AddOp(txscript.OP_EQUAL)

	builder.AddOp(txscript.OP_IFDUP)

	// adding judge refund check
	builder.AddOp(txscript.OP_NOTIF)
	builder.AddData(judgeBtcPK.SerializeCompressed())
	builder.AddOp(txscript.OP_CHECKSIG)
	builder.AddOp(txscript.OP_NOTIF)
	builder.AddInt64(unlockHeight + int64(delayPeriod))
	builder.AddOp(txscript.OP_CHECKSEQUENCEVERIFY)
	builder.AddOp(txscript.OP_DROP)
	builder.AddOp(txscript.OP_ENDIF)
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

func GenerateAddress(unlock_height int64, oldReserveAddress string, oracleAddr string, dbconn *sql.DB) (string, []byte) {
	preimage, err := preimage()
	if err != nil {
		fmt.Println(err)
	}
	redeemScript, err := buildScript(preimage, unlock_height, oracleAddr)
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

func proposeAddress(accountName string, reserveId uint64, roundId uint64, oldAddress string, oracleAddr string, dbconn *sql.DB) {
	number := fmt.Sprintf("%v", viper.Get("unlocking_time"))
	unlockingTimeInBlocks, _ := strconv.Atoi(number)

	// temporary till staking is implemented
	number = fmt.Sprintf("%v", viper.Get("height_diff_between_judges"))
	heightDiffBetweenJudges, _ := strconv.Atoi(number)

	var lastSweepAddress btcOracleTypes.SweepAddress
	addresses := db.QuerySweepAddressesOrderByHeight(dbconn, 1)
	if len(addresses) == 0 {
		fmt.Println("address proposer : no Sweep address found")
		return
	}

	lastSweepAddress = addresses[0]

	unlockHeight := lastSweepAddress.Unlock_height + int64(unlockingTimeInBlocks) + int64(heightDiffBetweenJudges)
	newReserveAddress, hexScript := GenerateAndRegisterNewProposedAddress(dbconn, accountName, unlockHeight, oldAddress, oracleAddr)

	cosmos_client := comms.GetCosmosClient()
	msg := &bridgetypes.MsgProposeSweepAddress{
		BtcScript:    hexScript,
		BtcAddress:   newReserveAddress,
		JudgeAddress: oracleAddr,
		ReserveId:    reserveId,
		RoundId:      roundId,
	}

	comms.SendTransactionSweepAddressProposal(accountName, cosmos_client, msg)
	db.InsertProposedAddress(dbconn, oldAddress, newReserveAddress, unlockHeight, int64(roundId), int64(reserveId))

	fmt.Println("finishing propose Address after proposing")
}

func processProposeAddress(accountName string, oracleAddr string, dbconn *sql.DB) {
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

		var currentReservesForThisJudge []btcOracleTypes.BtcReserve
		reserves := comms.GetBtcReserves()
		for _, reserve := range reserves.BtcReserves {
			if reserve.JudgeAddress == oracleAddr {
				currentReservesForThisJudge = append(currentReservesForThisJudge, reserve)
			}
		}

		if len(currentReservesForThisJudge) == 0 {
			time.Sleep(2 * time.Minute)
			fmt.Println("no judge")
			continue
		}

		currentJudgeReserve := currentReservesForThisJudge[0]

		reserveIdForProposal, _ := strconv.Atoi(currentJudgeReserve.ReserveId)
		if reserveIdForProposal == 1 {
			reserveIdForProposal = len(reserves.BtcReserves)
		} else {
			reserveIdForProposal = reserveIdForProposal - 1
		}

		var reserveToBeUpdated btcOracleTypes.BtcReserve
		for _, reserve := range reserves.BtcReserves {
			tempId, _ := strconv.Atoi(reserve.ReserveId)
			if tempId == reserveIdForProposal {
				reserveToBeUpdated = reserve
				break
			}
		}

		RoundId, _ := strconv.Atoi(reserveToBeUpdated.RoundId)
		proposed := db.CheckIfAddressIsProposed(dbconn, int64(RoundId+1))
		if proposed {
			continue
		}

		for _, attestation := range resp.Attestations {
			height, _ := strconv.Atoi(attestation.Proposal.Height)

			if !attestation.Observed {
				continue
			}

			judges := comms.GetRegisteredJudges()

			addresses := db.QuerySweepAddressesByHeight(dbconn, uint64(height+sweepInitateBlockHeight), false)
			if len(addresses) <= 0 {
				if len(judges.Judges) == 1 {
					addresses = db.QuerySweepAddressesByHeight(dbconn, uint64(height+sweepInitateBlockHeight), true)
				}
				if len(addresses) <= 0 {
					continue
				}
			}

			fmt.Println("Sweep Address found, proposing address for reserve : {}, round : {}", reserveIdForProposal, RoundId+1)
			proposeAddress(accountName, uint64(reserveIdForProposal), uint64(RoundId+1), reserveToBeUpdated.ReserveAddress, oracleAddr, dbconn)
		}

	}
}

func GenerateAndRegisterNewProposedAddress(dbconn *sql.DB, accountName string, height int64, oldReserveAddress string, oracleAddr string) (string, string) {
	newSweepAddress, script := GenerateAddress(height, oldReserveAddress, oracleAddr, dbconn)
	registerAddressOnForkscanner(newSweepAddress)
	hexScript := hex.EncodeToString(script)
	return newSweepAddress, hexScript
}

func generateAndRegisterNewBtcReserveAddress(dbconn *sql.DB, accountName string, height int64, oracleAddr string) string {
	newSweepAddress, reserveScript := GenerateAddress(height, "", oracleAddr, dbconn)
	registerReserveAddressOnNyks(accountName, newSweepAddress, reserveScript, oracleAddr)
	registerAddressOnForkscanner(newSweepAddress)

	return newSweepAddress
}

func registerReserveAddressOnNyks(accountName string, address string, script []byte, oracleAddr string) {

	cosmos := comms.GetCosmosClient()

	reserveScript := hex.EncodeToString(script)

	msg := &bridgetypes.MsgRegisterReserveAddress{
		ReserveScript:  reserveScript,
		ReserveAddress: address,
		JudgeAddress:   oracleAddr,
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
