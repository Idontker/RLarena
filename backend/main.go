package main

import (
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/joho/godotenv"

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
	// InitHttpHandler_Match_Making()

	InitHttpHandler_Frontend_Handler()

	// Start periodic job to ensure games are running
	go runPeriodicJob()

	// Start server
	log.Println("Server is starting...")
	log.Println("http://localhost:8081")
	log.Println(http.ListenAndServe(":8081", nil))
}

func runPeriodicJob() {
	// ticker := time.NewTicker(1 * time.Minute)  // Run every minute
	ticker := time.NewTicker(10 * time.Second) // Run every 30s
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := ensureGamesAreRunning()
			if err != nil {
				slog.Error("Error ensuring games are running (regular)", "error", err)
			}
		}
	}
}
