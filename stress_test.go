package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/bech32"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ignite/cli/ignite/pkg/cosmosclient"
	"github.com/spf13/viper"
	bridgetypes "github.com/twilight-project/nyks/x/bridge/types"
)

var txids []string
var limit int
var secondsWait int

func TestDepositAddress(t *testing.T) {

	kr, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendTest, "/root/.nyks/keyring", nil)
	if err != nil {
		log.Fatalf("failed to open keyring: %v", err)
	}

	// txids = append(txids, "8fe487104de3725d07ba93dafc300d5351c01893ec909a22ed19aad8061c8477")
	// txids = append(txids, "8797380dd4658eb25e77954939cde0a880e659a6005d3a053d24835600d2dd75")
	// txids = append(txids, "8fe487104de3725d07ba93dafc300d5351c01893ec909a22ed19aad8061c8474")
	// txids = append(txids, "8fe487104de3725d07ba93dafc300d5351c01893ec909a22ed19aad8061c8473")
	// txids = append(txids, "8fe487104de3725d07ba93dafc300d5351c01893ec909a22ed19aad8061c8472")
	limit = 10
	// secondsWait = 3

	// initialize()
	// accountName := fmt.Sprintf("%v", viper.Get("accountName"))
	// time.Sleep(time.Duration(secondsWait) * time.Second)
	// registerJudge(accountName)

	// cosmos := getCosmosClient()

	// resevreAddresses := tregisterReserveAddress()
	// depositAddresses, _ := tgenerateBitcoinAddresses()
	_, _ = tgenerateTwilightAddresses(kr)

	// taddFunds(twilightAddress, cosmos)

	// tregisterDepositAddress(depositAddresses, twilightAddress, cosmos, kr)
	// tconfirmBtcTransaction(depositAddresses, resevreAddresses)
	// twithdrawalBtc(depositAddresses, resevreAddresses, cosmos)

	// fmt.Println("Press 'Enter' to continue...")
	// scanner := bufio.NewScanner(os.Stdin)
	// scanner.Scan()

	// newSweepAddresses := tproposeAddress(resevreAddresses)
	// sweeptxs := tsendUnsignedSweeptx(resevreAddresses, newSweepAddresses)
	// refundtxs := tsendUnsignedRefundtx(resevreAddresses, sweeptxs)
	// tsendSignedRefundtx(resevreAddresses, refundtxs)
	// tsendSignedSweeptx(resevreAddresses, sweeptxs)
	// tsendSendSweepProposal(newSweepAddresses, cosmos)
}

func taddFunds(twilightAddress []string, cosmos cosmosclient.Client) {
	accountName := fmt.Sprintf("%v", viper.Get("accountName"))
	amount := sdk.NewCoins(sdk.NewCoin("nyks", sdk.NewInt(20000)))
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

func tproposeAddress(resevreAddresses []string) []string {
	pAddresses := make([]string, 25)
	accountName := fmt.Sprintf("%v", viper.Get("accountName"))
	for i, rAddr := range resevreAddresses {
		newSweepAddress, script := generateAddress(int64(limit), rAddr)
		cosmos_client := getCosmosClient()
		msg := &bridgetypes.MsgProposeSweepAddress{
			BtcScript:    hex.EncodeToString(script),
			BtcAddress:   newSweepAddress,
			JudgeAddress: oracleAddr,
			ReserveId:    uint64(i),
			RoundId:      uint64(2),
		}
		sendTransactionSweepAddressProposal(accountName, cosmos_client, msg)
		time.Sleep(time.Duration(secondsWait) * time.Second)
		pAddresses[i] = newSweepAddress
		fmt.Println("new proposed address: ", newSweepAddress)
	}
	return pAddresses

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

func tconfirmBtcTransaction(depositAddresses []string, reserveAddresses []string) {
	accountName := fmt.Sprintf("%v", viper.Get("accountName"))
	for i := 0; i < limit; i++ {
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
		time.Sleep(time.Duration(secondsWait) * time.Second)
	}
}

func tgenerateBitcoinAddresses() ([]string, error) {
	addresses := make([]string, limit)
	for i := 0; i < limit; i++ {
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

func tgenerateTwilightAddresses(kr keyring.Keyring) ([]string, error) {
	addresses := make([]string, limit)
	customPrefix := "twilight"
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(customPrefix, customPrefix+"pub")

	for i := 0; i < limit; i++ {
		// Generate a new private key
		privateKey := secp256k1.GenPrivKey()
		name := "AccountName" + fmt.Sprint(i)
		info, err := kr.SavePubKey(name, privateKey.PubKey(), hd.Secp256k1Type)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		addresses[i] = info.GetAddress().String()
		fmt.Println(name, "  :  ", info.GetAddress().String())
	}

	return addresses, nil
}

func tregisterDepositAddress(btcAddresses []string, twilightAddresses []string, kr keyring.Keyring) {
	fmt.Println("registering deposit address ")
	cosmos := getCosmosClient()
	for i, addr := range btcAddresses {
		accountName := fmt.Sprintf("AccountName%d", i)
		fmt.Println(accountName, " : ", twilightAddresses[i])

		// info, err := kr.Key(accountName)
		// if err != nil {
		// 	fmt.Printf("failed to get key info for account %s: %v\n", accountName, err)
		// 	continue
		// }

		msg := &bridgetypes.MsgRegisterBtcDepositAddress{
			BtcDepositAddress:     addr,
			BtcSatoshiTestAmount:  50000,
			TwilightStakingAmount: 10000,
			TwilightAddress:       twilightAddresses[i],
		}

		// txf := tx.NewFactoryCLI(cosmos.Context().WithKeyring(kr), nil).
		// 	WithFees("1000nyks").
		// 	WithGas(200000)

		// txBuilder, _ := txf.BuildUnsignedTx(msg)

		// // Fetch account number and sequence
		// accNum, accSeq, err := authtypes.NewAccountRetriever(clientCtx).GetAccountNumberSequence(info.GetAddress())
		// if err != nil {
		// 	return err
		// }

		// // Sign the transaction
		// signerData := authtypes.SignerData{
		// 	AccountNumber: accNum,
		// 	Sequence:      accSeq,
		// 	ChainID:       clientCtx.ChainID,
		// }
		// err = tx.SignWithPrivKey(clientCtx.TxConfig.SignModeHandler().DefaultMode(), signerData, txBuilder, info.GetPrivKey(), clientCtx.TxConfig, accSeq)
		// return err

		_, err := cosmos.BroadcastTx(accountName, msg)
		if err != nil {
			fmt.Println("error in registering deposit address : ", err)
		}
		time.Sleep(time.Duration(secondsWait) * time.Second)
	}
}

func twithdrawalBtc(btcAddresses []string, twilightAddresses []string, cosmos cosmosclient.Client) {
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
		time.Sleep(time.Duration(secondsWait) * time.Second)
	}
}

func tsendUnsignedSweeptx(reserveAddresses []string, pAddresses []string) []string {
	txHexes := make([]string, 25)
	accountName := fmt.Sprintf("%v", viper.Get("accountName"))
	utxos := []Utxo{
		{Txid: "4f3c9b8f82f611e38f068342e37d6f083d74e64b2ccf7e8b4aee217aebad8fb4", Vout: 0, Amount: 1000000},
		{Txid: "7d6f083d74e64b2ccf7e8b4aee217aebad8fb44f3c9b8f82f611e38f068342e3", Vout: 1, Amount: 1000000},
		{Txid: "e38f068342e37d6f083d74e64b2ccf7e8b4aee217aebad8fb44f3c9b8f82f611", Vout: 2, Amount: 1000000},
		// Add more Utxo structs here as needed
	}

	for i, addr := range reserveAddresses {
		withdrawRequests := getWithdrawSnapshot(uint64(i), uint64(1)).WithdrawRequests
		sweepTxHex, sweepTxId, _, _ := generateSweepTx(addr, *&pAddresses[i], accountName, withdrawRequests, int64(1000), utxos)
		sendUnsignedSweepTx(uint64(i), uint64(2), sweepTxHex, sweepTxId, accountName)
		time.Sleep(time.Duration(secondsWait) * time.Second)
		txHexes[i] = sweepTxHex
	}
	return txHexes
}

func tsendSignedSweeptx(reserveAddresses []string, sweeptxs []string) {
	accountName := fmt.Sprintf("%v", viper.Get("accountName"))
	for i, _ := range reserveAddresses {
		broadcastSweeptxNYKS(sweeptxs[i], accountName, uint64(i), uint64(1))
		time.Sleep(time.Duration(secondsWait) * time.Second)
	}
}

func tsendSendSweepProposal(pAddress []string, cosmos cosmosclient.Client) {
	accountName := fmt.Sprintf("%v", viper.Get("accountName"))
	for i, addr := range pAddress {
		msg := &bridgetypes.MsgSweepProposal{
			ReserveId:             uint64(i),
			NewReserveAddress:     addr,
			JudgeAddress:          oracleAddr,
			BtcRelayCapacityValue: 0,
			BtcTxHash:             "4f3c9b8f82f611e38f068342e37d6f083d74e64b2ccf7e8b4aee217aebad8fb4",
			UnlockHeight:          0,
			RoundId:               uint64(1),
			BtcBlockNumber:        0,
		}
		sendTransactionSweepProposal(accountName, cosmos, msg)
		time.Sleep(time.Duration(secondsWait) * time.Second)
	}
}

func tsendSignedRefundtx(reserveAddresses []string, refundTx []string) {
	accountName := fmt.Sprintf("%v", viper.Get("accountName"))
	for i, _ := range reserveAddresses {
		broadcastRefundtxNYKS(refundTx[i], accountName, uint64(i), uint64(1))
		time.Sleep(time.Duration(secondsWait) * time.Second)
	}
}

func tsendUnsignedRefundtx(reserveAddresses []string, sweeptxs []string) []string {
	refundtxHexes := make([]string, 25)
	accountName := fmt.Sprintf("%v", viper.Get("accountName"))
	for i, _ := range reserveAddresses {
		refundTxHex, _ := generateRefundTx(sweeptxs[i], "", uint64(i), uint64(1))
		sendUnsignedRefundTx(refundTxHex, uint64(i), uint64(2), accountName)
		time.Sleep(time.Duration(secondsWait) * time.Second)
		refundtxHexes[i] = refundTxHex
	}
	return refundtxHexes
}
