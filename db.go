package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/spf13/viper"
)

func init_db() *sql.DB {
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", viper.Get("DB_host"), viper.Get("DB_port"), viper.Get("DB_user"), viper.Get("DB_password"), viper.Get("DB_name"))
	db, err := sql.Open("postgres", psqlconn)
	if err != nil {
		panic(err)
	}
	return db
}

func insert_notifications(element WatchtowerNotification) {

	_, err := dbconn.Exec("INSERT into notification VALUES ($1, $2, $3, $4, $5)",
		element.Block,
		element.Receiving,
		element.Satoshis,
		element.Height,
		element.Txid,
	)
	if err != nil {
		log.Fatalf("An error occured while executing query: %v", err)
	}
}

func query_notification() []WatchtowerNotification {
	DB_reader, err := dbconn.Query("select * from notification where archived = false")
	if err != nil {
		log.Fatalf("An error occured while executing query: %v", err)
	}
	defer DB_reader.Close()
	addresses := make([]WatchtowerNotification, 0)

	for DB_reader.Next() {
		address := WatchtowerNotification{}
		err := DB_reader.Scan(
			address.Block,
			address.Receiving,
			address.Satoshis,
			address.Height,
			address.Txid,
			address.archived,
		)

		if err != nil {
			fmt.Println(err)
		}
		addresses = append(addresses, address)
	}
	return addresses
}
