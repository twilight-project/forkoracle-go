package main

import (
	"database/sql"
	"flag"
	"fmt"
	"net/url"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

func main() {

	viper.AddConfigPath("./configs")
	viper.SetConfigName("config") // Register config file name (no extension)
	viper.SetConfigType("json")   // Look for specific type
	viper.ReadInConfig()

	accountName := fmt.Sprintf("%v", viper.Get("accountName"))

	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", viper.Get("DB_host"), viper.Get("DB_port"), viper.Get("DB_user"), viper.Get("DB_password"), viper.Get("DB_name"))
	db, err := sql.Open("postgres", psqlconn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	var addr = flag.String("addr", fmt.Sprintf("%v:%d", viper.Get("forkscanner_host"), viper.Get("forkscanner_ws_port")), "http service address")
	flag.Parse()
	forkscanner_url := url.URL{Scheme: "ws", Host: *addr, Path: "/"}

	orchestrator(accountName, forkscanner_url, db)
}
