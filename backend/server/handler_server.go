package server

// import (
// 	"fmt"
// 	"net/http"
// 	"strconv"
// )

// func lookUpGame(w http.ResponseWriter, r *http.Request) (GameState, error) {
// 	idStr := r.URL.Query().Get("id")
// 	id, err := strconv.Atoi(idStr)

// 	if err != nil {
// 		http.Error(w, "Invalid ID", http.StatusBadRequest)
// 		return GameState{}, err
// 	}

// 	state, ok := stateMap[id]
// 	if !ok {
// 		http.Error(w, "Game not found", http.StatusNotFound)
// 		return GameState{}, fmt.Errorf("game with ID %d not found", id)
// 	}
// 	return state, nil

// }

// func serveCreateGame(w http.ResponseWriter, r *http.Request) {
// 	id := len(stateMap) + 1
// 	rows, err := strconv.Atoi(r.URL.Query().Get("rows"))
// 	if err != nil {
// 		log.Println("Invalid rows parameter", rows, err)
// 		http.Error(w, "Invalid rows parameter", http.StatusBadRequest)
// 		return
// 	}
// 	cols, err := strconv.Atoi(r.URL.Query().Get("cols"))
// 	if err != nil {
// 		log.Println("Invalid cols parameter", cols, err)
// 		http.Error(w, "Invalid cols parameter", http.StatusBadRequest)
// 		return
// 	}

// 	// create a board
// 	board := make([][]int, rows)
// 	for i := range board {
// 		board[i] = make([]int, cols)
// 	}

// 	// setup players
// 	for i := 0; i < cols; i++ {
// 		board[0][i] = 1      // Player 1
// 		board[rows-1][i] = 2 // Player 2
// 	}

// 	history := []Turn{}

// 	state := GameState{
// 		Rows:    rows,
// 		Cols:    cols,
// 		History: history,
// 		Board:   board,
// 	}
// 	stateMap[id] = state
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(state)
// }

// func serveState(w http.ResponseWriter, r *http.Request) {
// 	state, err := lookUpGame(w, r)
// 	if err != nil {
// 		return
// 	}
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(state)
// }

// func serveGameOver(w http.ResponseWriter, r *http.Request) {
// 	state, err := lookUpGame(w, r)
// 	if err != nil {
// 		return
// 	}

// 	gameOver := state.IsEnd()
// 	winner := state.GetWinner()
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(map[string]interface{}{
// 		"gameOver": gameOver,
// 		"winner":   winner,
// 	})
// }

// func serveGetMoves(w http.ResponseWriter, r *http.Request) {
// 	state, err := lookUpGame(w, r)
// 	if err != nil {
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(state.PossibleMoves())

// }

// func servePerformAction(w http.ResponseWriter, r *http.Request) {
// 	idStr := r.URL.Query().Get("id")
// 	id, err := strconv.Atoi(idStr)

// 	if err != nil {
// 		http.Error(w, "Invalid ID", http.StatusBadRequest)
// 		return
// 	}
// 	state, ok := stateMap[id]
// 	if !ok {
// 		http.Error(w, "Game not found", http.StatusNotFound)
// 		return
// 	}

// 	var action Turn
// 	err = json.NewDecoder(r.Body).Decode(&action)
// 	if err != nil {
// 		http.Error(w, "Invalid action (Parsing errors)", http.StatusBadRequest)
// 		return
// 	}

// 	valid := state.applyAction(action)
// 	if !valid {
// 		http.Error(w, "Invalid action (Against the rules)", http.StatusBadRequest)
// 	}
// 	stateMap[id] = state
// 	fmt.Println("(after2) history", state.History)
// 	fmt.Println("(after3) history", stateMap[id].History)

// }
