package main

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"github.com/twilight-project/forkoracle-go/address"
	"github.com/twilight-project/forkoracle-go/bridge"
	"github.com/twilight-project/forkoracle-go/db"
	"github.com/twilight-project/forkoracle-go/eventhandler"
	"github.com/twilight-project/forkoracle-go/judge"
	"github.com/twilight-project/forkoracle-go/keyring"
	"github.com/twilight-project/forkoracle-go/orchestrator"
	"github.com/twilight-project/forkoracle-go/servers"
	btcOracleTypes "github.com/twilight-project/forkoracle-go/types"
	"github.com/twilight-project/forkoracle-go/utils"
)

func initialize() (string, string, *sql.DB, keyring.Keyring) {
	utils.InitConfigFile()
	mockIn := strings.NewReader("")
	keyring_dir := viper.GetString("keyring_dir")
	keyring_name := viper.GetString("keyring_name")
	kr, err := keyring.New("nyks", keyring.BackendFile, keyring_dir, mockIn)
	if err != nil {
		fmt.Println(err)
	}
	keys_list, _ := kr.List()
	if keys_list == nil {
		info, mnemonic, err := kr.NewMnemonic(keyring_name, keyring.English, "m/44'/118'/0'/0/0", keyring.DefaultBIP39Passphrase, hd.Secp256k1)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
		fmt.Println(info.GetPubKey().String())
		fmt.Println("your mnemonic has been printed in the mnemonic.txt file in the keyring directory. Please keep copy it and delete the file.")
		utils.WriteToFile(keyring_dir+"mnemonic.txt", mnemonic)
	}
	dbconn := db.InitDB()
	key, err := kr.Key(keyring_name)
	if err != nil {
		panic(err)
	}
	valAddr, oracleAddr := utils.SetDelegator(key.GetPubKey().String())
	return valAddr, oracleAddr, dbconn, kr
}

func main() {

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

	valAddr, oracleAddr, dbconn, keyring := initialize()
	fmt.Println("valAddr : ", valAddr)
	fmt.Println("oracleAddr : ", oracleAddr)

	accountName := fmt.Sprintf("%v", viper.Get("accountName"))
	fmt.Println("account name : ", accountName)
	var forkscanner_host = fmt.Sprintf("%v:%v", viper.Get("forkscanner_host"), viper.Get("forkscanner_ws_port"))
	forkscanner_url := url.URL{Scheme: "ws", Host: forkscanner_host, Path: "/"}
	activeJudge := viper.GetBool("judge")
	signer := viper.GetBool("signer")

	time.Sleep(30 * time.Second)

	go orchestrator.Orchestrator(accountName, forkscanner_url, oracleAddr)

	if activeJudge {
		judge.InitJudge(accountName, dbconn, oracleAddr, valAddr)
	}

	time.Sleep(1 * time.Minute)
	if activeJudge {
		go startJudge(accountName, dbconn, oracleAddr, valAddr, WsHub, keyring, latestRefundTxHash)
	} else {
		time.Sleep(2 * time.Minute)
	}

	time.Sleep(1 * time.Minute)
	go startBridge(accountName, forkscanner_url, dbconn, latestSweepTxHash, oracleAddr, keyring, valAddr, WsHub)
	// go servers.PubsubServer(WsHub, upgrader)
	if signer {
		go startTransactionSigner(accountName, keyring, dbconn, oracleAddr, valAddr, WsHub)
	}
	servers.Prometheus_server(latestSweepTxHash, latestRefundTxHash)
	fmt.Println("exiting main")
}

func startTransactionSigner(accountName string, keyring keyring.Keyring, dbconn *sql.DB, oracleAddr string, valAddr string, WsHub *btcOracleTypes.Hub) {
	fmt.Println("starting Transaction Signer")
	go eventhandler.NyksEventListener("unsigned_tx_refund", accountName, "signing_refund", keyring, dbconn, oracleAddr, valAddr, WsHub, nil)
	eventhandler.NyksEventListener("broadcast_tx_refund", accountName, "signing_sweep", keyring, dbconn, oracleAddr, valAddr, WsHub, nil)
	fmt.Println("finishing bridge")
}

func startJudge(accountName string, dbconn *sql.DB, oracleAddr string, valAddr string, WsHub *btcOracleTypes.Hub, keyring keyring.Keyring, latestRefundTxHash *prometheus.GaugeVec) {
	fmt.Println("starting judge")
	go address.ProcessProposeAddress(accountName, oracleAddr, dbconn)
	go judge.BroadcastOnBtc(dbconn)
	go eventhandler.NyksEventListener("propose_sweep_address", accountName, "sweep_process", keyring, dbconn, oracleAddr, valAddr, WsHub, nil)
	go eventhandler.NyksEventListener("broadcast_tx_refund", accountName, "signed_sweep_process", keyring, dbconn, oracleAddr, valAddr, WsHub, latestRefundTxHash)
	go eventhandler.NyksEventListener("unsigned_tx_sweep", accountName, "refund_process", keyring, dbconn, oracleAddr, valAddr, WsHub, nil)
	eventhandler.NyksEventListener("unsigned_tx_refund", accountName, "signed_refund_process", keyring, dbconn, oracleAddr, valAddr, WsHub, nil)
}

func startBridge(accountName string, forkscanner_url url.URL, dbconn *sql.DB, latestSweepTxHash *prometheus.GaugeVec, oracleAddr string,
	keyring keyring.Keyring, valAddr string, WsHub *btcOracleTypes.Hub) {
	fmt.Println("starting bridge")
	address.RegisterAddressOnValidators(dbconn)
	go eventhandler.NyksEventListener("propose_sweep_address", accountName, "register_res_addr_validators", keyring, dbconn, oracleAddr, valAddr, WsHub, nil)
	go bridge.WatchAddress(forkscanner_url, dbconn)
	bridge.KDeepService(accountName, dbconn, latestSweepTxHash, oracleAddr)
	fmt.Println("finishing bridge")
}

// func main() {
// 	mockIn := strings.NewReader("password\npassword\n")
// 	kr, err := keyring.New("nyks", keyring.BackendFile, "", mockIn)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	keys_list, _ := kr.List()
// 	if keys_list == nil {
// 		info, mnemonic, err := kr.NewMnemonic("btc_keyring", keyring.English, "m/44'/118'/0'/0/0", keyring.DefaultBIP39Passphrase, hd.Secp256k1)
// 		if err != nil {
// 			fmt.Println(err)
// 		}
// 		fmt.Println(info.GetPubKey())
// 		fmt.Println(mnemonic)
// 	}
// 	k, err := kr.Key("btc_keyring")
// 	fmt.Println(k.GetPubKey())
// 	kr.SignTx(k.GetName(), nil, nil, nil)
// 	fmt.Println(k.GetPubKey())
// }
