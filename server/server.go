package server

import (
	"database/sql"
	"log"
)

var db *sql.DB

func ConnectDB() {
	var err error
	const connStr = "admin:@tcp(127.0.0.1:3306)/bank_account_api?parseTime=true"
	db, err = sql.Open("mysql", connStr)
	if err != nil {
		log.Fatal(err)
	}
}

// func startServer() error {}
