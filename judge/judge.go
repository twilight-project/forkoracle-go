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

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
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
	unlockHeight int64, utxos []btcOracleTypes.Utxo, dbconn *sql.DB) (string, string, string, uint64, error) {

	wallet := viper.GetString("wallet_name")
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
		return "", "", "", 0, nil
	}

	var inputs []comms.TxInput
	var outputs []comms.TxOutput
	totalAmountTxIn := uint64(0)
	totalAmountTxOut := uint64(0)

	for _, u := range utxos {													//ideally the height should be masked with 0x0000ffff
		inputs = append(inputs, comms.TxInput{Txid: u.Txid, Vout: int64(u.Vout), Sequence: int64(wire.MaxTxInSequenceNum - 10)})
		totalAmountTxIn += u.Amount
	}

	for _, withdrawal := range withdrawRequests {
		a, err := strconv.Atoi(withdrawal.WithdrawAmount)
		if err != nil {
			fmt.Println("error while txout amount conversion : ", err)
			return "", "", "", 0, err
		}
		amount := utils.SatsToBtc(int64(a))
		outputs = append(outputs, comms.TxOutput{withdrawal.WithdrawAddress: float64(amount)})
		totalAmountTxOut = totalAmountTxOut + uint64(a)
	}

	c := totalAmountTxIn - totalAmountTxOut - 1000
	change := utils.SatsToBtc(int64(c))
	outputs = append([]comms.TxOutput{comms.TxOutput{newSweepAddress: float64(change)}}, outputs...)
	locktime := uint32(unlockHeight + int64(sweepPreblock))

	hexTx, err := comms.CreateRawTx(inputs, outputs, locktime, wallet)
	if err != nil {
		fmt.Println("error in creating raw tx : ", err)
		return "", "", "", 0, err
	}

	p, err := comms.CreatePsbt(inputs, outputs, locktime, wallet)
	if err != nil {
		fmt.Println("error in creating psbt : ", err)
		return "", "", "", 0, err
	}

	fmt.Println("transaction base64 psbt: ", p)

	psbt, err := utils.Base64ToHex(p)
	if err != nil {
		fmt.Println("error in converting psbt to hex : ", err)
		return "", "", "", 0, err
	}

	sweepTx, err := utils.CreateTxFromHex(hexTx)
	if err != nil {
		fmt.Println("error decoding tx : ", err)
		return "", "", "", 0, err
	}

	fmt.Println("transaction hex psbt: ", psbt)
	fmt.Println("transaction UnSigned Sweep: ", hexTx)
	return hexTx, psbt, sweepTx.TxHash().String(), totalAmountTxIn, nil
}

func generateRefundTx(txHex string, reserveId uint64, roundId uint64) (string, string, error) {
	wallet := viper.GetString("wallet_name")
	sweepTx, err := utils.CreateTxFromHex(txHex)
	if err != nil {
		fmt.Println("error decoding tx : ", err)
	}

	inputTx := sweepTx.TxHash().String()
	vout := 0 // since we are always setting the sweep tx at vout = 0

	var inputs []comms.TxInput
	var outputs []comms.TxOutput

	inputs = append(inputs, comms.TxInput{Txid: inputTx, Vout: int64(vout), Sequence: int64(wire.MaxTxInSequenceNum - 10)})

	refundSnapshots := comms.GetRefundSnapshot(reserveId, roundId)
	for _, refund := range refundSnapshots.RefundAccounts {
		a, err := strconv.Atoi(refund.Amount)
		if err != nil {
			fmt.Println("error in amount of refund snapshot : ", err)
			return "", "", err
		}

		outputs = append(outputs, comms.TxOutput{refund.BtcDepositAddress: float64(a)})
	}

	proposedAddress := comms.GetProposedSweepAddress(reserveId, roundId)
	if proposedAddress.ProposeSweepAddressMsg.BtcAddress == "" {
		fmt.Println("no proposed sweep address found")
		return "", "", errors.New("no proposed sweep address found")
	}

	script := utils.DecodeBtcScript(proposedAddress.ProposeSweepAddressMsg.BtcScript)
	height := utils.GetUnlockHeightFromScript(script)

	locktime := height + viper.GetInt64("sweep_preblock")

	refundTx, err := comms.CreateRawTx(inputs, outputs, uint32(locktime), wallet)
	if err != nil {
		fmt.Println("error in creating refund tx : ", err)
		return "", "", err
	}
	fmt.Println("transaction UnSigned Refund: ", refundTx)

	fmt.Println(inputs[0])
	fmt.Println(outputs)
	fmt.Println(locktime)
	fmt.Println(wallet)

	scriptPubKey := sweepTx.TxOut[0].PkScript
	fmt.Println("scriptPubKey : ", scriptPubKey)
	amount := sweepTx.TxOut[0].Value
	fmt.Println("amount : ", amount)

	_, addresses, _, err := txscript.ExtractPkScriptAddrs(scriptPubKey, &chaincfg.MainNetParams)
	if err != nil {
		log.Fatal(err)
	}

	if len(addresses) <= 0 {
		fmt.Println("error in extracting address from script")
	}

	fmt.Println("address : ", addresses[0].String())
	addrInfo, err := comms.GetAddressInfo(addresses[0].String(), wallet)
	if err != nil {
		fmt.Println("error in getting address info : ", err)
	}

	fmt.Println("desc : ", addrInfo.Desc)

	p, err := comms.CreatePsbtV1(inputs[0], outputs, uint32(locktime), scriptPubKey, amount)
	fmt.Println(p)
	psbt, _ := p.B64Encode()
	if err != nil {
		fmt.Println("error in converting psbt to base64 : ", err)
		return "", "", err
	}

	psbt, err = comms.UtxoUpdatePsbt(psbt, addrInfo.Desc, wallet)
	if err != nil {
		fmt.Println("error in updating psbt : ", err)
	}

	fmt.Println("transaction base64 refund psbt: ", psbt)

	return refundTx, psbt, nil
}

func generateSignedSweepTx(accountName string, sweepTx *wire.MsgTx, reserveId uint64, roundId uint64, currentReserveAddress btcOracleTypes.SweepAddress, judgeAddr string) []byte {
	wallet := viper.GetString("judge_btc_wallet_name")
	currentReserveScript := string(currentReserveAddress.Script)
	//encoded := hex.EncodeToString(currentReserveScript)
	fmt.Println("currentReserveScript in GenerateSignedSweepTx \n : ", currentReserveScript)
	decodedScript := utils.DecodeBtcScript(currentReserveScript)
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

		script, _ := hex.DecodeString(currentReserveScript)
		preimage := currentReserveAddress.Preimage

		// remove after watchtower is done
		txHex := comms.GetUnsignedSweepTx(reserveId, roundId).UnsignedTxSweepMsg.BtcUnsignedSweepTx
		// the Sweep tx sent to the chain is in Hex format
		// encode it into base64 before passing to the decodePsbt function
		psbt, _ := utils.HexToBase64(txHex)
		psbtStruct, err := comms.DecodePsbt(psbt, wallet)
		if err != nil {
			fmt.Println("error decoding psbt : inside processSweep Watchtower : ", err)
			return nil
		}
		currentReserveScript = psbtStruct.Inputs[0].WitnessScript.Asm 
		signedPsbt, err := comms.SignPsbt(psbt, wallet)
		if err != nil {
			fmt.Println("error signing psbt : inside processSweep Watchtower : ", err)
			return nil
		}

		if len(signedPsbt) <= 0 {
			fmt.Println("error signing psbt : inside processSweep Watchtower : ", err)
			return nil
		}
		watchtowerSig, _ := hex.DecodeString(signedPsbt[0])

		//////////////

		dummy := []byte{}
		for i := range sweepTx.TxIn {
			dataSig := make([][]byte, 0)
			for _, sig := range filteredSweepSignatures {
				sig, _ := hex.DecodeString(sig.SweepSignature[i])
				dataSig = append(dataSig, sig)
			}

			witness := wire.TxWitness{}
			witness = append(witness, watchtowerSig)
			witness = append(witness, dummy)
			witness = append(witness, preimage)
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
		err = sweepTx.Serialize(&signedSweepTx)
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

	newReserveScript := string(newReserveAddress.Script)
	//encoded := hex.EncodeToString(newReserveScript)
	fmt.Println("newReserveScript in GenerateSignedRefundTx \n : ", newReserveScript)
	decodedScript := utils.DecodeBtcScript(newReserveScript)
	minSignsRequired := utils.GetMinSignFromScript(decodedScript)
	if minSignsRequired < 1 {
		fmt.Println("INFO : MinSign required for refund is 0, which means there is a fault with sweep address script")
		return nil, btcOracleTypes.SweepAddress{}, errors.New("MinSign required for refund is 0, which means there is a fault with sweep address script")
	}
	pubkeys := utils.GetPublicKeysFromScript(decodedScript, int(minSignsRequired))

	for {
		time.Sleep(30 * time.Second)
		receiveRefundSignatures := comms.GetSignRefund(reserveId, roundId)
		fmt.Println("received refund signatures : \n", receiveRefundSignatures)
		filteredRefundSignatures := utils.OrderSignRefund(receiveRefundSignatures, newReserveAddress.Address, pubkeys, judgeAddr)

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
		dummy := []byte{}

		for i := 0; i < len(refundTx.TxIn); i++ {

			witness := wire.TxWitness{}
			witness = append(witness, dummy)
			witness = append(witness, dummy)
			witness = append(witness, preimageFalse)
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
	time.Sleep(5 * time.Minute)
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
		sweepTxHex, psbt, sweepTxId, _, err := generateSweepTx(currentSweepAddress.Address, *newSweepAddress, accountName, withdrawRequests, int64(height), utxos, dbconn)
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
		msg := bridgetypes.NewMsgUnsignedTxSweep(sweepTxId, psbt, uint64(reserveId), uint64(roundId+1), judgeAddr)
		comms.SendTransactionUnsignedSweepTx(accountName, cosmos, msg)
		db.InsertUnSignedSweeptx(dbconn, sweepTxHex, int64(reserveId), int64(roundId+1))
		db.MarkAddressArchived(dbconn, currentSweepAddress.Address)
		break
	}

	fmt.Println("finishing sweep process: final")
}

func ProcessRefund(accountName string, judgeAddr string, dbconn *sql.DB) {
	fmt.Println("Process unsigned Refund started")
    time.Sleep(2 * time.Minute)
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
		s, err := db.QueryUnSignedSweeptx(dbconn, int64(resId), int64(roundId+1))
		if err != nil {
			fmt.Println("error in getting unsigned sweep tx : ", err)
			fmt.Println("finishing process refund")
			return
		}
		if len(s) <= 0 {
			fmt.Println("no unsigned sweep tx found")
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

	fmt.Println("Process unsigned Refund=================")
	fmt.Println("Reserve ID : ", reserveId)
	fmt.Println("Round ID : ", roundId)

	sweepTxs, err := db.QueryUnSignedSweeptx(dbconn, int64(reserveId), int64(roundId+1))
	fmt.Println("sweep tx from DB in refund process: ", sweepTxs)
	if err != nil {
		fmt.Println("error in getting unsigned sweep tx : ", err)
		fmt.Println("finishing process refund")
		return
	}
	if len(sweepTxs) <= 0 {
		fmt.Println("no unsigned sweep tx found")
		return
	}

	sweepAddresses := comms.GetProposedSweepAddress(uint64(reserveId), uint64(roundId+1))
	fmt.Println("sweep address from chain: ", sweepAddresses.ProposeSweepAddressMsg.BtcAddress)
	if sweepAddresses.ProposeSweepAddressMsg.BtcAddress == "" {
		fmt.Println("issue with sweep address while creating refund tx")
		fmt.Println("finishing refund process")
		return
	}

	refundTxHex, psbt, err := generateRefundTx(sweepTxs[0].Tx, uint64(reserveId), uint64(roundId+1))
	if err != nil {
		fmt.Println("issue creating refund tx")
		fmt.Println("finishing refund process")
		return
	}

	cosmos := comms.GetCosmosClient()
	msg := bridgetypes.NewMsgUnsignedTxRefund(uint64(reserveId), uint64(roundId+1), psbt, judgeAddr)
	comms.SendTransactionUnsignedRefundTx(accountName, cosmos, msg)
	db.InsertUnSignedRefundtx(dbconn, refundTxHex, int64(reserveId), int64(roundId+1))
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

	s, err := db.QueryUnSignedSweeptx(dbconn, int64(reserveId), int64(reserveId))
	if err != nil {
		fmt.Println("error in getting unsigned sweep tx : ", err)
		fmt.Println("finishing signed sweep process")
		return
	}

	if len(s) <= 0 {
		fmt.Println("no unsigned sweep tx found")
		return
	}

	sweepTx, err := utils.CreateTxFromHex(s[0].Tx)
	if err != nil {
		fmt.Println("error decoding sweep txHex in ProcessSignedSweep: inside judge")
		fmt.Println(err)
	}
	// get the current reserve address / input address for the sweep tx
	//reserveAddresses := db.QueryUnsignedSweepAddressByScript(dbconn, string(sweepTx.TxIn[0].Witness[0]))
	sweepAddresses := comms.GetProposedSweepAddress(uint64(reserveId), uint64(roundId+1))
	proposeSweepAddressBTC:=  sweepAddresses.ProposeSweepAddressMsg.BtcAddress
	// get DB info about the address
	sweepAddressDB := db.QuerySweepAddress(dbconn, proposeSweepAddressBTC)
	if len(sweepAddressDB) <= 0 {
		fmt.Println("No address found")
		return
	}
	// get the parent address of the sweep address
	reserverAddressStr := sweepAddressDB[0].Parent_address
	// get the reserve address info from the DB
	reserveAddresses := db.QuerySweepAddress(dbconn, reserverAddressStr)
	if len(reserveAddresses) <= 0 {
		fmt.Println("No address found")
		return
	}
	// if len(reserveAddresses) == 0 {
	// 	fmt.Println("No address found")
	// 	return
	// }
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
		refundTxs, err := db.QueryUnSignedRefundtx(dbconn, int64(resId), int64(roundId+1))
		if err != nil {
			fmt.Println("error in getting unsigned refund tx : ", err)
			fmt.Println("finishing signed refund process")
			continue
		}

		if len(refundTxs) <= 0 {
			fmt.Println("no unsigned refund tx found")
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

	refundTxs, err := db.QueryUnSignedRefundtx(dbconn, int64(reserveId), int64(roundId+1))
	if err != nil {
		fmt.Println("error in getting unsigned refund tx : ", err)
		fmt.Println("finishing signed refund process with error")
		
		return
	}
	if len(refundTxs) <= 0 {
		fmt.Println("no unsigned refund tx found in the database")
		fmt.Println("finishing signed refund process with error")
		return
	}

	unsignedRefundTxHex := refundTxs[0].Tx
	fmt.Println("unsigned refund tx hex in Process refundTx: \n", unsignedRefundTxHex)
	refundTx, err := utils.CreateTxFromHex(unsignedRefundTxHex)
	if err != nil {
		fmt.Println("error decoding refund txhex in processSignedRefund: inside judge")
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
