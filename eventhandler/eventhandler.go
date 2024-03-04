package eventhandler

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
	"github.com/tyler-smith/go-bip32"
)

func NyksEventListener(event string, accountName string, functionCall string, masterPrivateKey *bip32.Key, dbconn *sql.DB) {
	headers := make(map[string][]string)
	headers["Content-Type"] = []string{"application/json"}
	nyksd_url := fmt.Sprintf("%v", viper.Get("nyksd_socket_url"))
	conn, _, err := websocket.DefaultDialer.Dial(nyksd_url, headers)
	if err != nil {
		fmt.Println("nyks event listerner dial:", err)
	}
	defer conn.Close()

	// Set up ping/pong connection health check
	pingPeriod := 30 * time.Second
	pongWait := 60 * time.Second
	stopChan := make(chan struct{}) // Create the stop channel

	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	go func() {
		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			case <-stopChan: // Listen to the stop channel
				return
			}
		}
	}()

	payload := `{
        "jsonrpc": "2.0",
        "method": "subscribe",
        "id": 0,
        "params": {
            "query": "tm.event='Tx' AND message.action='%s'"
        }
    }`
	payload = fmt.Sprintf(payload, event)

	err = conn.WriteMessage(websocket.TextMessage, []byte(payload))
	if err != nil {
		fmt.Println("error in nyks event handler: ", err)
		stopChan <- struct{}{} // Signal goroutine to stop
		return
	}

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("error in nyks event handler: ", err)
			stopChan <- struct{}{} // Signal goroutine to stop
			return
		}

		// var event Event
		// err = json.Unmarshal(message, &event)
		// if err != nil {
		// 	fmt.Println("error unmarshalling event: ", err)
		// 	continue
		// }

		// fmt.Print("event : ", event)
		// fmt.Print("event : ", message)

		// if event.Method == "subscribe" && event.Params.Query == fmt.Sprintf("tm.event='Tx' AND message.action='%s'", event) {
		// 	continue
		// }

		switch functionCall {
		case "signed_sweep_process":
			go processSignedSweep(accountName)
		case "refund_process":
			go processRefund(accountName)
		case "signed_refund_process":
			go processSignedRefund(accountName)
		case "register_res_addr_validators":
			go registerAddressOnValidators()
		case "signing_sweep":
			go processTxSigningSweep(accountName)
		case "signing_refund":
			go processTxSigningRefund(accountName)
		case "sweep_process":
			go processSweep(accountName)
		default:
			log.Println("Unknown function :", functionCall)
		}
	}
}
