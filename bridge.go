package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
	"github.com/twilight-project/nyks/x/bridge/types"
)

func watchAddress(url url.URL) {
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
		log.Println("error in address watcher: ", err)
		return
	}

	fmt.Println("registered on address watcher")

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("error in address watcher: ", err)
			return
		}

		fmt.Printf("recv watchtower noti: %s", message)

		c := WatchtowerResponse{}
		err = json.Unmarshal(message, &c)
		if err != nil {
			fmt.Println("error in address watcher: ", err)
			continue
		}

		if len(c.Params) <= 0 {
			fmt.Println("first message from forkscanner, confirming subscription")
			continue
		}

		watchtower_notifications := c.Params
		for _, notification := range watchtower_notifications {
			insertNotifications(notification)
		}

	}

}

func kDeepService(accountName string) {
	fmt.Println("running k deep service")
	for {
		resp := getAttestations("5")
		if len(resp.Attestations) > 0 {
			fmt.Println("INFO : k deep : latest attestaion is : ", resp.Attestations[0])
			for _, attestation := range resp.Attestations {
				if attestation.Observed == true {
					fmt.Println("attestaion is Observed : ", attestation.Proposal.Height)
					height, err := strconv.ParseUint(attestation.Proposal.Height, 10, 64)
					if err != nil {
						fmt.Println(err)
					}
					kDeepCheck(accountName, uint64(height))
					break
				}
			}

		}
		time.Sleep(5 * time.Minute)
	}
}

func kDeepCheck(accountName string, height uint64) {
	fmt.Println("running k deep check for height : ", height)
	addresses := queryNotification()
	number := fmt.Sprintf("%v", viper.Get("confirmation_limit"))
	confirmations, _ := strconv.ParseUint(number, 10, 64)
	for _, a := range addresses {
		if height-a.Height >= confirmations {
			fmt.Println("reached height confirmations: ", height)
			confirmBtcTransactionOnNyks(accountName, a)
		}
	}
	fmt.Println("finishing k deep check for height : ", height)
}

func confirmBtcTransactionOnNyks(accountName string, data WatchtowerNotification) {
	fmt.Println("inside confirm btc transaction")
	cosmos := getCosmosClient()

	deposit_address := getDepositAddress(data.Sending)

	if deposit_address.DepositAddress != data.Sending {
		fmt.Println("addresses don't match: ", deposit_address.DepositAddress, " : ", data.Sending)
		markProcessedNotifications(data)
		return
	}

	fmt.Println("Data for confirm BTC deposut : ", data)

	msg := &types.MsgConfirmBtcDeposit{
		ReserveAddress:         data.Receiving,
		DepositAmount:          data.Satoshis,
		Height:                 data.Height,
		Hash:                   data.Receiving_txid,
		TwilightDepositAddress: deposit_address.TwilightDepositAddress,
		OracleAddress:          oracleAddr,
	}
	fmt.Println("confirming btc transaction")
	sendTransactionConfirmBtcdeposit(accountName, cosmos, msg)
	fmt.Println("deleting notifiction after procesing")
	markProcessedNotifications(data)

}

func updatereserveaddresses() {
	for {
		time.Sleep(2 * time.Minute)
		registerAddressOnValidators()
	}
}

func startBridge(accountName string, forkscanner_url url.URL) {
	fmt.Println("starting bridge")
	if judge == false {
		go updatereserveaddresses()
	}
	go watchAddress(forkscanner_url)
	kDeepService(accountName)
}
