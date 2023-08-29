package main

import (
	"fmt"
	"log"
	"time"
)

func processTxSigning(accountName string) {
	for {
		SweepTxs := getUnsignedSweepTx()
		refundTxs := getUnsignedRefundTx()

		for _, tx := range SweepTxs.UnsignedTxSweepMsgs {
			sweepTx, err := createTxFromHex(tx.BtcUnsignedSweepTx)
			if err != nil {
				fmt.Println("error decoding sweep tx : inside processSweepTx : ", err)
				log.Fatal(err)
			}
			addresses := queryUnsignedSweepAddressByScript(sweepTx.TxIn[0].Witness[0])
			if len(addresses) <= 0 {
				continue
			}
			reserveAddress := addresses[0]

			if reserveAddress.Signed_sweep == true {
				continue
			}
			hexSweepSignature := signTx(sweepTx, reserveAddress.Script)

			fmt.Println("Sweep Signature : ", hexSweepSignature)
			sendSweepSign(hexSweepSignature, reserveAddress.Address, accountName)

			markAddressSignedSweep(reserveAddress.Address)
			if judge == false {
				markAddressArchived(reserveAddress.Address)
			}

			insertTransaction(sweepTx.TxHash().String(), reserveAddress.Address, 0)
		}

		for _, tx := range refundTxs.UnsignedTxRefundMsgs {
			refundTx, err := createTxFromHex(tx.BtcUnsignedRefundTx)
			if err != nil {
				fmt.Println("error decoding sweep tx : inside processSweepTx : ", err)
				log.Fatal(err)
			}
			addresses := queryUnsignedRefundAddressByScript(refundTx.TxIn[0].Witness[0])
			if len(addresses) <= 0 {
				continue
			}
			reserveAddress := addresses[0]

			if reserveAddress.Signed_refund == true {
				continue
			}
			hexRefundSignature := signTx(refundTx, reserveAddress.Script)

			fmt.Println("Refund Signature : ", hexRefundSignature)
			sendRefundSign(hexRefundSignature, reserveAddress.Address, accountName)
			markAddressSignedRefund(reserveAddress.Address)
		}

		time.Sleep(1 * time.Minute)
	}
}

func startTransactionSigner(accountName string) {
	fmt.Println("starting Transaction Signer")
	processTxSigning(accountName)
}
