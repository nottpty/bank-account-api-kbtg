package main

import (
	"log"

	"bank-account-api-kbtg/server"
)

func main() {
	server.ConnectDB()
	log.Fatal(startServer())
}
