package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/twilight-project/nyks/x/bridge/types"
)

func watch_address(url url.URL) {
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
		resp := get_deposit_addresses()

		for _, address := range resp.addresses {
			for _, element := range watch_notification {
				if address.depositAddress == element.Receiving {
					insert_notifications(element)
				}
			}
		}

	}

}

func k_deep_service(accountName string, url url.URL) {

	for {
		resp := get_attestations()
		if resp.Attestations != nil {
			attestation := resp.Attestations[0]

			if attestation.Observed == true {
				height, err := strconv.ParseUint(attestation.Proposal.Height, 10, 64)
				if err != nil {
					fmt.Println(err)
				}
				k_deep_check(accountName, uint64(height))
			}

		}

	}
}

func k_deep_check(accountName string, height uint64) {
	addresses := query_notification()
	for _, a := range addresses {
		if height-a.Height > 3 {
			Confirm_BTc_Transaction_on_nyks(accountName, a)
		}
	}
}

func Confirm_BTc_Transaction_on_nyks(accountName string, data WatchtowerNotification) {

	cosmos := get_cosmos_client()
	oracle_address := get_cosmos_address(accountName, cosmos)

	deposit_addresses := get_deposit_addresses()

	for _, a := range deposit_addresses.addresses {
		if a.depositAddress == data.Receiving {
			msg := &types.MsgConfirmBtcDeposit{
				DepositAddress:         data.Receiving,
				DepositAmount:          data.Satoshis,
				Height:                 data.Height,
				Hash:                   data.Txid,
				TwilightDepositAddress: a.twilightDepositAddress,
				BtcOracleAddress:       oracle_address.String(),
			}

			send_transaction(accountName, cosmos, msg, "MsgConfirmBtcDeposit")
		}
	}

}

func start_bridge(accountName string, forkscanner_url url.URL) {

	go watch_address(forkscanner_url)
	go k_deep_service(accountName, forkscanner_url)

}
