package judge

import (
	"bytes"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/btcsuite/btcd/wire"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"github.com/twilight-project/forkoracle-go/address"
	"github.com/twilight-project/forkoracle-go/comms"
	db "github.com/twilight-project/forkoracle-go/db"
	btcOracleTypes "github.com/twilight-project/forkoracle-go/types"
	utils "github.com/twilight-project/forkoracle-go/utils"
	bridgetypes "github.com/twilight-project/nyks/x/bridge/types"
)

func generateSweepTx(sweepAddress string, newSweepAddress string,
	accountName string, withdrawRequests []btcOracleTypes.WithdrawRequest,
	unlockHeight int64, utxos []btcOracleTypes.Utxo, dbconn *sql.DB) (string, string, uint64, error) {

	fmt.Println(withdrawRequests)
	fmt.Println("sweep address : ", newSweepAddress)
	number := fmt.Sprintf("%v", viper.Get("sweep_preblock"))
	sweepPreblock, _ := strconv.Atoi(number)

	if len(utxos) <= 0 {
		// need to decide if this needs to be enabled
		// addr := generateAndRegisterNewAddress(accountName, height+noOfMultisigs, sweepAddress.Address)
		fmt.Println("INFO : No funds in address : ", sweepAddress, " generating new address : ")
		db.MarkAddressSignedRefund(dbconn, sweepAddress)
		db.MarkAddressSignedSweep(dbconn, sweepAddress)
		db.MarkAddressArchived(dbconn, sweepAddress)
		return "", "", 0, nil
	}

	totalAmountTxIn := uint64(0)
	totalAmountTxOut := uint64(0)

	sweepTx := wire.NewMsgTx(wire.TxVersion)
	for _, utxo := range utxos {
		txIn, err := utils.CreateTxIn(utxo)
		if err != nil {
			fmt.Println("error while add tx in : ", err)
			return "", "", 0, err
		}
		totalAmountTxIn = totalAmountTxIn + utxo.Amount
		txIn.Sequence = uint32(unlockHeight)
		sweepTx.AddTxIn(txIn)
	}

	txOut, err := utils.CreateTxOut(newSweepAddress, int64(0))
	if err != nil {
		log.Println("error with txout", err)
		return "", "", 0, err
	}
	sweepTx.AddTxOut(txOut)

	for _, withdrawal := range withdrawRequests {
		amount, err := strconv.Atoi(withdrawal.WithdrawAmount)
		if err != nil {
			fmt.Println("error while txout amount conversion : ", err)
			return "", "", 0, err
		}
		txOut, err := utils.CreateTxOut(withdrawal.WithdrawAddress, int64(amount))
		if err != nil {
			fmt.Println("error while txout : ", err)
			return "", "", 0, err
		}
		totalAmountTxOut = totalAmountTxOut + uint64(amount)
		sweepTx.AddTxOut(txOut)
	}

	// need to be worked on
	if int64(totalAmountTxIn-totalAmountTxOut) > 0 {
		sweepTx.TxOut[0].Value = int64(totalAmountTxIn - totalAmountTxOut)
	}

	requiredFee, err := utils.GetFeeFromBtcNode(sweepTx)
	if err != nil {
		fmt.Println("error in getting fee from btc node : ", err)
		return "", "", 0, err
	}

	newReserveOutput := sweepTx.TxOut[0]
	if newReserveOutput.Value < int64(requiredFee) {
		fmt.Println("Change output is smaller than required fee")
		return "", "", 0, nil
	}

	// Deduct the fee from the change output
	newReserveOutput.Value = newReserveOutput.Value - int64(requiredFee)
	sweepTx.TxOut[0] = newReserveOutput

	script := db.QuerySweepAddressScript(dbconn, sweepAddress)
	witness := wire.TxWitness{}
	witness = append(witness, script)
	sweepTx.TxIn[0].Witness = witness
	sweepTx.LockTime = uint32(unlockHeight + int64(sweepPreblock))

	var UnsignedTx bytes.Buffer
	err = sweepTx.Serialize(&UnsignedTx)
	if err != nil {
		fmt.Println("error in serializing sweep tx : ", err)
		return "", "", 0, err
	}
	hexTx := hex.EncodeToString(UnsignedTx.Bytes())
	fmt.Println("transaction UnSigned Sweep: ", hexTx)
	txid := sweepTx.TxHash().String()

	return hexTx, txid, totalAmountTxIn, nil
}

func generateRefundTx(txHex string, script string, reserveId uint64, roundId uint64) (string, error) {
	sweepTx, err := utils.CreateTxFromHex(txHex)
	if err != nil {
		fmt.Println("error decoding tx : ", err)
	}

	inputTx := sweepTx.TxHash().String()
	vout := 0 // since we are always setting the sweep tx at vout = 0

	utxo := btcOracleTypes.Utxo{
		Txid:   inputTx,
		Vout:   uint32(vout),
		Amount: 0,
	}
	refundTx := wire.NewMsgTx(wire.TxVersion)
	txIn, err := utils.CreateTxIn(utxo)
	if err != nil {
		fmt.Println("error while add tx in : ", err)
		return "", err
	}
	txIn.Sequence = wire.MaxTxInSequenceNum - 10
	refundTx.AddTxIn(txIn)

	refundSnapshots := comms.GetRefundSnapshot(reserveId, roundId)
	for _, refund := range refundSnapshots.RefundAccounts {
		amount, err := strconv.Atoi(refund.Amount)
		if err != nil {
			fmt.Println("error in amount of refund snapshot : ", err)
			return "", err
		}
		txout, err := utils.CreateTxOut(refund.BtcDepositAddress, int64(amount))
		if err != nil {
			fmt.Println("error while add tx out : ", err)
			return "", err
		}
		refundTx.AddTxOut(txout)
	}

	// feeRate := getBtcFeeRate()
	// baseSize := sweepTx.SerializeSizeStripped()
	// totalSize := sweepTx.SerializeSize()
	// weight := (baseSize * 3) + totalSize
	// vsize := (weight + 3) / 4

	// // // Calculate the required fee
	// requiredFee := vsize * feeRate.Priority

	// validators := getDelegateAddresses()

	// feeAdjustment := requiredFee / len(validators.Addresses)

	// for i, output := range refundTx.TxOut {
	// 	refundTx.TxOut[i].Value = output.Value - int64(feeAdjustment)
	// }

	locktime := uint32(5)
	refundTx.LockTime = locktime

	scriptbytes, _ := hex.DecodeString(script)
	witness := wire.TxWitness{}
	witness = append(witness, scriptbytes)
	refundTx.TxIn[0].Witness = witness

	var UnsignedTx bytes.Buffer
	err = refundTx.Serialize(&UnsignedTx)
	if err != nil {
		fmt.Println("error in serializing refund tx : ", err)
	}
	hexTx := hex.EncodeToString(UnsignedTx.Bytes())
	fmt.Println("transaction UnSigned Refund: ", hexTx)

	return hexTx, nil
}

func generateSignedSweepTx(accountName string, sweepTx *wire.MsgTx, reserveId uint64, roundId uint64, currentReserveAddress btcOracleTypes.SweepAddress) []byte {
	currentReserveScript := currentReserveAddress.Script
	encoded := hex.EncodeToString(currentReserveScript)
	decodedScript := utils.DecodeBtcScript(encoded)
	minSignsRequired := utils.GetMinSignFromScript(decodedScript)
	if minSignsRequired < 1 {
		fmt.Println("INFO : MinSign required for sweep is 0, which means there is a fault with sweep address script")
		return nil
	}
	pubkeys := utils.GetPublicKeysFromScript(decodedScript, int(minSignsRequired))

	for {
		time.Sleep(30 * time.Second)

		receivedSweepSignatures := comms.GetSignSweep(reserveId, roundId)
		filteredSweepSignatures := utils.FilterAndOrderSignSweep(receivedSweepSignatures, pubkeys)

		if len(filteredSweepSignatures) <= 0 {
			fmt.Println("INFO: ", "no sweep signature found")
			continue
		}

		if len(filteredSweepSignatures)/int(minSignsRequired) < 1 {
			fmt.Println("INFO: ", "not enough sweep signatures")
			continue
		}

		script := currentReserveAddress.Script
		preimage := currentReserveAddress.Preimage

		for i := range sweepTx.TxIn {
			dataSig := make([][]byte, 0)
			for _, sig := range filteredSweepSignatures {
				sig, _ := hex.DecodeString(sig.SweepSignature[i])
				dataSig = append(dataSig, sig)
			}
			witness := wire.TxWitness{}
			witness = append(witness, preimage)
			dummy := []byte{}
			witness = append(witness, dummy)
			for j := 0; j < int(minSignsRequired); j++ {
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

func generateSignedRefundTx(accountName string, refundTx *wire.MsgTx, reserveId uint64, roundId uint64, dbconn *sql.DB, oracleAddr string) ([]byte, btcOracleTypes.SweepAddress, error) {
	addrs := comms.GetProposedSweepAddress(reserveId, roundId)
	if addrs.ProposeSweepAddressMsg.BtcAddress == "" {
		return nil, btcOracleTypes.SweepAddress{}, nil
	}

	addresses := db.QuerySweepAddress(dbconn, addrs.ProposeSweepAddressMsg.BtcAddress)
	if len(addresses) <= 0 {
		fmt.Println("address not found in DB")
		return nil, btcOracleTypes.SweepAddress{}, nil
	}
	newReserveAddress := addresses[0]

	newReserveScript := newReserveAddress.Script
	encoded := hex.EncodeToString(newReserveScript)
	decodedScript := utils.DecodeBtcScript(encoded)
	minSignsRequired := utils.GetMinSignFromScript(decodedScript)
	if minSignsRequired < 1 {
		fmt.Println("INFO : MinSign required for refund is 0, which means there is a fault with sweep address script")
		return nil, btcOracleTypes.SweepAddress{}, errors.New("MinSign required for refund is 0, which means there is a fault with sweep address script")
	}
	pubkeys := utils.GetPublicKeysFromScript(decodedScript, int(minSignsRequired))

	for {
		time.Sleep(30 * time.Second)
		receiveRefundSignatures := comms.GetSignRefund(reserveId, roundId)
		filteredRefundSignatures, JudgeSign := utils.OrderSignRefund(receiveRefundSignatures, newReserveAddress.Address, pubkeys, oracleAddr)

		if len(filteredRefundSignatures) <= 0 {
			continue
		}

		if len(filteredRefundSignatures)/int(minSignsRequired) < 1 {
			fmt.Println("INFO: ", "not enough refund signatures")
			continue
		}

		dataSig := make([][]byte, 0)

		for _, sig := range filteredRefundSignatures {
			sig, _ := hex.DecodeString(sig.RefundSignature[0])
			dataSig = append(dataSig, sig)
		}

		script := newReserveAddress.Script
		preimageFalse, _ := address.Preimage()

		for i := 0; i < len(refundTx.TxIn); i++ {

			witness := wire.TxWitness{}
			witness = append(witness, []byte(JudgeSign.RefundSignature[0]))
			witness = append(witness, preimageFalse)
			dummy := []byte{}
			witness = append(witness, dummy)
			for j := 0; j < int(minSignsRequired); j++ {
				witness = append(witness, dataSig[j])
			}
			witness = append(witness, script)
			refundTx.TxIn[i].Witness = witness
		}

		var signedRefundTx bytes.Buffer
		err := refundTx.Serialize(&signedRefundTx)
		if err != nil {
			fmt.Println("Signed Refund : ", err)
		}

		return signedRefundTx.Bytes(), newReserveAddress, nil
	}
}

//temp use function judge sign

// func signByJudge(tx *wire.MsgTx, script []byte) []string {
// 	return signTx(tx, script)
// }

func InitJudge(accountName string, dbconn *sql.DB, oracleAddr string, valAddr string) {
	fmt.Println("init judge")
	addr := db.QueryAllSweepAddresses(dbconn)
	if len(addr) <= 0 {
		time.Sleep(2 * time.Minute)
		initReserve(accountName, oracleAddr, valAddr, dbconn)
	}
}

func initReserve(accountName string, oracleAddr string, valAddr string, dbconn *sql.DB) {
	fmt.Println("init reserve")
	height := 0

	number := fmt.Sprintf("%v", viper.Get("unlocking_time"))
	unlockingTimeInBlocks, _ := strconv.Atoi(number)

	judges := comms.GetRegisteredJudges()
	if len(judges.Judges) == 0 {
		utils.RegisterJudge(accountName, oracleAddr, valAddr)
	} else {
		registered := false
		for _, judge := range judges.Judges {
			if judge.JudgeAddress == oracleAddr {
				registered = true
			}
		}
		if !registered {
			utils.RegisterJudge(accountName, oracleAddr, valAddr)
		}
	}

	for {
		resp := comms.GetAttestations("1")
		if len(resp.Attestations) <= 0 {
			time.Sleep(30 * time.Second)
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

	_ = address.GenerateAndRegisterNewBtcReserveAddress(dbconn, accountName, int64(height+unlockingTimeInBlocks), oracleAddr)
	fmt.Println("judge initialized")
}

func ProcessSweep(accountName string, dbconn *sql.DB, oracleAddr string) {
	fmt.Println("Process Sweep unsigned started")
	time.Sleep(2 * time.Minute)
	number := fmt.Sprintf("%v", viper.Get("sweep_preblock"))
	sweepInitateBlockHeight, _ := strconv.Atoi(number)

	resp := comms.GetAttestations("20")
	if len(resp.Attestations) <= 0 {
		time.Sleep(1 * time.Minute)
		fmt.Println("no attestaions (start judge)")
		fmt.Println("finishing sweep process")
		return
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
		fmt.Println("finishing sweep process : no judge found")
		return
	}

	for _, attestation := range resp.Attestations {
		height, _ := strconv.Atoi(attestation.Proposal.Height)

		if !attestation.Observed {
			continue
		}

		addresses := db.QuerySweepAddressesByHeight(dbconn, uint64(height+sweepInitateBlockHeight), true)
		if len(addresses) <= 0 {
			continue
		}

		fmt.Println("sweep address found")

		currentSweepAddress := addresses[0]
		utxos := db.QueryUtxo(dbconn, currentSweepAddress.Address)
		if len(utxos) <= 0 {
			// need to decide if this needs to be enabled
			// addr := generateAndRegisterNewAddress(accountName, height+noOfMultisigs, sweepAddress.Address)
			fmt.Println("INFO : No funds in address : ", currentSweepAddress.Address, " generating new address : ")
			fmt.Println("finishing sweep process")
			return
		}

		var newSweepAddress *string
		var reserveTobeProcessed btcOracleTypes.BtcReserve

		minRoundId := 50000000
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
			sweepAddresses := comms.GetProposedSweepAddress(uint64(currentReserveId), uint64(currentRoundId+1))
			if sweepAddresses.ProposeSweepAddressMsg.BtcAddress == "" {
				fmt.Println("no proposed sweep address found ")
				time.Sleep(2 * time.Minute)
				continue
			}
			newSweepAddress = &sweepAddresses.ProposeSweepAddressMsg.BtcAddress
			break
		}

		withdrawRequests := comms.GetWithdrawSnapshot(uint64(currentReserveId), uint64(currentRoundId+1)).WithdrawRequests
		sweepTxHex, sweepTxId, _, err := generateSweepTx(currentSweepAddress.Address, *newSweepAddress, accountName, withdrawRequests, int64(height), utxos, dbconn)
		if err != nil {
			fmt.Println("Error in generating a Sweep transaction: ", err)
			fmt.Println("finishing sweep process")
			return
		}
		if sweepTxHex == "" {
			fmt.Println("INFO: ", "no sweep tx generated because no funds in current address")
			fmt.Println("finishing sweep process")
			time.Sleep(1 * time.Minute)
			return
		}
		cosmos := comms.GetCosmosClient()
		msg := bridgetypes.NewMsgUnsignedTxSweep(sweepTxId, sweepTxHex, uint64(currentReserveId), uint64(currentRoundId+1), oracleAddr)
		comms.SendTransactionUnsignedSweepTx(accountName, cosmos, msg)

		db.MarkAddressArchived(dbconn, currentSweepAddress.Address)
	}

	fmt.Println("finishing sweep process")
}

func ProcessRefund(accountName string, oracleAddr string) {
	fmt.Println("Process unsigned Refund started")

	reserves := comms.GetBtcReserves()
	var currentReservesForThisJudge []btcOracleTypes.BtcReserve
	for _, reserve := range reserves.BtcReserves {
		if reserve.JudgeAddress == oracleAddr {
			currentReservesForThisJudge = append(currentReservesForThisJudge, reserve)
		}
	}

	var reserveTobeProcessed *btcOracleTypes.BtcReserve
	minRoundId := 500000000
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

	sweepTxs := comms.GetUnsignedSweepTx(uint64(reserveIdForRefund), uint64(currentRoundId+1))
	if sweepTxs.Code > 0 {
		fmt.Println("refund : no unsigned sweep tx found : ", reserveIdForRefund, "   ", uint64(currentRoundId+1))
		fmt.Println("finishing refund process")
		return
	}

	sweeptx := sweepTxs.UnsignedTxSweepMsg

	sweepAddresses := comms.GetProposedSweepAddress(uint64(reserveIdForRefund), uint64(currentRoundId+1))
	if sweepAddresses.ProposeSweepAddressMsg.BtcAddress == "" {
		fmt.Println("issue with sweep address while creating refund tx")
		fmt.Println("finishing refund process")
		return
	}

	refundTxHex, err := generateRefundTx(sweeptx.BtcUnsignedSweepTx, sweepAddresses.ProposeSweepAddressMsg.BtcScript, uint64(reserveIdForRefund), uint64(currentRoundId+1))
	if err != nil {
		fmt.Println("issue creating refund tx")
		fmt.Println("finishing refund process")
		return
	}
	cosmos := comms.GetCosmosClient()
	msg := bridgetypes.NewMsgUnsignedTxRefund(uint64(reserveIdForRefund), uint64(currentRoundId+1), refundTxHex, oracleAddr)
	comms.SendTransactionUnsignedRefundTx(accountName, cosmos, msg)

	fmt.Println("finishing refund process")
}

func ProcessSignedSweep(accountName string, oracleAddr string, dbconn *sql.DB) {
	fmt.Println("Process signed sweep started")

	var currentReservesForThisJudge []btcOracleTypes.BtcReserve
	reserves := comms.GetBtcReserves()
	for _, reserve := range reserves.BtcReserves {
		if reserve.JudgeAddress == oracleAddr {
			currentReservesForThisJudge = append(currentReservesForThisJudge, reserve)
		}
	}

	var reserveTobeProcessed btcOracleTypes.BtcReserve
	minRoundId := 500000000
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

	sweepTxs := comms.GetUnsignedSweepTx(uint64(reserveIdForSweepTx), uint64(roundIdForSweepTx))
	if sweepTxs.Code > 0 {
		fmt.Println("Signed Sweep: No Unsigned Sweep tx found : ", reserveIdForSweepTx, "   ", roundIdForSweepTx)
		fmt.Println("finishing signed sweep process")
		return
	}

	unsignedSweepTxHex := sweepTxs.UnsignedTxSweepMsg.BtcUnsignedSweepTx
	sweepTx, err := utils.CreateTxFromHex(unsignedSweepTxHex)
	if err != nil {
		fmt.Println("error decoding sweep tx : inside judge")
		fmt.Println(err)
	}

	reserveAddresses := db.QueryUnsignedSweepAddressByScript(dbconn, sweepTx.TxIn[0].Witness[0])

	if len(reserveAddresses) == 0 {
		fmt.Println("No address found")
		return
	}
	currentReserveAddress := reserveAddresses[0]

	if currentReserveAddress.BroadcastSweep {
		fmt.Println("Sweep tx already broadcasted")
		return
	}

	fmt.Println("Signed Sweep process : starting sign aggregation")

	signedSweepTx := generateSignedSweepTx(accountName, sweepTx, uint64(reserveIdForSweepTx), uint64(roundIdForSweepTx), currentReserveAddress)

	signedSweepTxHex := hex.EncodeToString(signedSweepTx)
	fmt.Println("Signed P2WSH Sweep transaction with preimage:", signedSweepTxHex)

	cosmos := comms.GetCosmosClient()
	msg := &bridgetypes.MsgBroadcastTxSweep{
		SignedSweepTx: signedSweepTxHex,
		JudgeAddress:  oracleAddr,
		ReserveId:     uint64(reserveIdForSweepTx),
		RoundId:       uint64(roundIdForSweepTx),
	}

	comms.SendTransactionBroadcastSweeptx(accountName, cosmos, msg)
	db.InsertSignedtx(dbconn, signedSweepTx, currentReserveAddress.Unlock_height)
	db.MarkAddressBroadcastedSweep(dbconn, currentReserveAddress.Address)
	address.UnRegisterAddressOnForkscanner(currentReserveAddress.Address)

	fmt.Println("finishing signed refund process")

}

func ProcessSignedRefund(accountName string, oracleAddr string, dbconn *sql.DB, WsHub *btcOracleTypes.Hub, latestRefundTxHash *prometheus.GaugeVec) {
	fmt.Println("Process signed Refund started")

	reserves := comms.GetBtcReserves()
	var currentReservesForThisJudge []btcOracleTypes.BtcReserve
	for _, reserve := range reserves.BtcReserves {
		if reserve.JudgeAddress == oracleAddr {
			currentReservesForThisJudge = append(currentReservesForThisJudge, reserve)
		}
	}

	var reserveTobeProcessed *btcOracleTypes.BtcReserve
	minRoundId := 50000000
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

	refundTxs := comms.GetUnsignedRefundTx(int64(reserveIdForSweep), int64(currentRoundId+1))

	if refundTxs.Code > 0 {
		fmt.Println("no unsigned refund tx found")
		fmt.Println("finishing signed refund process")
		return
	}

	unsignedRefundTxHex := refundTxs.UnsignedTxRefundMsg.BtcUnsignedRefundTx
	refundTx, err := utils.CreateTxFromHex(unsignedRefundTxHex)
	if err != nil {
		fmt.Println("error decoding sweep tx : inside judge")
		fmt.Println(err)
	}

	signedRefundTx, newReserveAddress, _ := generateSignedRefundTx(accountName, refundTx, uint64(reserveIdForSweep), uint64(currentRoundId+1), dbconn, oracleAddr)

	signedRefundTxHex := hex.EncodeToString(signedRefundTx)
	fmt.Println("Signed P2WSH Refund transaction with preimage:", signedRefundTxHex)

	cosmos := comms.GetCosmosClient()
	msg := &bridgetypes.MsgBroadcastTxRefund{
		SignedRefundTx: signedRefundTxHex,
		JudgeAddress:   oracleAddr,
		ReserveId:      uint64(reserveIdForSweep),
		RoundId:        uint64(currentRoundId + 1),
	}
	comms.SendTransactionBroadcastRefundtx(accountName, cosmos, msg)
	db.MarkAddressBroadcastedRefund(dbconn, newReserveAddress.Address)

	WsHub.Broadcast <- signedRefundTx

	latestRefundTxHash.Reset()
	latestRefundTxHash.WithLabelValues(refundTx.TxHash().String()).Set(float64(currentReserveId))

	fmt.Println("finishing signed refund process")
}

func BroadcastOnBtc(dbconn *sql.DB) {
	fmt.Println("Started Btc Broadcaster")
	for {
		resp := comms.GetAttestations("3")
		if len(resp.Attestations) <= 0 {
			time.Sleep(1 * time.Minute)
			fmt.Println("no attestaions (btc broadcaster)")
			continue
		}

		for _, attestation := range resp.Attestations {
			if !attestation.Observed {
				continue
			}
			height, _ := strconv.Atoi(attestation.Proposal.Height)
			txs := db.QuerySignedTx(dbconn, int64(height))
			for _, tx := range txs {
				transaction := hex.EncodeToString(tx)
				wireTransaction, err := utils.CreateTxFromHex(transaction)
				if err != nil {
					fmt.Println("error decodeing signed transaction btc broadcaster : ", err)
				}
				utils.BroadcastBtcTransaction(wireTransaction)
				db.DeleteSignedTx(dbconn, tx)
			}
		}
	}
}
