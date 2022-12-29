package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/ignite/cli/ignite/pkg/cosmosclient"
	"github.com/twilight-project/nyks/x/bridge/types"
)

func generateAddress() string {
	address := "bc1q0vsja3m9xtskmxwrjmkz2phgd8n95k9f8gcy4s"
	return address
}

func registerReserveAddressOnNyks(accountName string, address string) {

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
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
		fmt.Println(err)
	}

	judge_address, err := cosmos.Address(accountName)
	if err != nil {
		log.Fatal(err)
	}

	msg := &types.MsgRegisterReserveAddress{
		ReserveScript: address,
		JudgeAddress:  judge_address.String(),
	}

	// store response in txResp
	txResp, err := cosmos.BroadcastTx(accountName, msg)
	if err != nil {
		fmt.Println(err)
	}

	// print response from broadcasting a transaction
	fmt.Print("MsgRegisterReserveAddress:\n\n")
	fmt.Println(txResp)
}

func registerAddressOnForkscanner(address string) {
	dt := time.Now().UTC()
	dt = dt.AddDate(1, 0, 0)

	request_body := map[string]interface{}{
		"method":  "add_watched_addresses",
		"id":      1,
		"jsonrpc": "2.0",
		"params": map[string]interface{}{
			"add": []interface{}{
				map[string]string{
					"address":     address,
					"watch_until": dt.Format(time.RFC3339),
				},
			},
		},
	}

	data, err := json.Marshal(request_body)
	if err != nil {
		log.Fatalf("Post: %v", err)
	}
	fmt.Println(string(data))

	resp, err := http.Post("http://0.0.0.0:8339", "application/json", strings.NewReader(string(data)))
	if err != nil {
		log.Fatalf("Post: %v", err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("ReadAll: %v", err)
	}
	result := make(map[string]interface{})
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	log.Println(result)

}

func startJudge(accountName string) {
	address := generateAddress()
	registerReserveAddressOnNyks(accountName, address)
	registerAddressOnForkscanner(address)
}
