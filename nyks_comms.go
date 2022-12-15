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

func send_transaction(accountName string, cosmos cosmosclient.Client, data interface{}, msgtype string) {

	switch msgtype {
	case "SeenBtcChainTip":
		msg, ok := data.(forktypes.MsgSeenBtcChainTip)
		if ok {
			_, err := cosmos.BroadcastTx(accountName, &msg)
			if err != nil {
				log.Println(err)
			}
		}
	case "MsgConfirmBtcDeposit":
		msg, ok := data.(bridgetypes.MsgConfirmBtcDeposit)
		if ok {
			_, err := cosmos.BroadcastTx(accountName, &msg)
			if err != nil {
				log.Println(err)
			}
		}

	default:
		panic("undefined message type")
	}

}

func get_cosmos_client() cosmosclient.Client {
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

func get_cosmos_address(accountName string, cosmos cosmosclient.Client) sdktypes.AccAddress {
	address, err := cosmos.Address(accountName)
	if err != nil {
		log.Fatal(err)
	}
	return address
}

func get_deposit_addresses() QueryDepositAddressResp {
	nyksd_url := fmt.Sprintf("%v", viper.Get("nyksd_url"))
	resp, err := http.Get(nyksd_url + "/api/twilight-project/nyks/bridge/registered_btc_deposit_addresses")
	if err != nil {
		log.Fatalln(err)
	}
	//We Read the response body on the line below.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	a := QueryDepositAddressResp{}
	err = json.Unmarshal(body, &a)
	if err != nil {
		fmt.Println(err)
	}

	return a
}

func get_attestations() AttestaionBlock {
	nyksd_url := fmt.Sprintf("%v", viper.Get("nyksd_url"))
	resp, err := http.Get(nyksd_url + "/api/twilight-project/nyks/nyks/attestations?limit=1&order_by=desc")
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
