package main

import (
	"database/sql"
	"fmt"
	"net/url"
	"time"

	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"github.com/tyler-smith/go-bip32"

	"github.com/gorilla/websocket"
	address "github.com/twilight-project/forkoracle-go/address"
	"github.com/twilight-project/forkoracle-go/bridge"
	db "github.com/twilight-project/forkoracle-go/db"
	"github.com/twilight-project/forkoracle-go/eventhandler"
	"github.com/twilight-project/forkoracle-go/judge"
	"github.com/twilight-project/forkoracle-go/orchestrator"
	"github.com/twilight-project/forkoracle-go/servers"
	btcOracleTypes "github.com/twilight-project/forkoracle-go/types"
	utils "github.com/twilight-project/forkoracle-go/utils"
	wallet "github.com/twilight-project/forkoracle-go/wallet"
)

func initialize() (string, string, *sql.DB, *bip32.Key) {
	utils.InitConfigFile()
	btcPubkey, masterPrivateKey := wallet.InitWallet()
	dbconn := db.InitDB()
	valAddr, oracleAddr := utils.SetDelegator(btcPubkey)
	return valAddr, oracleAddr, dbconn, masterPrivateKey
}

func main() {
	var activeJudge bool

	var upgrader = websocket.Upgrader{}
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
	valAddr, oracleAddr, dbconn, masterPrivateKey := initialize()
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
		go startJudge(accountName, dbconn, oracleAddr, valAddr, WsHub, masterPrivateKey, latestRefundTxHash)
	} else {
		time.Sleep(2 * time.Minute)
	}

	time.Sleep(1 * time.Minute)
	go startBridge(accountName, forkscanner_url, dbconn, latestSweepTxHash, oracleAddr, masterPrivateKey, valAddr, WsHub)
	go servers.PubsubServer(WsHub, upgrader)
	go startTransactionSigner(accountName, masterPrivateKey, dbconn, oracleAddr, valAddr, WsHub)
	servers.Prometheus_server(latestSweepTxHash, latestRefundTxHash)
	fmt.Println("exiting main")
}

func startTransactionSigner(accountName string, masterPrivateKey *bip32.Key, dbconn *sql.DB, oracleAddr string, valAddr string, WsHub *btcOracleTypes.Hub) {
	fmt.Println("starting Transaction Signer")
	go eventhandler.NyksEventListener("unsigned_tx_refund", accountName, "signing_refund", masterPrivateKey, dbconn, oracleAddr, valAddr, WsHub, nil)
	eventhandler.NyksEventListener("broadcast_tx_refund", accountName, "signing_sweep", masterPrivateKey, dbconn, oracleAddr, valAddr, WsHub, nil)
	fmt.Println("finishing bridge")
}

func startJudge(accountName string, dbconn *sql.DB, oracleAddr string, valAddr string, WsHub *btcOracleTypes.Hub, masterPrivateKey *bip32.Key, latestRefundTxHash *prometheus.GaugeVec) {
	fmt.Println("starting judge")
	go address.ProcessProposeAddress(accountName, oracleAddr, dbconn)
	go judge.BroadcastOnBtc(dbconn)
	go eventhandler.NyksEventListener("propose_sweep_address", accountName, "sweep_process", masterPrivateKey, dbconn, oracleAddr, valAddr, WsHub, nil)
	go eventhandler.NyksEventListener("broadcast_tx_refund", accountName, "signed_sweep_process", masterPrivateKey, dbconn, oracleAddr, valAddr, WsHub, latestRefundTxHash)
	go eventhandler.NyksEventListener("unsigned_tx_sweep", accountName, "refund_process", masterPrivateKey, dbconn, oracleAddr, valAddr, WsHub, nil)
	eventhandler.NyksEventListener("unsigned_tx_refund", accountName, "signed_refund_process", masterPrivateKey, dbconn, oracleAddr, valAddr, WsHub, nil)
}

func startBridge(accountName string, forkscanner_url url.URL, dbconn *sql.DB, latestSweepTxHash *prometheus.GaugeVec, oracleAddr string,
	masterPrivateKey *bip32.Key, valAddr string, WsHub *btcOracleTypes.Hub) {
	fmt.Println("starting bridge")
	address.RegisterAddressOnValidators(dbconn)
	go eventhandler.NyksEventListener("propose_sweep_address", accountName, "register_res_addr_validators", masterPrivateKey, dbconn, oracleAddr, valAddr, WsHub, nil)
	go bridge.WatchAddress(forkscanner_url, dbconn)
	bridge.KDeepService(accountName, dbconn, latestSweepTxHash, oracleAddr)
	fmt.Println("finishing bridge")
}

// func main() {

// 	initialize()
// 	// accountName := fmt.Sprintf("%v", viper.Get("accountName"))
// 	sweeptx := getUnsignedSweepTx(1, 1)
// 	tx := sweeptx.UnsignedTxSweepMsg.BtcUnsignedSweepTx
// 	sweepTx, _ := createTxFromHex(tx)

// 	signatureSweep := getSignSweep(1, 1)
// 	x := sweepTx.TxIn[0].Witness[0]
// 	hx := hex.EncodeToString(x)
// 	decodedscript := decodeBtcScript(hx)
// 	min := getMinSignFromScript(decodedscript)
// 	pubkeys := getPublicKeysFromScript(decodedscript, int(min))

// 	t := filterAndOrderSignSweep(signatureSweep, pubkeys)

// 	fmt.Println(t)
// }
