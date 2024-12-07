package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"

	// _ "github.com/mattn/go-sqlite3"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Warning: Could not load .env file, using system environment variables")
	}

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
