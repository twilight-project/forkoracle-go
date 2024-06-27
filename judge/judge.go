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

	// requiredFee , err := utils.GetFeeFromBtcNode(sweepTx)
	// if err != nil {
	// 	fmt.Println("error in getting fee from btc node : ", err)
	// 	return "", "", 0, err
	// }

	// newReserveOutput := sweepTx.TxOut[0]
	// if newReserveOutput.Value < int64(requiredFee) {
	// 	fmt.Println("Change output is smaller than required fee")
	// 	return "", "", 0, nil
	// }

	// // Deduct the fee from the change output
	// newReserveOutput.Value = newReserveOutput.Value - int64(requiredFee)
	// sweepTx.TxOut[0] = newReserveOutput

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

func generateSignedSweepTx(accountName string, sweepTx *wire.MsgTx, reserveId uint64, roundId uint64, currentReserveAddress btcOracleTypes.SweepAddress, judgeAddr string) []byte {
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
		filteredSweepSignatures := utils.FilterAndOrderSignSweep(receivedSweepSignatures, pubkeys, judgeAddr)

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

func generateSignedRefundTx(accountName string, refundTx *wire.MsgTx, reserveId uint64, roundId uint64, dbconn *sql.DB, judgeAddr string) ([]byte, btcOracleTypes.SweepAddress, error) {
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
		filteredRefundSignatures, JudgeSign := utils.OrderSignRefund(receiveRefundSignatures, newReserveAddress.Address, pubkeys, judgeAddr)

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

// func InitJudge(accountName string, dbconn *sql.DB, oracleAddr string, valAddr string) {
// 	fmt.Println("init judge")
// 	addr := db.QueryAllSweepAddresses(dbconn)
// 	if len(addr) <= 0 {
// 		time.Sleep(2 * time.Minute)
// 		initReserve(accountName, oracleAddr, valAddr, dbconn)
// 	}
// }

func InitReserve(accountName string, judgeAddr string, valAddr string, dbconn *sql.DB) {
	fmt.Println("init reserve")
	height := 0

	number := fmt.Sprintf("%v", viper.Get("unlocking_time"))
	unlockingTimeInBlocks, _ := strconv.Atoi(number)

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
		panic("No fragment found with the this judge address")
	}

	threshold, err := strconv.Atoi(fragment.Threshold)
	if err != nil {
		fmt.Println("error converting to int : ", err)
		panic(err)
	}

	if len(fragment.Signers) < threshold {
		fmt.Println("INFO : Not enough signers to initialize reserve, exiting...")
		panic("not enough signers")
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
	fragmentId, _ := strconv.Atoi(fragment.FragmentId)

	_ = address.GenerateAndRegisterNewBtcReserveAddress(dbconn, accountName, int64(height+unlockingTimeInBlocks), judgeAddr, fragmentId)
	fmt.Println("judge initialized")
}

func ProcessSweep(accountName string, dbconn *sql.DB, judgeAddr string) {
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
			fmt.Println("INFO : No funds in address : ", currentSweepAddress.Address, " generating new address : ")
			fmt.Println("finishing sweep process: no funds")
			return
		}

		var newSweepAddress *string
		var reserve btcOracleTypes.BtcReserve

		for _, r := range comms.GetBtcReserves().BtcReserves {
			if r.ReserveAddress == currentSweepAddress.Address {
				reserve = r
			}
		}

		roundId, _ := strconv.Atoi(reserve.RoundId)
		reserveId, _ := strconv.Atoi(reserve.ReserveId)

		for {
			sweepAddresses := comms.GetProposedSweepAddress(uint64(reserveId), uint64(roundId+1))
			if sweepAddresses.ProposeSweepAddressMsg.BtcAddress == "" {
				fmt.Println("no proposed sweep address found ")
				time.Sleep(2 * time.Minute)
				continue
			}
			newSweepAddress = &sweepAddresses.ProposeSweepAddressMsg.BtcAddress
			break
		}

		withdrawRequests := comms.GetWithdrawSnapshot(uint64(reserveId), uint64(roundId+1)).WithdrawRequests
		sweepTxHex, sweepTxId, _, err := generateSweepTx(currentSweepAddress.Address, *newSweepAddress, accountName, withdrawRequests, int64(height), utxos, dbconn)
		if err != nil {
			fmt.Println("Error in generating a Sweep transaction: ", err)
			fmt.Println("finishing sweep process: error in generating a Sweep transaction")
			return
		}
		if sweepTxHex == "" {
			fmt.Println("INFO: ", "no sweep tx generated because no funds in current address")
			fmt.Println("finishing sweep process")
			time.Sleep(1 * time.Minute)
			return
		}
		cosmos := comms.GetCosmosClient()
		msg := bridgetypes.NewMsgUnsignedTxSweep(sweepTxId, sweepTxHex, uint64(reserveId), uint64(roundId+1), judgeAddr)
		comms.SendTransactionUnsignedSweepTx(accountName, cosmos, msg)

		db.MarkAddressArchived(dbconn, currentSweepAddress.Address)
		break
	}

	fmt.Println("finishing sweep process: final")
}

func ProcessRefund(accountName string, judgeAddr string, dbconn *sql.DB) {
	fmt.Println("Process unsigned Refund started")

	fragments := comms.GetAllFragments()
	var fragment btcOracleTypes.Fragment
	for _, f := range fragments.Fragments {
		if f.JudgeAddress == judgeAddr {
			fragment = f
			break
		}
	}

	reserves := comms.GetBtcReserves()
	var ownedReserves []btcOracleTypes.BtcReserve
	for _, r := range reserves.BtcReserves {
		if utils.StringInSlice(r.ReserveId, fragment.ReserveIds) {
			ownedReserves = append(ownedReserves, r)
		}
	}

	var reserve btcOracleTypes.BtcReserve
	found := false
	for _, r := range ownedReserves {
		resId, _ := strconv.Atoi(r.ReserveId)
		roundId, _ := strconv.Atoi(r.RoundId)
		sweepTxs := comms.GetUnsignedSweepTx(uint64(resId), uint64(roundId+1))
		if sweepTxs.Code > 0 {
			fmt.Println("refund : no unsigned sweep tx found : ", resId, "   ", uint64(roundId+1))
			fmt.Println("finishing refund process")
			continue
		}
		reserve = r
		found = true
		break
	}
	if !found {
		fmt.Println("No Sweep tx found")
		return
	}

	reserveId, _ := strconv.Atoi(reserve.ReserveId)
	roundId, _ := strconv.Atoi(reserve.RoundId)

	sweepTxs := comms.GetUnsignedSweepTx(uint64(reserveId), uint64(roundId+1))
	sweeptx := sweepTxs.UnsignedTxSweepMsg

	sweepAddresses := comms.GetProposedSweepAddress(uint64(reserveId), uint64(roundId+1))
	if sweepAddresses.ProposeSweepAddressMsg.BtcAddress == "" {
		fmt.Println("issue with sweep address while creating refund tx")
		fmt.Println("finishing refund process")
		return
	}

	refundTxHex, err := generateRefundTx(sweeptx.BtcUnsignedSweepTx, sweepAddresses.ProposeSweepAddressMsg.BtcScript, uint64(reserveId), uint64(roundId+1))
	if err != nil {
		fmt.Println("issue creating refund tx")
		fmt.Println("finishing refund process")
		return
	}
	cosmos := comms.GetCosmosClient()
	msg := bridgetypes.NewMsgUnsignedTxRefund(uint64(reserveId), uint64(roundId+1), refundTxHex, judgeAddr)
	comms.SendTransactionUnsignedRefundTx(accountName, cosmos, msg)

	fmt.Println("finishing refund process")
}

func ProcessSignedSweep(accountName string, judgeAddr string, dbconn *sql.DB) {
	fmt.Println("Process signed sweep started")

	fragments := comms.GetAllFragments()
	var fragment btcOracleTypes.Fragment
	for _, f := range fragments.Fragments {
		if f.JudgeAddress == judgeAddr {
			fragment = f
			break
		}
	}

	reserves := comms.GetBtcReserves()
	var ownedReserves []btcOracleTypes.BtcReserve
	for _, r := range reserves.BtcReserves {
		if utils.StringInSlice(r.ReserveId, fragment.ReserveIds) {
			ownedReserves = append(ownedReserves, r)
		}
	}

	var reserve btcOracleTypes.BtcReserve
	found := false
	for _, r := range ownedReserves {
		resId, _ := strconv.Atoi(r.ReserveId)
		roundId, _ := strconv.Atoi(r.RoundId)
		sweepTxs := comms.GetUnsignedSweepTx(uint64(resId), uint64(roundId+1))
		if sweepTxs.Code > 0 {
			fmt.Println("refund : no unsigned sweep tx found : ", resId, "   ", uint64(roundId+1))
			fmt.Println("finishing refund process")
			continue
		}
		reserve = r
		found = true
		break
	}

	if !found {
		fmt.Println("No Sweep tx found")
		return
	}

	reserveId, _ := strconv.Atoi(reserve.ReserveId)
	roundId, _ := strconv.Atoi(reserve.RoundId)

	sweepTxs := comms.GetUnsignedSweepTx(uint64(reserveId), uint64(roundId+1))
	if sweepTxs.Code > 0 {
		fmt.Println("Signed Sweep: No Unsigned Sweep tx found : ", reserveId, "   ", roundId+1)
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

	signedSweepTx := generateSignedSweepTx(accountName, sweepTx, uint64(reserveId), uint64(roundId+1), currentReserveAddress, judgeAddr)

	signedSweepTxHex := hex.EncodeToString(signedSweepTx)
	fmt.Println("Signed P2WSH Sweep transaction with preimage:", signedSweepTxHex)

	cosmos := comms.GetCosmosClient()
	msg := &bridgetypes.MsgBroadcastTxSweep{
		SignedSweepTx: signedSweepTxHex,
		JudgeAddress:  judgeAddr,
		ReserveId:     uint64(reserveId),
		RoundId:       uint64(roundId + 1),
	}

	comms.SendTransactionBroadcastSweeptx(accountName, cosmos, msg)
	db.InsertSignedtx(dbconn, signedSweepTx, currentReserveAddress.Unlock_height)
	db.MarkAddressBroadcastedSweep(dbconn, currentReserveAddress.Address)
	address.UnRegisterAddressOnForkscanner(currentReserveAddress.Address)

	fmt.Println("finishing signed sweep process")

}

func ProcessSignedRefund(accountName string, judgeAddr string, dbconn *sql.DB, WsHub *btcOracleTypes.Hub, latestRefundTxHash *prometheus.GaugeVec) {
	fmt.Println("Process signed Refund started")

	fragments := comms.GetAllFragments()
	var fragment btcOracleTypes.Fragment
	for _, f := range fragments.Fragments {
		if f.JudgeAddress == judgeAddr {
			fragment = f
			break
		}
	}

	reserves := comms.GetBtcReserves()
	var ownedReserves []btcOracleTypes.BtcReserve
	for _, r := range reserves.BtcReserves {
		if utils.StringInSlice(r.ReserveId, fragment.ReserveIds) {
			ownedReserves = append(ownedReserves, r)
		}
	}

	var reserve btcOracleTypes.BtcReserve
	found := false
	for _, r := range ownedReserves {
		resId, _ := strconv.Atoi(r.ReserveId)
		roundId, _ := strconv.Atoi(r.RoundId)
		sweepTxs := comms.GetUnsignedRefundTx(int64(resId), int64(roundId+1))
		if sweepTxs.Code > 0 {
			fmt.Println("refund : no unsigned refund tx found : ", resId, "   ", uint64(roundId+1))
			fmt.Println("finishing signed refund process")
			continue
		}
		reserve = r
		found = true
		break
	}

	if !found {
		fmt.Println("No refund tx found")
		return
	}

	reserveId, _ := strconv.Atoi(reserve.ReserveId)
	roundId, _ := strconv.Atoi(reserve.RoundId)

	refundTxs := comms.GetUnsignedRefundTx(int64(reserveId), int64(roundId+1))

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

	signedRefundTx, newReserveAddress, _ := generateSignedRefundTx(accountName, refundTx, uint64(reserveId), uint64(roundId+1), dbconn, judgeAddr)

	signedRefundTxHex := hex.EncodeToString(signedRefundTx)
	fmt.Println("Signed P2WSH Refund transaction with preimage:", signedRefundTxHex)

	cosmos := comms.GetCosmosClient()
	msg := &bridgetypes.MsgBroadcastTxRefund{
		SignedRefundTx: signedRefundTxHex,
		JudgeAddress:   judgeAddr,
		ReserveId:      uint64(reserveId),
		RoundId:        uint64(roundId + 1),
	}
	comms.SendTransactionBroadcastRefundtx(accountName, cosmos, msg)
	db.MarkAddressBroadcastedRefund(dbconn, newReserveAddress.Address)

	// WsHub.broadcast <- signedRefundTx

	// latestRefundTxHash.Reset()
	// latestRefundTxHash.WithLabelValues(refundTx.TxHash().String()).Set(float64(currentReserveId))

	fmt.Println("finishing signed refund process")
}

// func BroadcastOnBtc(dbconn *sql.DB) {
// 	fmt.Println("Started Btc Broadcaster")
// 	for {
// 		resp := comms.GetAttestations("3")
// 		if len(resp.Attestations) <= 0 {
// 			time.Sleep(1 * time.Minute)
// 			fmt.Println("no attestaions (btc broadcaster)")
// 			continue
// 		}

// 		for _, attestation := range resp.Attestations {
// 			if !attestation.Observed {
// 				continue
// 			}
// 			height, _ := strconv.Atoi(attestation.Proposal.Height)
// 			txs := db.QuerySignedTx(dbconn, int64(height))
// 			for _, tx := range txs {
// 				transaction := hex.EncodeToString(tx)
// 				wireTransaction, err := utils.CreateTxFromHex(transaction)
// 				if err != nil {
// 					fmt.Println("error decodeing signed transaction btc broadcaster : ", err)
// 				}
// 				utils.BroadcastBtcTransaction(wireTransaction)
// 				db.DeleteSignedTx(dbconn, tx)
// 			}
// 		}
// 	}
// }
