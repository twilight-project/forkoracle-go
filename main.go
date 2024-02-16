package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"github.com/tyler-smith/go-bip32"

	"github.com/gorilla/websocket"
)

var dbconn *sql.DB
var masterPrivateKey *bip32.Key
var judge bool
var oracleAddr string
var valAddr string
var upgrader = websocket.Upgrader{}
var WsHub *Hub

var latestSweepTxHash = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "latest_sweep_tx_hash",
		Help: "Hash of the latest swept transaction.",
	},
	[]string{"hash"},
)

func initialize() {
	initConfigFile()
	btcPubkey := initWallet()
	dbconn = initDB()
	setDelegator(btcPubkey)
}

func main() {

	initialize()

	accountName := fmt.Sprintf("%v", viper.Get("accountName"))
	fmt.Println("account name : ", accountName)
	var forkscanner_host = fmt.Sprintf("%v:%v", viper.Get("forkscanner_host"), viper.Get("forkscanner_ws_port"))
	forkscanner_url := url.URL{Scheme: "ws", Host: forkscanner_host, Path: "/"}
	if accountName == "validator-sfo" || accountName == "validator-ams" || accountName == "validator-stg" {
		judge = true
	}

	time.Sleep(30 * time.Second)

	go orchestrator(accountName, forkscanner_url)

	initJudge(accountName)

	time.Sleep(1 * time.Minute)
	if judge {
		go startJudge(accountName)
	} else {
		time.Sleep(2 * time.Minute)
	}

	time.Sleep(1 * time.Minute)
	go startBridge(accountName, forkscanner_url)
	go pubsubServer()
	go startTransactionSigner(accountName)
	prometheus_server()
	fmt.Println("exiting main")
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

func (c *Client) writePump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}

func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, _ := upgrader.Upgrade(w, r, nil)
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	go client.writePump()
}

func pubsubServer() {
	fmt.Println("starting pubsub server")
	WsHub = &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}

	go WsHub.run()

	http.HandleFunc("/tapinscription", func(w http.ResponseWriter, r *http.Request) {
		serveWs(WsHub, w, r)
	})

	err := http.ListenAndServe(":2300", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func prometheus_server() {
	// Create a new instance of a registry
	reg := prometheus.NewRegistry()

	// Optional: Add Go module build info.
	reg.MustRegister(
		latestSweepTxHash,
	)

	// Register the promhttp handler with the registry
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	// Simple health check endpoint
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Server is running"))
	})

	// Start the server
	log.Println("Starting prometheus server on :2555")
	if err := http.ListenAndServe(":2555", nil); err != nil {
		log.Fatalf("Error starting server: %s", err)
	}
}

// func main() {

// 	x := "830572 OP_CHECKLOCKTIMEVERIFY OP_DROP 4 02ca505bf28698f0b6c26114a725f757b88d65537dd52a5b6455a9cac9581f1055 033e72f302ba2133eddd0c7416943d4fed4e7c60db32e6b8c58895d3b26e24f927 038b38721dbb1427fd9c65654f87cb424517df717ee2fea8b0a5c376a173494167 03b03fe3da02ac2d43a1c2ebcfc7b0497e89cc9f62b513c0fc14f10d3d1a2cd5e6 03bb3694e798f018a157f9e6dfb51b91f70a275443504393040892b52e45b255c3 03e2f80f2f5eb646df3e0642ae137bf13f5a9a6af4c05688e147c64e8fae196fe1 6 OP_CHECKMULTISIGVERIFY OP_SIZE 32 OP_EQUALVERIFY OP_HASH160 e3dc0fa779409b9a538ab9f1ac4e66c08c657059 OP_EQUAL OP_IFDUP OP_NOTIF 02ca505bf28698f0b6c26114a725f757b88d65537dd52a5b6455a9cac9581f1055 OP_CHECKSIG OP_NOTIF 830583 OP_CHECKSEQUENCEVERIFY OP_DROP OP_ENDIF OP_ENDIF"

// 	y := getPublicKeysFromScript(x, 4)
// 	// z, _ := convertHextoDec(y[0])
// 	fmt.Println(y)
// }
