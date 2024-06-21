package main

import (
	"database/sql"
	"fmt"
	"net/url"
	"time"

	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"

	"github.com/twilight-project/forkoracle-go/address"
	"github.com/twilight-project/forkoracle-go/bridge"
	db "github.com/twilight-project/forkoracle-go/db"
	"github.com/twilight-project/forkoracle-go/eventhandler"
	"github.com/twilight-project/forkoracle-go/judge"
	"github.com/twilight-project/forkoracle-go/orchestrator"
	"github.com/twilight-project/forkoracle-go/servers"
	btcOracleTypes "github.com/twilight-project/forkoracle-go/types"
	utils "github.com/twilight-project/forkoracle-go/utils"
)

func initialize() (string, string, *sql.DB) {
	utils.InitConfigFile()
	btcPubkey := utils.GetBtcPublicKey()
	dbconn := db.InitDB()
	valAddr, oracleAddr := utils.SetDelegator(btcPubkey)
	return valAddr, oracleAddr, dbconn
}

func main() {
	var activeJudge bool

	// var upgrader = websocket.Upgrader{}
	var WsHub *btcOracleTypes.Hub

	var latestSweepTxHash = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "latest_sweep_tx_hash",
			Help: "Hash of the latest swept transaction.",
		},
		[]string{"hash"},
	)

	var latestRefundTxHash = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "latest_refund_tx_hash",
			Help: "Hash of the latest swept transaction.",
		},
		[]string{"hash"},
	)

	valAddr, oracleAddr, dbconn := initialize()
	fmt.Println("valAddr : ", valAddr)
	fmt.Println("oracleAddr : ", oracleAddr)

	accountName := fmt.Sprintf("%v", viper.Get("accountName"))
	fmt.Println("account name : ", accountName)
	var forkscanner_host = fmt.Sprintf("%v:%v", viper.Get("forkscanner_host"), viper.Get("forkscanner_ws_port"))
	forkscanner_url := url.URL{Scheme: "ws", Host: forkscanner_host, Path: "/"}
	if accountName == "validator-sfo" || accountName == "validator-ams" || accountName == "validator-stg" {
		activeJudge = true
	}

	time.Sleep(30 * time.Second)

	go orchestrator.Orchestrator(accountName, forkscanner_url, oracleAddr)

	if activeJudge {
		judge.InitJudge(accountName, dbconn, oracleAddr, valAddr)
	}

	time.Sleep(1 * time.Minute)
	if activeJudge {
		go startJudge(accountName, dbconn, oracleAddr, valAddr, WsHub, latestRefundTxHash)
	} else {
		time.Sleep(2 * time.Minute)
	}

	time.Sleep(1 * time.Minute)
	go startBridge(accountName, forkscanner_url, dbconn, latestSweepTxHash, oracleAddr, valAddr, WsHub)
	// go servers.PubsubServer(WsHub, upgrader)
	go startTransactionSigner(accountName, dbconn, oracleAddr, valAddr, WsHub)
	servers.Prometheus_server(latestSweepTxHash, latestRefundTxHash)
	fmt.Println("exiting main")
}

func startTransactionSigner(accountName string, dbconn *sql.DB, oracleAddr string, valAddr string, WsHub *btcOracleTypes.Hub) {
	fmt.Println("starting Transaction Signer")
	go eventhandler.NyksEventListener("unsigned_tx_refund", accountName, "signing_refund", dbconn, oracleAddr, valAddr, WsHub, nil)
	eventhandler.NyksEventListener("broadcast_tx_refund", accountName, "signing_sweep", dbconn, oracleAddr, valAddr, WsHub, nil)
	fmt.Println("finishing bridge")
}

func startJudge(accountName string, dbconn *sql.DB, oracleAddr string, valAddr string, WsHub *btcOracleTypes.Hub, latestRefundTxHash *prometheus.GaugeVec) {
	fmt.Println("starting judge")
	go address.ProcessProposeAddress(accountName, oracleAddr, dbconn)
	go judge.BroadcastOnBtc(dbconn)
	go eventhandler.NyksEventListener("propose_sweep_address", accountName, "sweep_process", dbconn, oracleAddr, valAddr, WsHub, nil)
	go eventhandler.NyksEventListener("broadcast_tx_refund", accountName, "signed_sweep_process", dbconn, oracleAddr, valAddr, WsHub, latestRefundTxHash)
	go eventhandler.NyksEventListener("unsigned_tx_sweep", accountName, "refund_process", dbconn, oracleAddr, valAddr, WsHub, nil)
	eventhandler.NyksEventListener("unsigned_tx_refund", accountName, "signed_refund_process", dbconn, oracleAddr, valAddr, WsHub, nil)
}

func startBridge(accountName string, forkscanner_url url.URL, dbconn *sql.DB, latestSweepTxHash *prometheus.GaugeVec, oracleAddr string, valAddr string, WsHub *btcOracleTypes.Hub) {
	fmt.Println("starting bridge")
	address.RegisterAddressOnValidators(dbconn)
	go eventhandler.NyksEventListener("propose_sweep_address", accountName, "register_res_addr_validators", dbconn, oracleAddr, valAddr, WsHub, nil)
	go bridge.WatchAddress(forkscanner_url, dbconn)
	bridge.KDeepService(accountName, dbconn, latestSweepTxHash, oracleAddr)
	fmt.Println("finishing bridge")
}

// func main() {

// 	initialize()
// 	// accountName := fmt.Sprintf("%v", viper.Get("accountName"))
// 	tx := "01000000000101588bcebce384c4849575e8014826945b19d05eec9ed0df7b2c95b017f8993f5700000000006cee0c00011027000000000000220020252ab890aebdbe6dc2b29340edffa3947996362654436fe2c3d0852de4f0d21401fd1e0103eceb0cb1755421038b38721dbb1427fd9c65654f87cb424517df717ee2fea8b0a5c376a17349416721033e72f302ba2133eddd0c7416943d4fed4e7c60db32e6b8c58895d3b26e24f9272103bb3694e798f018a157f9e6dfb51b91f70a275443504393040892b52e45b255c32103b03fe3da02ac2d43a1c2ebcfc7b0497e89cc9f62b513c0fc14f10d3d1a2cd5e62102ca505bf28698f0b6c26114a725f757b88d65537dd52a5b6455a9cac9581f10552103e2f80f2f5eb646df3e0642ae137bf13f5a9a6af4c05688e147c64e8fae196fe156af82012088a9149dc1332b2da58986c341cf8bbd97d3869efc10918773642102ca505bf28698f0b6c26114a725f757b88d65537dd52a5b6455a9cac9581f1055ac6403f7eb0cb275686871ee0c00"
// 	sweepTx, _ := utils.CreateTxFromHex(tx)
// 	utils.SignTx(sweepTx, []byte{})

// }

// nyksd tx bridge sign-refund 1 1 03b03fe3da02ac2d43a1c2ebcfc7b0497e89cc9f62b513c0fc14f10d3d1a2cd5e6 3045022100978ef8d96b62a0738b5c4a109720985609fb5ca39244b09905979d3373cfb78802206e6ddf5653d1a7fe5be605dd12b5fcfc022950c29711f39d05d5fac8bf81de8001 --from validator-fra --chain-id nyks  --keyring-backend test
