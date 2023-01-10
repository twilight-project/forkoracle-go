package main

import (
	"database/sql"
	"fmt"

	"github.com/spf13/viper"
)

func initDB() *sql.DB {
	psqlconn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", viper.Get("DB_host"), viper.Get("DB_port"), viper.Get("DB_user"), viper.Get("DB_password"), viper.Get("DB_name"))
	db, err := sql.Open("postgres", psqlconn)
	if err != nil {
		panic(err)
	}
	return db
}

func insertNotifications(element WatchtowerNotification) {

	fmt.Println("inside insert DB")
	_, err := dbconn.Exec("INSERT into notification VALUES ($1, $2, $3, $4, $5, $6, $7)",
		element.Block,
		element.Receiving,
		element.Satoshis,
		element.Height,
		element.Receiving_txid,
		false,
		element.Sending,
	)
	if err != nil {
		fmt.Println("An error occured while executing query: ", err)
	}
}

func markProcessedNotifications(element WatchtowerNotification) {

	_, err := dbconn.Exec("update notification set archived = true where txid = $1 and sending = $2",
		element.Receiving_txid,
		element.Sending,
	)
	if err != nil {
		fmt.Println("An error occured while executing query: ", err)
	}
}

func queryNotification() []WatchtowerNotification {
	fmt.Println("inside query notification")
	DB_reader, err := dbconn.Query("select * from notification where archived = false")
	if err != nil {
		fmt.Println("An error occured while executing query: ", err)
	}

	defer DB_reader.Close()
	addresses := make([]WatchtowerNotification, 0)

	for DB_reader.Next() {
		address := WatchtowerNotification{}
		err := DB_reader.Scan(
			&address.Block,
			&address.Receiving,
			&address.Satoshis,
			&address.Height,
			&address.Receiving_txid,
			&address.Archived,
			&address.Sending,
		)
		if err != nil {
			fmt.Println(err)
		}
		addresses = append(addresses, address)
	}
	return addresses
}
