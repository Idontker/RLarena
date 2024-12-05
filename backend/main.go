package main

import (
	"log"
	"net/http"
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
	log.Fatal(http.ListenAndServe(":8081", nil))
}
