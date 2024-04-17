package servers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	btcOracleTypes "github.com/twilight-project/forkoracle-go/types"
)

func serveWs(hub *btcOracleTypes.Hub, w http.ResponseWriter, r *http.Request, upgrader websocket.Upgrader) {
	conn, _ := upgrader.Upgrade(w, r, nil)
	client := &btcOracleTypes.Client{Hub: hub, Conn: conn, Send: make(chan []byte, 256)}
	client.Hub.Register <- client

	go client.WritePump()
}

func PubsubServer(hub *btcOracleTypes.Hub, upgrader websocket.Upgrader) {
	fmt.Println("starting pubsub server")
	hub = &btcOracleTypes.Hub{
		Broadcast:  make(chan []byte),
		Register:   make(chan *btcOracleTypes.Client),
		Unregister: make(chan *btcOracleTypes.Client),
		Clients:    make(map[*btcOracleTypes.Client]bool),
	}

	go hub.Run()

	http.HandleFunc("/tapinscription", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r, upgrader)
	})

	err := http.ListenAndServe(":2300", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func Prometheus_server(latestSweepTxHash *prometheus.GaugeVec, latestRefundTxHash *prometheus.GaugeVec) {
	// Create a new instance of a registry
	reg := prometheus.NewRegistry()

	// Optional: Add Go module build info.
	reg.MustRegister(
		latestSweepTxHash,
		latestRefundTxHash,
	)

	// Register the promhttp handler with the registry
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	// Simple health check endpoint
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("Server is running"))
		if err != nil {
			fmt.Printf("Error writing response: %s \n", err)
		}
	})

	// Start the server
	log.Println("Starting prometheus server on :2555")
	if err := http.ListenAndServe(":2555", nil); err != nil {
		log.Fatalf("Error starting server: %s", err)
	}
}
