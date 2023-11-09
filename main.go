package main

import (
	"database/sql"
	"fmt"
	"net/url"
	"time"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"github.com/tyler-smith/go-bip32"
)

var dbconn *sql.DB
var masterPrivateKey *bip32.Key
var judge bool
var oracleAddr string
var valAddr string

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
	if accountName == "validator-sfo" || accountName == "validator-ams" {
		judge = true
	}

	time.Sleep(30 * time.Second)

	go orchestrator(accountName, forkscanner_url)

	initJudge(accountName)

	time.Sleep(1 * time.Minute)
	if judge == true {
		go startJudge(accountName)
	} else {
		time.Sleep(5 * time.Minute)
	}

	time.Sleep(1 * time.Minute)
	go startBridge(accountName, forkscanner_url)
	startTransactionSigner(accountName)

	//for local testing
	// x := "0100000000010223ee2ee6d96a7bc7211dc9d01a6388038a58ab7134035371041eee87346b9c580100000000fffffffff79cd6c051c1a7469ea0d905235ae9b2ec385dadc61eecc165a1dc03044b688d0300000000ffffffff02f37a010000000000160014e9533474b0d578081b3c446eea8b3d629c4835e6ad11000000000000160014db7092896f663c9d74e677469b10b79293b7f8580247304402207998c5a29c69d42c8ef924e6aafc9a8fd25e07449f3681786da8c6fc87ab347d02201d16c0b118fc20f3797500a18d1792c9301ec58c61f5bcef71bedd95f6e027b1012103c3f95c948d228cd9ca7f351b14f528c8fb5bfab7f2f240515ecbe2f80ccda12202473044022048f699c2e9968a141e51c79c4dbff2db81ffd13bdbfa2c4996d78dba9580ec25022023f4ff1456549c0906adc0ee754040f6987005b1e0161639b295d2938fa59b220121021bd6eac3ebd73a4c77ec6756ab21f05f0c93bb6f9e6cb0c59904117076dcc75800000000"

	// tx, _ := createTxFromHex(x)
	// feeRate := getBtcFeeRate()
	// baseSize := tx.SerializeSizeStripped()
	// totalSize := tx.SerializeSize()
	// weight := (baseSize * 3) + totalSize
	// vsize := (weight + 3) / 4
	// requiredFee := vsize * feeRate.Priority
	// fmt.Println(requiredFee)

}
