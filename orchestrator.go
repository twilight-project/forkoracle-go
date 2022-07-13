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

func orchestrator() {
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
			process_message(message)
			log.Printf("recv: %s", message)
		}
	}()

	tunnel := make(chan bool)

	payload := `{"jsonrpc": "2.0", "id": 1, "method": "subscribe_forks", "params": null}`

	go func() {

		// Setting the value of channel
		tunnel <- true
	}()

	for {
		select {
		case <-done:
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
func process_message(message []byte) {
	var c BlockData
	err := json.Unmarshal(message, &c)
	if err != nil {
		fmt.Printf("Unmarshal: %v\n", err)
	}

	log.Println("new_message", c.ChainTip)

	for i, item := range c.ChainTip {
		fmt.Printf("Row: %v\n", i)
		fmt.Printf("Row: %v\n", item[1].Node)
		fmt.Println(c.ChainTip[i][i].Block)
		send_transaction(item[0])
	}

}

// func orchestrator() {
// 	tick := time.Tick(5000 * time.Millisecond)
// 	for range tick {
// 		fmt.Println("Tick")
// 		chaintip := get_active_chaintip()
// 		send_transaction((chaintip))
// 	}
// }

// func get_active_chaintip() *ChainTip {

// 	rpcClient := jsonrpc.NewClient("http://127.0.0.1:8339")
// 	response, err := rpcClient.Call("getforks")

// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	var c []*ChainTip
// 	response.GetObject(&c)

// 	//fmt.Printf("%v\n", c[0].Block)
// 	return c[0]
// }
