package main

import (
	"bytes"
	"database/sql"
	"encoding/hex"
	"flag"
	"fmt"
	"os/exec"
	"strings"
	"sync"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"github.com/tyler-smith/go-bip32"
)

var dbconn *sql.DB
var masterPrivateKey *bip32.Key
var judge bool
var wg sync.WaitGroup

func initialize() {
	//config setup
	viper.AddConfigPath("./configs")
	viper.SetConfigName("config") // Register config file name (no extension)
	viper.SetConfigType("json")   // Look for specific type
	viper.ReadInConfig()

	//wallet setup
	new_wallet := flag.Bool("new_wallet", false, "set to true if you want to create a new wallet")
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

	// db connection
	dbconn = initDB()
	fmt.Println("DB initialized")
	btcPublicKey := hex.EncodeToString(masterPrivateKey.PublicKey().Key)
	accountName := fmt.Sprintf("%v", viper.Get("accountName"))

	command := fmt.Sprintf("nyksd keys show %s --bech val -a --keyring-backend test", accountName)
	args := strings.Fields(command)
	cmd := exec.Command(args[0], args[1:]...)

	valAddr, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	command = fmt.Sprintf("nyksd keys show %s -a --keyring-backend test", accountName)
	args = strings.Fields(command)
	cmd = exec.Command(args[0], args[1:]...)

	OrchAddr, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	// register delegate address

	command = fmt.Sprintf("nyksd tx nyks set-delegate-addresses %s %s %s --from %s --chain-id nyks --keyring-backend test -y", string(valAddr), string(OrchAddr), btcPublicKey, accountName)

	fmt.Println("command : ", command)
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

	// var addr = fmt.Sprintf("%v:%v", viper.Get("forkscanner_host"), viper.Get("forkscanner_ws_port"))
	// forkscanner_url := url.URL{Scheme: "ws", Host: addr, Path: "/"}

	wg.Add(1)
	// go orchestrator(accountName, forkscanner_url)
	// time.Sleep(1 * time.Minute)
	if accountName == "alice" {
		judge = true
		// go initJudge(accountName)
	}

	// time.Sleep(1 * time.Minute)
	go startJudge(accountName)

	wg.Wait()

}

// package main

// import (
// 	"database/sql"
// 	"fmt"
// 	"sync"

// 	"github.com/btcsuite/btcd/chaincfg"
// 	"github.com/btcsuite/btcd/txscript"
// 	"github.com/btcsuite/btcutil"
// 	"github.com/tyler-smith/go-bip32"
// )

// var dbconn *sql.DB
// var masterPrivateKey *bip32.Key
// var judge bool
// var wg sync.WaitGroup

// func main() {
// 	addressString := "bc1q85z24t6eq87ru6vp9kvt9uvuvzzx84w9kkhhhumjdda6vqfw2cxsfxfyj6"

// 	// Decode the Bitcoin address.
// 	address, err := btcutil.DecodeAddress(addressString, &chaincfg.MainNetParams)
// 	if err != nil {
// 		fmt.Println("Error decoding address:", err)
// 		return
// 	}

// 	// Generate the pay-to-address script.
// 	script, err := txscript.PayToAddrScript(address)
// 	if err != nil {
// 		fmt.Println("Error generating pay-to-address script:", err)
// 		return
// 	}

// 	fmt.Printf("Generated script: %x\n", script)
// }
