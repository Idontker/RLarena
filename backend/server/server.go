package server

import (
	"log"
	"net/http"
	"os"
)

// Map to store id-to-state mapping
// var gameState = GameState{
// 	Rows: 8,
// 	Cols: 8,
// 	History: []Turn{
// 		{DestRow: 1, DestCol: 1, SourceRow: 0, SourceCol: 0, Player: 1},
// 	},
// 	Board: [][]int{
// 		{0, 1, 1, 1, 0, 0, 0, 0},
// 		{0, 1, 0, 0, 0, 0, 2, 0},
// 		{0, 0, 0, 0, 0, 0, 0, 0},
// 		{0, 0, 0, 1, 0, 0, 0, 0},
// 		{0, 0, 0, 0, 0, 0, 0, 0},
// 		{0, 0, 0, 0, 0, 0, 2, 0},
// 		{0, 0, 0, 0, 0, 0, 0, 0},
// 		{0, 0, 0, 0, 0, 0, 0, 0},
// 	},
// }

var logger = log.New(os.Stdout, "[http]: ", log.LstdFlags)

func LogRequest(r *http.Request) {
	// logger.Printf("Request: %s %s", r.Method, r.URL.String())
}

func StartServer() {

	// paths: /game
	InitHttpHandler_Game_Handler()

	// paths: /user
	InitHttpHandler_Users()

	// paths: /match
	InitHttpHandler_Match_Making()

	InitHttpHandler_Frontend_Handler()

	// Start server
	logger.Println("Server is starting...")
	logger.Println("http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))

}
