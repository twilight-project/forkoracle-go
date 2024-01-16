package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ignite/cli/ignite/pkg/cosmosclient"
	"github.com/spf13/viper"
	bridgetypes "github.com/twilight-project/nyks/x/bridge/types"
)

var txids []string
var limit int
var round int
var secondsWait int

func TestDepositAddress(t *testing.T) {

	kr, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendTest, "/root/.nyks/", nil)
	if err != nil {
		log.Fatalf("failed to open keyring: %v", err)
	}

	limit = 10
	secondsWait = 3
	round = 1

	initialize()
	accountName := fmt.Sprintf("%v", viper.Get("accountName"))
	time.Sleep(time.Duration(secondsWait) * time.Second)
	registerJudge(accountName)
	cosmos := getCosmosClient()
	tregisterReserveAddress()

	for j := 1; j <= round; j++ {
		reserveAddresses := getBtcReserves().BtcReserves
		txids = generateRandomHex(64)
		depositAddresses, _ := tgenerateBitcoinAddresses()
		twilightAddress, _ := tgenerateTwilightAddresses(kr)

		taddFunds(twilightAddress, cosmos)

		tregisterDepositAddress(depositAddresses, twilightAddress, cosmos)
		tconfirmBtcTransaction(depositAddresses, reserveAddresses)
		twithdrawalBtc(depositAddresses, twilightAddress, cosmos)

		time.Sleep(1 * time.Minute)

		for i, addr := range reserveAddresses {
			newSweepAddress := tproposeAddress(addr.ReserveAddress, uint64(i+i), uint64(j))
			time.Sleep(1 * time.Minute)
			sweeptx := tsendUnsignedSweeptx(addr.ReserveAddress, newSweepAddress, uint64(i+i), uint64(j))
			refundtx := tsendUnsignedRefundtx(addr.ReserveAddress, sweeptx, uint64(i+i), uint64(j))
			tsendSignedRefundtx(addr.ReserveAddress, refundtx, uint64(i+i), uint64(j))
			tsendSignedSweeptx(addr.ReserveAddress, sweeptx, uint64(i+i), uint64(j))
			tsendSendSweepProposal(newSweepAddress, cosmos, uint64(i+i), uint64(j))
		}
	}
}

func generateRandomHex(n int) []string {
	var hexStrings []string
	for i := 0; i < limit; i++ {
		bytes := make([]byte, n/2)
		if _, err := rand.Read(bytes); err != nil {
			return nil
		}
		hexStrings = append(hexStrings, fmt.Sprintf("%x", bytes))
	}
	return hexStrings
}

func taddFunds(twilightAddress []string, cosmos cosmosclient.Client) {
	fmt.Println("adding funds")
	cosmos.Context().WithBroadcastMode(flags.BroadcastAsync)
	accountName := fmt.Sprintf("%v", viper.Get("accountName"))
	amount := sdk.NewCoins(sdk.NewCoin("nyks", sdk.NewInt(100000)))
	for _, taddr := range twilightAddress {
		msg := &banktypes.MsgSend{
			FromAddress: oracleAddr,
			ToAddress:   taddr,
			Amount:      amount,
		}
		cosmos.BroadcastTx(accountName, msg)
		time.Sleep(time.Duration(secondsWait) * time.Second)
		fmt.Println("sending funds : ", taddr)
	}
}

func tproposeAddress(resevreAddress string, reserve uint64, round uint64) string {
	accountName := fmt.Sprintf("%v", viper.Get("accountName"))
	newSweepAddress, script := generateAddress(1000000, resevreAddress)
	cosmos_client := getCosmosClient()
	msg := &bridgetypes.MsgProposeSweepAddress{
		BtcScript:    hex.EncodeToString(script),
		BtcAddress:   newSweepAddress,
		JudgeAddress: oracleAddr,
		ReserveId:    reserve,
		RoundId:      round,
	}
	sendTransactionSweepAddressProposal(accountName, cosmos_client, msg)
	fmt.Println("new proposed address: ", newSweepAddress)
	time.Sleep(time.Duration(secondsWait) * time.Second)
	return newSweepAddress

}

func tregisterReserveAddress() []string {
	addresses := make([]string, 25)
	accountName := fmt.Sprintf("%v", viper.Get("accountName"))
	for i := 0; i < 25; i++ {
		addresses[i] = generateAndRegisterNewBtcReserveAddress(accountName, 100)
		time.Sleep(time.Duration(secondsWait) * time.Second)
	}
	return addresses
}

func tconfirmBtcTransaction(depositAddresses []string, reserveAddresses []BtcReserve) {
	accountName := fmt.Sprintf("%v", viper.Get("accountName"))
	for i := 0; i < limit; i++ {
		tx := WatchtowerNotification{
			Block:            "00000000000000000003239eae998dc7ad3585c2a08a3afc94d5a2721d1a2608",
			Height:           1000,
			Receiving:        reserveAddresses[i%25].ReserveAddress,
			Satoshis:         50000,
			Receiving_txid:   txids[i],
			Sending_txinputs: []WatchtowerTxInput{},
			Archived:         false,
			Receiving_vout:   uint64(i),
			Sending:          depositAddresses[i],
			Sending_vout:     -1,
		}
		confirmBtcTransactionOnNyks(accountName, tx)
		time.Sleep(time.Duration(secondsWait) * time.Second)
	}
}

func tgenerateBitcoinAddresses() ([]string, error) {
	addresses := make([]string, limit)
	for i := 0; i < limit; i++ {
		// Derive a new private key
		privateKey, err := btcec.NewPrivateKey(btcec.S256())
		if err != nil {
			log.Fatal(err)
		}

		// Generate the public key from the private key.
		pubKey := privateKey.PubKey()

		// Serialize the compressed public key.
		serializedPubKey := pubKey.SerializeCompressed()

		// Generate P2WPKH (Pay to Witness Public Key Hash) script.
		witnessProgram := btcutil.Hash160(serializedPubKey)
		address, err := btcutil.NewAddressWitnessPubKeyHash(witnessProgram, &chaincfg.MainNetParams)
		if err != nil {
			log.Fatal(err)
		}

		// Generate the bech32 encoded SegWit address.
		segwitAddress := address.EncodeAddress()
		addresses[i] = segwitAddress
		fmt.Println("SegWit Address:", segwitAddress)
	}
	return addresses, nil
}

func tgenerateTwilightAddresses(kr keyring.Keyring) ([]string, error) {
	addresses := make([]string, limit)
	customPrefix := "twilight"
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(customPrefix, customPrefix+"pub")

	for i := 0; i < limit; i++ {
		name := "AccountName" + fmt.Sprint(i)
		info, _, err := kr.NewMnemonic(name, keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		addresses[i] = info.GetAddress().String()
		fmt.Println(name, "  :  ", info.GetAddress().String())
	}

	return addresses, nil
}

func tregisterDepositAddress(btcAddresses []string, twilightAddresses []string, cosmos cosmosclient.Client) {
	fmt.Println("registering deposit address ")
	for i, addr := range btcAddresses {
		accountName := fmt.Sprintf("AccountName%d", i)

		msg := &bridgetypes.MsgRegisterBtcDepositAddress{
			BtcDepositAddress:     addr,
			BtcSatoshiTestAmount:  50000,
			TwilightStakingAmount: 50000,
			TwilightAddress:       twilightAddresses[i],
		}

		_, err := cosmos.BroadcastTx(accountName, msg)
		if err != nil {
			fmt.Println("error in registering deposit address : ", err)
		}
		time.Sleep(time.Duration(secondsWait) * time.Second)
	}
}

func twithdrawalBtc(btcAddresses []string, twilightAddress []string, cosmos cosmosclient.Client) {
	fmt.Println("creating withdraw requests")
	for i, addr := range btcAddresses {
		accountName := fmt.Sprintf("AccountName%d", i)
		msg := &bridgetypes.MsgWithdrawBtcRequest{
			WithdrawAddress: addr,
			ReserveId:       uint64((i + 1) % 25),
			WithdrawAmount:  30000,
			TwilightAddress: twilightAddress[i],
		}
		_, err := cosmos.BroadcastTx(accountName, msg)
		if err != nil {
			fmt.Println("error in registering withdraw address : ", err)
		}
		time.Sleep(time.Duration(secondsWait) * time.Second)
	}
}

func tsendUnsignedSweeptx(reserveAddress string, pAddress string, reserve uint64, round uint64) string {
	accountName := fmt.Sprintf("%v", viper.Get("accountName"))
	utxos := []Utxo{
		{Txid: "4f3c9b8f82f611e38f068342e37d6f083d74e64b2ccf7e8b4aee217aebad8fb4", Vout: 0, Amount: 1000000},
		{Txid: "7d6f083d74e64b2ccf7e8b4aee217aebad8fb44f3c9b8f82f611e38f068342e3", Vout: 1, Amount: 1000000},
		{Txid: "e38f068342e37d6f083d74e64b2ccf7e8b4aee217aebad8fb44f3c9b8f82f611", Vout: 2, Amount: 1000000},
		// Add more Utxo structs here as needed
	}
	withdrawRequests := getWithdrawSnapshot(reserve, round).WithdrawRequests
	fmt.Println(withdrawRequests)
	sweepTxHex, sweepTxId, _, _ := generateSweepTx(reserveAddress, *&pAddress, accountName, withdrawRequests, int64(1000), utxos)
	sendUnsignedSweepTx(reserve, round, sweepTxHex, sweepTxId, accountName)
	time.Sleep(time.Duration(secondsWait) * time.Second)

	return sweepTxHex
}

func tsendSignedSweeptx(reserveAddress string, sweeptx string, reserve uint64, round uint64) {
	accountName := fmt.Sprintf("%v", viper.Get("accountName"))
	broadcastSweeptxNYKS(sweeptx, accountName, reserve, round)
	time.Sleep(time.Duration(secondsWait) * time.Second)
}

func tsendSendSweepProposal(pAddress string, cosmos cosmosclient.Client, reserve uint64, round uint64) {
	accountName := fmt.Sprintf("%v", viper.Get("accountName"))
	msg := &bridgetypes.MsgSweepProposal{
		ReserveId:             reserve,
		NewReserveAddress:     pAddress,
		JudgeAddress:          oracleAddr,
		BtcRelayCapacityValue: 0,
		BtcTxHash:             "4f3c9b8f82f611e38f068342e37d6f083d74e64b2ccf7e8b4aee217aebad8fb4",
		UnlockHeight:          0,
		RoundId:               round,
		BtcBlockNumber:        0,
	}
	sendTransactionSweepProposal(accountName, cosmos, msg)
	time.Sleep(time.Duration(secondsWait) * time.Second)
}

func tsendSignedRefundtx(reserveAddress string, refundTx string, reserve uint64, round uint64) {
	accountName := fmt.Sprintf("%v", viper.Get("accountName"))
	broadcastRefundtxNYKS(refundTx, accountName, reserve, round)
	time.Sleep(time.Duration(secondsWait) * time.Second)
}

func tsendUnsignedRefundtx(reserveAddress string, sweeptx string, reserve uint64, round uint64) string {
	accountName := fmt.Sprintf("%v", viper.Get("accountName"))
	refundTxHex, _ := generateRefundTx(sweeptx, "", reserve, round)
	sendUnsignedRefundTx(refundTxHex, reserve, round, accountName)
	time.Sleep(time.Duration(secondsWait) * time.Second)
	return refundTxHex
}
