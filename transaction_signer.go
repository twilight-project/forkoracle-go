package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	comms "github.com/twilight-project/forkoracle-go/comms"
	db "github.com/twilight-project/forkoracle-go/db"
	eventhandler "github.com/twilight-project/forkoracle-go/eventhandler"
	utils "github.com/twilight-project/forkoracle-go/utils"
	wallet "github.com/twilight-project/forkoracle-go/wallet"
	bridgetypes "github.com/twilight-project/nyks/x/bridge/types"
	"github.com/tyler-smith/go-bip32"
)

func processTxSigningSweep(accountName string, masterPrivateKey *bip32.Key, dbconn *sql.DB) {
	fmt.Println("starting Sweep Tx Signer")
	SweepTxs := comms.GetAllUnsignedSweepTx()

	for _, tx := range SweepTxs.UnsignedTxSweepMsgs {
		reserveId, _ := strconv.Atoi(tx.ReserveId)
		roundId, _ := strconv.Atoi(tx.RoundId)

		for {
			correspondingRefundTx := comms.GetBroadCastedRefundTx(uint64(reserveId), uint64(roundId))
			if correspondingRefundTx.ReserveId != "" {
				break
			}
			fmt.Printf("corresponding refund tx not found  reserve Id: %d roundid : %d\n", reserveId, roundId)
			time.Sleep(1 * time.Minute)
		}

		fmt.Printf("corresponding refund tx found  reserve Id: %d roundid : %d\n", reserveId, roundId)

		sweepTx, err := utils.CreateTxFromHex(tx.BtcUnsignedSweepTx)
		if err != nil {
			fmt.Println("error decoding sweep tx : inside processSweepTx : ", err)
			log.Fatal(err)
		}

		addresses := db.QueryUnsignedSweepAddressByScript(dbconn, sweepTx.TxIn[0].Witness[0])
		if len(addresses) <= 0 {
			fmt.Println("signing: no address")
			continue
		}

		reserveAddress := addresses[0]

		if reserveAddress.Signed_sweep {
			continue
		}

		sweepSignatures := utils.SignTx(dbconn, masterPrivateKey, sweepTx, reserveAddress.Script)

		fmt.Println("Sweep Signature : ", sweepSignatures)
		cosmos := comms.GetCosmosClient()
		msg := &bridgetypes.MsgSignSweep{
			ReserveId:        uint64(reserveId),
			RoundId:          uint64(roundId),
			SignerPublicKey:  wallet.GetBtcPublicKey(masterPrivateKey),
			SweepSignature:   sweepSignatures,
			BtcOracleAddress: reserveAddress.Address,
		}

		comms.SendTransactionSignSweep(accountName, cosmos, msg)

		db.MarkAddressSignedSweep(dbconn, reserveAddress.Address)
		db.InsertTransaction(dbconn, sweepTx.TxHash().String(), reserveAddress.Address, uint64(reserveId), uint64(roundId))
	}

	fmt.Println("finishing sweep tx signer")
}

func processTxSigningRefund(accountName string, masterPrivateKey *bip32.Key, dbconn *sql.DB) {
	fmt.Println("starting Refund Tx Signer")
	refundTxs := comms.GetAllUnsignedRefundTx()

	for _, tx := range refundTxs.UnsignedTxRefundMsgs {
		refundTx, err := utils.CreateTxFromHex(tx.BtcUnsignedRefundTx)
		if err != nil {
			fmt.Println("error decoding sweep tx : inside processSweepTx : ", err)
			log.Fatal(err)
		}

		fmt.Println("signing refund tx")
		addresses := db.QueryUnsignedRefundAddressByScript(dbconn, refundTx.TxIn[0].Witness[0])
		if len(addresses) <= 0 {
			continue
		}
		reserveAddress := addresses[0]

		if reserveAddress.Signed_refund {
			continue
		}
		refundSignature := utils.SignTx(dbconn, masterPrivateKey, refundTx, reserveAddress.Script)

		reserveId, _ := strconv.Atoi(tx.ReserveId)
		roundId, _ := strconv.Atoi(tx.RoundId)

		fmt.Println("Refund Signature : ", refundSignature)
		cosmos := comms.GetCosmosClient()
		msg := &bridgetypes.MsgSignRefund{
			ReserveId:        uint64(reserveId),
			RoundId:          uint64(roundId),
			SignerPublicKey:  wallet.GetBtcPublicKey(masterPrivateKey),
			RefundSignature:  []string{refundSignature[0]},
			BtcOracleAddress: reserveAddress.Address,
		}

		comms.SendTransactionSignRefund(accountName, cosmos, msg)

		db.MarkAddressSignedRefund(dbconn, reserveAddress.Address)
	}

	fmt.Println("finishing refund tx signer")
}

func startTransactionSigner(accountName string, masterPrivateKeys *bip32.Key, dbconn *sql.DB) {
	fmt.Println("starting Transaction Signer")
	go eventhandler.NyksEventListener("unsigned_tx_refund", accountName, "signing_refund", masterPrivateKey, dbconn)
	eventhandler.NyksEventListener("broadcast_tx_refund", accountName, "signing_sweep", masterPrivateKey, dbconn)
	fmt.Println("finishing bridge")
}
