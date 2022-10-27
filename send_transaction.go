package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/ignite/cli/ignite/pkg/cosmosclient"
	"github.com/twilight-project/nyks/x/nyks/types"
)

type ChainTip struct {
	Block           string `json:"block"`
	Height          int64  `json:"height"`
	ID              int64  `json:"id"`
	Node            int64  `json:"node"`
	Parent_chaintip *int64 `json:"parent_chaintip,omitempty"`
	Status          string `json:"status"`
}

type BlockData struct {
	Method   string     `json:"method"`
	ChainTip []ChainTip `json:"params,omitempty"`
}

func send_transaction(accountName string, c ChainTip) {

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	homePath := filepath.Join(home, ".nyks")

	cosmosOptions := []cosmosclient.Option{
		cosmosclient.WithHome(homePath),
	}

	config := sdktypes.GetConfig()
	config.SetBech32PrefixForAccount("twilight", "twilight"+"pub")

	// create an instance of cosmosclient
	cosmos, err := cosmosclient.New(context.Background(), cosmosOptions...)
	if err != nil {
		log.Fatal(err)
	}

	// get account from the keyring by account name and return a bech32 address
	address, err := cosmos.Address(accountName)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(address.String())
	// define a message to create a post
	msg := &types.MsgSeenBtcChainTip{
		Creator:             address.String(),
		Height:              uint64(c.Height),
		Hash:                c.Block,
		OrchestratorAddress: address.String(),
	}

	// broadcast a transaction from account `alice` with the message to create a post
	// store response in txResp
	txResp, err := cosmos.BroadcastTx(accountName, msg)
	if err != nil {
		log.Println(err)
	}

	// print response from broadcasting a transaction
	fmt.Print("MsgSeenBtcChainTip:\n\n")
	fmt.Println(txResp)

}
