package transaction_signer

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	comms "github.com/twilight-project/forkoracle-go/comms"
	db "github.com/twilight-project/forkoracle-go/db"
	utils "github.com/twilight-project/forkoracle-go/utils"
	bridgetypes "github.com/twilight-project/nyks/x/bridge/types"
)

func ProcessTxSigningSweep(accountName string, dbconn *sql.DB, oracleAddr string) {
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

		sweepSignatures := utils.SignTx(sweepTx, reserveAddress.Script)

		fmt.Println("Sweep Signature : ", sweepSignatures)
		cosmos := comms.GetCosmosClient()
		msg := &bridgetypes.MsgSignSweep{
			ReserveId:        uint64(reserveId),
			RoundId:          uint64(roundId),
			SignerPublicKey:  utils.GetBtcPublicKey(),
			SweepSignature:   sweepSignatures,
			BtcOracleAddress: oracleAddr,
		}

		comms.SendTransactionSignSweep(accountName, cosmos, msg)

		db.MarkAddressSignedSweep(dbconn, reserveAddress.Address)
		db.InsertTransaction(dbconn, sweepTx.TxHash().String(), reserveAddress.Address, uint64(reserveId), uint64(roundId))
	}

	fmt.Println("finishing sweep tx signer")
}

func ProcessTxSigningRefund(accountName string, dbconn *sql.DB, oracleAddr string) {
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
		refundSignature := utils.SignTx(refundTx, reserveAddress.Script)

		reserveId, _ := strconv.Atoi(tx.ReserveId)
		roundId, _ := strconv.Atoi(tx.RoundId)

		fmt.Println("Refund Signature : ", refundSignature)
		cosmos := comms.GetCosmosClient()
		msg := &bridgetypes.MsgSignRefund{
			ReserveId:        uint64(reserveId),
			RoundId:          uint64(roundId),
			SignerPublicKey:  utils.GetBtcPublicKey(),
			RefundSignature:  []string{refundSignature[0]},
			BtcOracleAddress: oracleAddr,
		}

		comms.SendTransactionSignRefund(accountName, cosmos, msg)

		db.MarkAddressSignedRefund(dbconn, reserveAddress.Address)
	}

	fmt.Println("finishing refund tx signer")
}
