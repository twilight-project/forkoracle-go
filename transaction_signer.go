package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"time"
)

func processSweepTx(accountName string) {

	for {
		SweepProposal := getAttestationsSweepProposal()

		for _, attestation := range SweepProposal.Attestations {
			sweeptxHex := attestation.Proposal.BtcSweepTx
			refundtxHex := attestation.Proposal.BtcRefundTx
			result := strings.Split(refundtxHex, "++")
			refundtxHex = result[0]
			newReserveStr := result[1]

			reserveAddress := attestation.Proposal.ReserveAddress
			addrs := querySweepAddress(reserveAddress)
			if len(addrs) <= 0 {
				continue
			}

			currentReserveAddr := addrs[0]

			if currentReserveAddr.Signed == false {
				sweeptx, err := createTxFromHex(sweeptxHex)
				if err != nil {
					fmt.Println("error decoding sweep tx : inside processSweepTx : ", err)
					log.Fatal(err)
				}

				refundtx, err := createTxFromHex(refundtxHex)
				if err != nil {
					fmt.Println("error decoding sweep tx : inside processSweepTx : ", err)
					log.Fatal(err)
				}

				addrs := querySweepAddress(newReserveStr)
				if len(addrs) <= 0 {
					continue
				}

				newReserveAddress := addrs[0]

				sweepSignature := signTx(sweeptx, currentReserveAddr.Script)
				refundSignature := signTx(refundtx, newReserveAddress.Script)

				hexSweepSignature := hex.EncodeToString(sweepSignature)
				fmt.Println("Sweep Signature : ", hexSweepSignature)
				sendSweepSign(hexSweepSignature, reserveAddress, accountName)

				hexRefundSignature := hex.EncodeToString(refundSignature)
				fmt.Println("Refund Signature : ", hexRefundSignature)
				sendSweepSign(hexSweepSignature, currentReserveAddr.Address, accountName)
				sendRefundSign(hexRefundSignature, newReserveAddress.Address, accountName)

				markSweepAddressSigned(reserveAddress)
			}
		}
		time.Sleep(1 * time.Minute)
	}
}

func startTransactionSigner(accountName string) {
	fmt.Println("starting Transaction Signer")
	processSweepTx(accountName)
}
