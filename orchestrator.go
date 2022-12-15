package main

import (
	"encoding/json"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	"github.com/twilight-project/nyks/x/forks/types"
)

func orchestrator(accountName string, forkscanner_url url.URL) {
	log.SetFlags(0)

	go start_bridge(accountName, forkscanner_url)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	log.Printf("connecting to %s", forkscanner_url.String())

	c, _, err := websocket.DefaultDialer.Dial(forkscanner_url.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			process_message(accountName, message)
			log.Printf("recv: %s", message)
		}
	}()

	tunnel := make(chan bool)

	payload := `{"jsonrpc": "2.0", "id": 1, "method": "subscribe_active_fork", "params": null}`

	go func() {

		// Setting the value of channel
		tunnel <- true
	}()

	for {
		select {
		case <-done:
			log.Println("done")
			return
		case <-tunnel:
			err := c.WriteMessage(websocket.TextMessage, []byte(payload))
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

// This function is a buffer between websocket and send_transaction to add future functionality
func process_message(accountName string, message []byte) {
	var c BlockData
	err := json.Unmarshal(message, &c)
	if err != nil {
		log.Printf("Unmarshal: %v\n", err)
	}

	log.Println("new_message test", c)

	active_chaintips := c.ChainTip

	if len(active_chaintips) <= 0 {
		log.Println("first mesaage or empty list")
		return
	}

	log.Printf("active chain tip : ", active_chaintips[0])
	log.Printf("Row: %v\n", active_chaintips[0].Node)
	log.Println(active_chaintips[0].Block)

	active_chaintip := active_chaintips[0]

	cosmos_client := get_cosmos_client()
	cosmos_address := get_cosmos_address(accountName, cosmos_client)

	msg := &types.MsgSeenBtcChainTip{
		Height:           uint64(active_chaintip.Height),
		Hash:             active_chaintip.Block,
		BtcOracleAddress: cosmos_address.String(),
	}

	send_transaction(accountName, cosmos_client, msg, "SeenBtcChainTip")

}
