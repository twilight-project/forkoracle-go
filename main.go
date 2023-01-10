package main

import (
	"database/sql"
	"flag"
	"fmt"
	"net/url"
	"time"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

var dbconn *sql.DB

func main() {

	viper.AddConfigPath("/testnet/btc-oracle/configs")
	viper.SetConfigName("config") // Register config file name (no extension)
	viper.SetConfigType("json")   // Look for specific type
	viper.ReadInConfig()

	accountName := fmt.Sprintf("%v", viper.Get("accountName"))
	dbconn = initDB()

	var addr = flag.String("addr", fmt.Sprintf("%v:%v", viper.Get("forkscanner_host"), viper.Get("forkscanner_ws_port")), "http service address")
	flag.Parse()
	forkscanner_url := url.URL{Scheme: "ws", Host: *addr, Path: "/"}

	go startJudge(accountName)
	time.Sleep(1 * time.Minute)
	orchestrator(accountName, forkscanner_url)
}
