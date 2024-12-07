package main

import (
	"log"
	"net/http"

	// _ "github.com/mattn/go-sqlite3"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

func main() {
	// server.StartServer()

	// paths: /game
	InitHttpHandler_Game_Handler()

	// paths: /user
	InitHttpHandler_Users()

	// paths: /match
	InitHttpHandler_Match_Making()

	InitHttpHandler_Frontend_Handler()

	// Start server
	log.Println("Server is starting...")
	log.Println("http://localhost:8081")
	log.Println(http.ListenAndServe(":8081", nil))

}
