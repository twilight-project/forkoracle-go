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
	bridgetypes "github.com/twilight-project/nyks/x/bridge/types"
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
	watchedTx := queryWatchedTransactions()
	number := fmt.Sprintf("%v", viper.Get("confirmation_limit"))
	confirmations, _ := strconv.ParseUint(number, 10, 64)
	for _, a := range addresses {
		if height-a.Height >= confirmations {
			fmt.Println("reached height confirmations: ", height)
			confirmBtcTransactionOnNyks(accountName, a)

			for _, tx := range watchedTx {
				if a.Receiving_txid == tx.Txid {
					cosmos := getCosmosClient()
					msg := &bridgetypes.MsgSweepProposal{
						ReserveId:             0,
						ReserveAddress:        tx.Address,
						JudgeAddress:          oracleAddr,
						BtcRelayCapacityValue: 0,
						TotalValue:            a.Satoshis,
						PrivatePoolValue:      0,
						PublicValue:           0,
						FeePool:               0,
						BtcRefundTx:           "010000000001019fc4e4a9cec5690c84f3773c383936a82c107167e2af7e4c7af48212fc390f880000000000f5ffffff021027000000000000160014a96c26be9d01a07078751ce66c46986db5e3c609a86100000000000022002010118cefe99eb61b69dc77601e7861b3dc80135e319538b96b401cdb2151a61f072038e542280948922ae2b9b64c39407b4581b862d8baa791d28b914e7f5e1fc9780100473044022045a0e3360112645a7e73abe8e09c8057e2ed09695f5ea6d905f9cb32627000ee02206a451d2cc07de37d1a8167522dcf1018fe259883c6f5ef12d2d1c5e6b07a303781483045022100de82176a9c45ba63bca9dbe65540d283acd0b8feda6a947179607191a13048e40220593b33dfbfc86dd9458e215ff7f5b3d5dea494ed7be98de2c1a515d0d49595a9814730440220735207b13f20376e0d87c3470e9b8c0c78c9cda912c60eff26e4e61505e5b3c002204519d13f71be59fa9c0871762bddf9e6b732ee37485a394d0aa04131cf50b67581483045022100b4eacf4b88f75634c3c31fef4a57ea014aa699b203a35224b96b5f62eb0e0e8e02202d7f1f052a74631d6f02185a6e98e69ab5d55101564f2626af16090a1e0584ce81f6532102ca505bf28698f0b6c26114a725f757b88d65537dd52a5b6455a9cac9581f105521038b38721dbb1427fd9c65654f87cb424517df717ee2fea8b0a5c376a1734941672103b03fe3da02ac2d43a1c2ebcfc7b0497e89cc9f62b513c0fc14f10d3d1a2cd5e62103bb3694e798f018a157f9e6dfb51b91f70a275443504393040892b52e45b255c32103e2f80f2f5eb646df3e0642ae137bf13f5a9a6af4c05688e147c64e8fae196fe155ae6382012088a9147ac0d86ad0f170195b6593c67958ac1a9a4120ee87672102ca505bf28698f0b6c26114a725f757b88d65537dd52a5b6455a9cac9581f1055ac640396460cb275686800000000",
						BtcSweepTx:            "010000000001019fc4e4a9cec5690c84f3773c383936a82c107167e2af7e4c7af48212fc390f880000000000f5ffffff021027000000000000160014a96c26be9d01a07078751ce66c46986db5e3c609a86100000000000022002010118cefe99eb61b69dc77601e7861b3dc80135e319538b96b401cdb2151a61f072038e542280948922ae2b9b64c39407b4581b862d8baa791d28b914e7f5e1fc9780100473044022045a0e3360112645a7e73abe8e09c8057e2ed09695f5ea6d905f9cb32627000ee02206a451d2cc07de37d1a8167522dcf1018fe259883c6f5ef12d2d1c5e6b07a303781483045022100de82176a9c45ba63bca9dbe65540d283acd0b8feda6a947179607191a13048e40220593b33dfbfc86dd9458e215ff7f5b3d5dea494ed7be98de2c1a515d0d49595a9814730440220735207b13f20376e0d87c3470e9b8c0c78c9cda912c60eff26e4e61505e5b3c002204519d13f71be59fa9c0871762bddf9e6b732ee37485a394d0aa04131cf50b67581483045022100b4eacf4b88f75634c3c31fef4a57ea014aa699b203a35224b96b5f62eb0e0e8e02202d7f1f052a74631d6f02185a6e98e69ab5d55101564f2626af16090a1e0584ce81f6532102ca505bf28698f0b6c26114a725f757b88d65537dd52a5b6455a9cac9581f105521038b38721dbb1427fd9c65654f87cb424517df717ee2fea8b0a5c376a1734941672103b03fe3da02ac2d43a1c2ebcfc7b0497e89cc9f62b513c0fc14f10d3d1a2cd5e62103bb3694e798f018a157f9e6dfb51b91f70a275443504393040892b52e45b255c32103e2f80f2f5eb646df3e0642ae137bf13f5a9a6af4c05688e147c64e8fae196fe155ae6382012088a9147ac0d86ad0f170195b6593c67958ac1a9a4120ee87672102ca505bf28698f0b6c26114a725f757b88d65537dd52a5b6455a9cac9581f1055ac640396460cb275686800000000",
					}
					sendTransactionSweepProposal(accountName, cosmos, msg)
					markTransactionProcessed(tx.Txid)

				}
			}
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
	go kDeepService(accountName)
}
