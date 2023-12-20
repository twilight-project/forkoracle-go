package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/bech32"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/viper"
	bridgetypes "github.com/twilight-project/nyks/x/bridge/types"
)

var txids []string

func TestDepositAddress(t *testing.T) {
	txids = append(txids, "8fe487104de3725d07ba93dafc300d5351c01893ec909a22ed19aad8061c8477")
	txids = append(txids, "8797380dd4658eb25e77954939cde0a880e659a6005d3a053d24835600d2dd75")
	txids = append(txids, "8fe487104de3725d07ba93dafc300d5351c01893ec909a22ed19aad8061c8474")
	txids = append(txids, "8fe487104de3725d07ba93dafc300d5351c01893ec909a22ed19aad8061c8473")
	txids = append(txids, "8fe487104de3725d07ba93dafc300d5351c01893ec909a22ed19aad8061c8472")

	initialize()
	accountName := fmt.Sprintf("%v", viper.Get("accountName"))
	time.Sleep(3 * time.Second)
	registerJudge(accountName)

	resevreAddresses := tregisterReserveAddress()
	depositAddresses, _ := tgenerateBitcoinAddresses(10000)
	twilightAddress, _ := tgenerateTwilightAddresses(10000)

	for _, taddr := range twilightAddress {
		command := fmt.Sprintf("nyksd tx bank send $(nyksd keys show validator-sfo -a --keyring-backend test) %s 20000nyks --keyring-backend test", taddr)
		args := strings.Fields(command)
		cmd := exec.Command(args[0], args[1:]...)
		_, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			return
		}
		time.Sleep(3 * time.Second)
	}

	tregisterDepositAddress(10000, depositAddresses, twilightAddress)
	tconfirmBtcTransaction(10000, depositAddresses, resevreAddresses)
	twithdrawalBtc(10000, depositAddresses, resevreAddresses)

	fmt.Println("Press 'Enter' to continue...")

	// Create a new scanner reading from standard input
	scanner := bufio.NewScanner(os.Stdin)

	// Wait for input
	scanner.Scan()

	for i, rAddr := range resevreAddresses {
		newSweepAddress, script := generateAddress(10000, rAddr)
		cosmos_client := getCosmosClient()
		msg := &bridgetypes.MsgProposeSweepAddress{
			BtcScript:    hex.EncodeToString(script),
			BtcAddress:   newSweepAddress,
			JudgeAddress: oracleAddr,
			ReserveId:    uint64(i),
			RoundId:      uint64(2),
		}
		sendTransactionSweepAddressProposal(accountName, cosmos_client, msg)
	}

}

func tregisterReserveAddress() []string {
	addresses := make([]string, 25)
	accountName := fmt.Sprintf("%v", viper.Get("accountName"))
	for i := 0; i < 25; i++ {
		addresses[i] = generateAndRegisterNewBtcReserveAddress(accountName, 100)
		time.Sleep(3 * time.Second)
	}
	return addresses
}

func tconfirmBtcTransaction(n int, depositAddresses []string, reserveAddresses []string) {
	accountName := fmt.Sprintf("%v", viper.Get("accountName"))
	for i := 0; i < n; i++ {
		tx := WatchtowerNotification{
			Block:            "00000000000000000003239eae998dc7ad3585c2a08a3afc94d5a2721d1a2608",
			Height:           1000,
			Receiving:        depositAddresses[i],
			Satoshis:         50000,
			Receiving_txid:   txids[i%5],
			Sending_txinputs: []WatchtowerTxInput{},
			Archived:         false,
			Receiving_vout:   uint64(i),
			Sending:          reserveAddresses[i%25],
			Sending_vout:     -1,
		}
		confirmBtcTransactionOnNyks(accountName, tx)
		time.Sleep(3 * time.Second)
	}
}

func tgenerateBitcoinAddresses(n int) ([]string, error) {
	addresses := make([]string, n)
	for i := 0; i < n; i++ {
		// Derive a new public key (non-standard approach)
		privateKey, err := btcec.NewPrivateKey(btcec.S256())
		if err != nil {
			return nil, err
		}
		pubKeyHash := btcutil.Hash160(privateKey.PubKey().SerializeCompressed())

		// Convert the public key hash to a bech32 encoded address
		x, _ := bech32.ConvertBits(pubKeyHash, 8, 5, true)
		segwitAddr, _ := bech32.Encode("bc", x)

		addresses[i] = segwitAddr
	}
	return addresses, nil
}

func tgenerateTwilightAddresses(n int) ([]string, error) {
	addresses := make([]string, n)
	for i := 0; i < n; i++ {
		customPrefix := "twilight"
		config := types.GetConfig()
		config.SetBech32PrefixForAccount(customPrefix, customPrefix+"pub")

		// Generate a new private key
		privateKey := secp256k1.GenPrivKey()
		publicKey := privateKey.PubKey()

		// Convert the public key to an address
		address := types.AccAddress(publicKey.Address())

		addresses[i] = address.String()
	}
	addresses[0] = "twilight1qskpa0sgd56nzuhlq6rf098quxx05quln22l9e"
	return addresses, nil
}

func tregisterDepositAddress(n int, btcAddresses []string, twilightAddresses []string) {
	cosmos := getCosmosClient()
	accountName := fmt.Sprintf("%v", viper.Get("accountName"))

	for i, addr := range btcAddresses {
		msg := &bridgetypes.MsgRegisterBtcDepositAddress{
			BtcDepositAddress:     addr,
			BtcSatoshiTestAmount:  50000,
			TwilightStakingAmount: 10000,
			TwilightAddress:       twilightAddresses[i],
		}
		_, err := cosmos.BroadcastTx(accountName, msg)
		if err != nil {
			fmt.Println("error in registering deposit address : ", err)
		}
		time.Sleep(3 * time.Second)
	}
}

func twithdrawalBtc(n int, btcAddresses []string, twilightAddresses []string) {
	cosmos := getCosmosClient()
	accountName := fmt.Sprintf("%v", viper.Get("accountName"))

	for i, addr := range btcAddresses {
		msg := &bridgetypes.MsgWithdrawBtcRequest{
			WithdrawAddress: addr,
			ReserveId:       uint64(i % 25),
			WithdrawAmount:  30000,
			TwilightAddress: twilightAddresses[i],
		}
		_, err := cosmos.BroadcastTx(accountName, msg)
		if err != nil {
			fmt.Println("error in registering deposit address : ", err)
		}
		time.Sleep(3 * time.Second)
	}
}

// func tRegisterJudgeTest(accountName string) {
// 	cosmos := getCosmosClient()
// 	msg := &bridgetypes.MsgRegisterJudge{
// 		Creator:          oracleAddr,
// 		JudgeAddress:     oracleAddr,
// 		ValidatorAddress: valAddr,
// 	}

// 	sendTransactionRegisterJudge(accountName, cosmos, msg)
// 	fmt.Println("registered Judge")
// }
