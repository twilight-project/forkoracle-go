package main

// contains all code communicating with nyksd chain (cosmos)
import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/ignite/cli/ignite/pkg/cosmosclient"
	"github.com/spf13/viper"
	bridgetypes "github.com/twilight-project/nyks/x/bridge/types"
	forktypes "github.com/twilight-project/nyks/x/forks/types"
)

func sendTransactionSeenBtcChainTip(accountName string, cosmos cosmosclient.Client, data *forktypes.MsgSeenBtcChainTip) {
	_, err := cosmos.BroadcastTx(accountName, data)
	if err != nil {
		fmt.Println("error in chaintip trnasaction", err)
	}
	fmt.Println("seen Chaintip transaction")
}

func sendTransactionConfirmBtcdeposit(accountName string, cosmos cosmosclient.Client, data *bridgetypes.MsgConfirmBtcDeposit) {

	_, err := cosmos.BroadcastTx(accountName, data)
	if err != nil {
		fmt.Println("error in confirm deposit transaction : ", err)
	}
}

func getCosmosClient() cosmosclient.Client {
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

	return cosmos
}

func getCosmosAddress(accountName string, cosmos cosmosclient.Client) sdktypes.AccAddress {
	address, err := cosmos.Address(accountName)
	if err != nil {
		log.Fatal(err)
	}
	return address
}

func getDepositAddresses() QueryDepositAddressResp {
	nyksd_url := fmt.Sprintf("%v", viper.Get("nyksd_url"))
	resp, err := http.Get(nyksd_url + "/twilight-project/nyks/bridge/registered_btc_deposit_addresses")
	if err != nil {
		fmt.Println("error getting deposit addresses : ", err)
	}
	//We Read the response body on the line below.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error getting deposit addresses body : ", err)
	}

	a := QueryDepositAddressResp{}
	err = json.Unmarshal(body, &a)
	if err != nil {
		fmt.Println("error unmarshalling deposit addresses : ", err)
	}
	return a
}

func getDepositAddress(address string) DepositAddress {
	nyksd_url := fmt.Sprintf("%v", viper.Get("nyksd_url"))
	resp, err := http.Get(nyksd_url + "/twilight-project/nyks/bridge/registered_btc_deposit_address/" + address)
	if err != nil {
		fmt.Println("error getting deposit addresses : ", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error getting deposit addresses body : ", err)
	}
	a := DepositAddress{}
	err = json.Unmarshal(body, &a)
	if err != nil {
		fmt.Println("error unmarshalling deposit addresses : ", err)
	}
	return a
}

func getAttestations() AttestaionBlock {
	nyksd_url := fmt.Sprintf("%v", viper.Get("nyksd_url"))
	resp, err := http.Get(nyksd_url + "/twilight-project/nyks/nyks/attestations?limit=1&order_by=desc")
	if err != nil {
		log.Fatalln(err)
	}
	//We Read the response body on the line below.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	a := AttestaionBlock{}
	err = json.Unmarshal(body, &a)

	return a
}
