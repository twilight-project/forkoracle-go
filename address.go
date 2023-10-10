package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/cosmos/btcutil"
	"github.com/spf13/viper"
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

func buildScript(preimage []byte, unlockHeight int64) ([]byte, error) {
	var judgeBtcPK *btcec.PublicKey
	var refundJudgeAddress string
	judges := getRegisteredJudges()
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
	delegateAddresses := getDelegateAddresses()
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
	builder.AddOp(txscript.OP_DROP)

	// adding preimage check if multisig passes
	builder.AddOp(txscript.OP_SIZE)
	builder.AddInt64(32)
	builder.AddOp(txscript.OP_EQUALVERIFY)
	builder.AddOp(txscript.OP_DROP)
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

func generateAddress(unlock_height int64, oldReserveAddress string) (string, []byte) {
	preimage, err := preimage()
	if err != nil {
		fmt.Println(err)
	}
	redeemScript, err := buildScript(preimage, unlock_height)
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

	insertSweepAddress(addressStr, redeemScript, preimage, int64(unlock_height), oldReserveAddress)

	return addressStr, redeemScript
}
