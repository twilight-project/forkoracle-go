package main

import (
	"bytes"
	"database/sql"
	"encoding/hex"
	"flag"
	"fmt"
	"net/url"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/btcsuite/btcd/btcec"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"github.com/tyler-smith/go-bip32"
)

var dbconn *sql.DB
var masterPrivateKey *bip32.Key
var judge bool
var oracleAddr string
var valAddr string
var wg sync.WaitGroup

func initialize() {
	//config setup
	viper.AddConfigPath("/testnet/btc-oracle/configs")
	viper.SetConfigName("config") // Register config file name (no extension)
	viper.SetConfigType("json")   // Look for specific type
	viper.ReadInConfig()

	//wallet setup
	new_wallet := flag.Bool("new_wallet", true, "set to true if you want to create a new wallet")
	mnemonic := flag.String("mnemonic", "", "mnemonic for the wallet, leave empty to generate a new nemonic")
	flag.Parse()

	var err error

	walletPassphrase := fmt.Sprintf("%v", viper.Get("wallet_passphrase"))
	if *new_wallet == true {
		if *mnemonic != "" {
			masterPrivateKey, err = create_wallet_from_mnemonic(*mnemonic, walletPassphrase)
			if err != nil {
				fmt.Println("Error creating wallet from mnemonic:", err)
				panic(err)
			}
		} else {
			masterPrivateKey, err = create_wallet(walletPassphrase)
			if err != nil {
				fmt.Println("Error creating wallet:", err)
				panic(err)
			}
		}
	} else {
		masterPrivateKey, err = load_wallet(walletPassphrase)
		if err != nil {
			fmt.Println("Error loading wallet:", err)
			panic(err)
		}
	}

	privkeybytes, err := masterPrivateKey.Serialize()
	if err != nil {
		fmt.Println("Error: converting private key to bytes : ", err)
	}

	privkey, _ := btcec.PrivKeyFromBytes(btcec.S256(), privkeybytes)

	btcPubkey := hex.EncodeToString(privkey.PubKey().SerializeCompressed())
	fmt.Println("Wallet initialized")

	// db connection
	dbconn = initDB()
	fmt.Println("DB initialized")

	accountName := fmt.Sprintf("%v", viper.Get("accountName"))
	command := fmt.Sprintf("nyksd keys show %s --bech val -a --keyring-backend test", accountName)
	args := strings.Fields(command)
	cmd := exec.Command(args[0], args[1:]...)

	valAddr_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	valAddr = string(valAddr_)
	oracleAddr = strings.ReplaceAll(oracleAddr, "\n", "")
	fmt.Println("Val Address : ", valAddr)

	command = fmt.Sprintf("nyksd keys show %s -a --keyring-backend test", accountName)
	args = strings.Fields(command)
	cmd = exec.Command(args[0], args[1:]...)

	oracleAddr_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	oracleAddr = string(oracleAddr_)
	oracleAddr = strings.ReplaceAll(oracleAddr, "\n", "")
	fmt.Println("Oracle Address : ", oracleAddr)

	command = fmt.Sprintf("nyksd tx nyks set-delegate-addresses %s %s %s --from %s --chain-id nyks --keyring-backend test -y", valAddr, oracleAddr, btcPubkey, accountName)
	fmt.Println("delegate command : ", command)

	args = strings.Fields(command)
	cmd = exec.Command(args[0], args[1:]...)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		panic(err)
	}

	fmt.Println("Delegate Address Set")
}

func main() {

	initialize()

	accountName := fmt.Sprintf("%v", viper.Get("accountName"))
	fmt.Println("account name : ", accountName)

	var forkscanner_host = fmt.Sprintf("%v:%v", viper.Get("forkscanner_host"), viper.Get("forkscanner_ws_port"))
	forkscanner_url := url.URL{Scheme: "ws", Host: forkscanner_host, Path: "/"}
	if accountName == "validator-sfo" {
		judge = true
	}
	time.Sleep(1 * time.Minute)

	wg.Add(1)
	go orchestrator(accountName, forkscanner_url)

	if judge == true {
		addr := queryAllSweepAddresses()
		if len(addr) <= 0 {
			wg.Add(1)
			time.Sleep(1 * time.Minute)
			go initJudge(accountName)
		}
	}

	if judge == false {
		time.Sleep(1 * time.Minute)
		resp := getReserveddresses()
		if len(resp.Addresses) > 0 {
			for _, address := range resp.Addresses {
				registerAddressOnForkscanner(address.ReserveAddress)
				reserveScript, _ := hex.DecodeString(address.ReserveScript)
				insertSweepAddress(address.ReserveAddress, reserveScript, nil, 0)
			}
		}
	}

	wg.Add(1)
	time.Sleep(1 * time.Minute)
	if judge == true {
		go startJudge(accountName)
	}

	time.Sleep(1 * time.Minute)
	startBridge(accountName, forkscanner_url)

	wg.Wait()

}
