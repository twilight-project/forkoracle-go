package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

func orchestrator(accountName string) {
	var addr = flag.String("addr", "0.0.0.0:8340", "http service address")

	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
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
		fmt.Printf("Unmarshal: %v\n", err)
	}

	log.Println("new_message test", c)

	active_chaintip := c.ChainTip

	if len(active_chaintip) <= 0 {
		log.Println("first mesaage or empty list")
		return
	}
	log.Printf("active chain tip : ", active_chaintip[0])
	log.Printf("Row: %v\n", active_chaintip[0].Node)
	log.Println(active_chaintip[0].Block)
	send_transaction(accountName, active_chaintip[0])

}
