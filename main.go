package main

import (
	"database/sql"
	"flag"
	"fmt"
	"net/url"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

var dbconn *sql.DB

func main() {

	viper.AddConfigPath("./configs")
	viper.SetConfigName("config") // Register config file name (no extension)
	viper.SetConfigType("json")   // Look for specific type
	viper.ReadInConfig()

	accountName := fmt.Sprintf("%v", viper.Get("accountName"))
	dbconn = init_db()

	var addr = flag.String("addr", fmt.Sprintf("%v:%d", viper.Get("forkscanner_host"), viper.Get("forkscanner_ws_port")), "http service address")
	flag.Parse()
	forkscanner_url := url.URL{Scheme: "ws", Host: *addr, Path: "/"}

	go start_judge(accountName)
	orchestrator(accountName, forkscanner_url)
}
