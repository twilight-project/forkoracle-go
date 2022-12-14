package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/websocket"
)

func watch_address(url url.URL, db *sql.DB) {
	conn, _, err := websocket.DefaultDialer.Dial(url.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	defer conn.Close()

	payload := `{
		"jsonrpc": "2.0",
		"id": "watched_address_checks",
		"method": "watched_address_checks",
		"params": {
			"watch": [],
			"watch_until": "2999-09-30T00:00:00.0Z"
		}
	}`

	err = conn.WriteMessage(websocket.TextMessage, []byte(payload))
	if err != nil {
		log.Println("write:", err)
		return
	}

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		//save in DB
		log.Printf("recv: %s", message)

		c := WatchtowerResponse{}
		err = json.Unmarshal(message, &c)
		if err != nil {
			fmt.Println(err)
		}

		watch_notification := c.Params

		resp, err := http.Get("https://nyks.twilight-explorer.com/api/twilight-project/nyks/bridge/registered_btc_deposit_addresses")
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
		for _, address := range a.addresses {
			for _, element := range watch_notification {
				if address.depositAddress == element.Receiving {
					_, err = db.Exec("INSERT into watched VALUES ($1, $2, $3, $4, $5)",
						element.Block,
						element.Receiving,
						element.Satoshis,
						element.Height,
						element.Txid,
					)
					if err != nil {
						log.Fatalf("An error occured while executing query: %v", err)
					}
				}
			}
		}

	}

}

func k_deep_service(accountName string, url url.URL, db *sql.DB) {

	for {
		resp, err := http.Get("https://nyks.twilight-explorer.com/api/twilight-project/nyks/nyks/attestations?limit=1&order_by=desc")
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

		if a.Attestations != nil {
			attestation := a.Attestations[0]

			if attestation.Observed == true {
				height, err := strconv.ParseUint(attestation.Proposal.Height, 10, 64)
				if err != nil {
					fmt.Println(err)
				}
				k_deep_check(accountName, uint64(height), db)
			}

		}

	}
}

func k_deep_check(accountName string, height uint64, db *sql.DB) {
	DB_reader, err := db.Query("select * from watched where archived = false")
	if err != nil {
		log.Fatalf("An error occured while executing query: %v", err)
	}
	defer DB_reader.Close()

	addresses := make([]WatchtowerNotification, 0)

	for DB_reader.Next() {
		address := WatchtowerNotification{}
		err := DB_reader.Scan(
			address.Block,
			address.Receiving,
			address.Satoshis,
			address.Height,
			address.Txid,
			address.archived,
		)

		if err != nil {
			fmt.Println(err)
		}
		addresses = append(addresses, address)
	}

	for _, a := range addresses {
		if height-a.Height > 3 {
			Confirm_BTc_Transaction_on_nyks(accountName, a)
		}
	}
}

func Confirm_BTc_Transaction_on_nyks(accountName string, data WatchtowerNotification) {

	// home, err := os.UserHomeDir()
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// homePath := filepath.Join(home, ".nyks")

	// cosmosOptions := []cosmosclient.Option{
	// 	cosmosclient.WithHome(homePath),
	// }

	// config := sdktypes.GetConfig()
	// config.SetBech32PrefixForAccount("twilight", "twilight"+"pub")

	// // create an instance of cosmosclient
	// cosmos, err := cosmosclient.New(context.Background(), cosmosOptions...)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// oracle_address, err := cosmos.Address(accountName)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// msg := &types.MsgConfirmBtcDeposit{
	// 	DepositAddress:         data.Receiving,
	// 	DepositAmount:          data.Satoshis,
	// 	BlockHeight:            data.Height,
	// 	BlockHash:              data.Txid,
	// 	TwilightDepositAddress: "",
	// 	BtcOracleAddress:       oracle_address.String(),
	// 	InputAddress:           data.Receiving,
	// }

}

func start_bridge(accountName string, forkscanner_url url.URL, db *sql.DB) {

	go watch_address(forkscanner_url, db)
	go k_deep_service(accountName, forkscanner_url, db)

	// go watch_address(address)

}
