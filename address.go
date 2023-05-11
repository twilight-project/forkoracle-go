package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/cosmos/btcutil"
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

func BuildScript(preimage []byte) ([]byte, error) {

	delegateAddresses := getDelegateAddresses()
	payment_hash := hash160(preimage)

	builder := txscript.NewScriptBuilder()
	builder.AddOp(txscript.OP_SIZE)
	builder.AddInt64(32)
	builder.AddOp(txscript.OP_EQUALVERIFY)
	builder.AddOp(txscript.OP_HASH160)
	builder.AddData(payment_hash)
	builder.AddOp(txscript.OP_EQUALVERIFY)

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
	}
	builder.AddInt64(int64(len(delegateAddresses.Addresses)))
	builder.AddOp(txscript.OP_CHECKMULTISIG)

	redeemScript, err := builder.Script()
	if err != nil {
		fmt.Println(err)
	}
	return redeemScript, nil
}

func BuildWitnessScript(redeemScript []byte) []byte {
	WitnessScript := sha256.Sum256(redeemScript)
	return WitnessScript[:]
}

func generateAddress(unlock_height int) (string, []byte) {
	preimage, err := preimage()
	if err != nil {
		fmt.Println(err)
	}
	redeemScript, err := BuildScript(preimage)
	if err != nil {
		fmt.Println(err)
	}
	WitnessScript := BuildWitnessScript(redeemScript)
	address, err := btcutil.NewAddressWitnessScriptHash(WitnessScript, &chaincfg.MainNetParams)
	if err != nil {
		fmt.Println("Error:", err)
		return "", nil
	}

	addressStr := address.String()
	fmt.Println("new address generated : ", addressStr)

	insertSweepAddress(addressStr, redeemScript, preimage, int64(unlock_height)+1000)

	return addressStr, redeemScript
}
