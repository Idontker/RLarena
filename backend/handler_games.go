package main

import (
	"encoding/json"
	"log/slog"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
)

var (
	games            map[int]Game = make(map[int]Game)
	mutexGames       sync.Mutex
	ActiveGames      []Game
	MutexActiveGames sync.Mutex
	nextID           int = 2
	mutexNextID      sync.Mutex
)

type Game struct {
	ID        int         `json:"id"`
	Player1ID int         `json:"player1_id"`
	Player2ID int         `json:"player2_id"`
	Outcome   int         `json:"outcome"`    // -1: draw, 0: ongoing, 1: win playerOne, 2: win playerTwo
	GameState *GameState  `json:"game_state"` // Additional field to store the state of the game
	mutex     *sync.Mutex `json:"-"`
}

func createGame(p1 Player, p2 Player) Game {
	var rows, cols int
	switch rand.Intn(4) {
	case 0:
		rows, cols = 3, 3
	case 1:
		rows, cols = 5, 3
	case 2:
		rows, cols = 8, 8
	case 3:
		rows, cols = 16, 16
	}

	// create a board
	board := make([][]int, rows)
	for i := range board {
		board[i] = make([]int, cols)
	}

	// setup players
	for i := 0; i < cols; i++ {
		board[0][i] = 1      // Player 1
		board[rows-1][i] = 2 // Player 2
	}

	history := []Turn{}

	state := GameState{
		Rows:    rows,
		Cols:    cols,
		History: history,
		Board:   board,
	}

	var id1, id2 int
	switch rand.Intn(2) {
	case 0:
		id1, id2 = p1.ID, p2.ID
	case 1:
		id2, id1 = p1.ID, p2.ID
	}

	mutexNextID.Lock()
	defer mutexNextID.Unlock()

	game := Game{
		ID:        nextID,
		Player1ID: id1,
		Player2ID: id2,
		Outcome:   0, // Game is ongoing
		GameState: &state,
		mutex:     &sync.Mutex{},
	}
	// tbh this should Lock is not nessesary
	mutexGames.Lock()
	games[nextID] = game
	mutexGames.Unlock()
	nextID++
	return game
}

func serveGameState(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")
	w.Header().Set("Content-Type", "application/json")

	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
	}

	game, exists := games[id]
	if !exists {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	slog.Debug("game pulled", "id", id, "game", game)
	game.GameState.Display()

	json.NewEncoder(w).Encode(game)
}

func applyAction(player Player, gameId int, action Turn) (bool, string) {
	game, exists := games[gameId]
	if !exists {
		return false, "game not found"
	}
	if (game.GameState.NextPlayer() == 1 && game.Player1ID != player.ID) ||
		(game.GameState.NextPlayer() == 2 && game.Player2ID != player.ID) {
		return false, "not your turn"
	}

	game.mutex.Lock()
	defer game.mutex.Unlock()

	valid := game.GameState.applyAction(action)
	if !valid {
		return false, "invalid action (against the rules)"
	}

	// update game outcome
	if game.GameState.IsEnd() {
		game.Outcome = game.GameState.GetWinner()

		// TODO: there is a minor race condition here, as the elo update is not synchronized
		// with the game history update. Therefore, the game history might be not in the
		// correct order, when comparing it with the elo gains/losses.
		// But, the elo calculations are correct.
		// Therefore, this is not a big issue and will be ignored for now
		_, e1, e2 := UpdateElo(game.Player1ID, game.Player2ID, game.Outcome)
		// assert.True(success, "Elo should be updated")

		player1 := players[game.Player1ID]
		// assert.True(ok1, "Player 1 id of a game should lead to a player")
		player2 := players[game.Player2ID]
		// assert.True(ok2, "Player 1 id of a game should lead to a player")

		mutex_p1 := getPlayerMutex(game.Player1ID)
		mutex_p2 := getPlayerMutex(game.Player2ID)

		mutex_p1.Lock()
		player1.GameHistory = append(player1.GameHistory, HistoryEntry{
			GameID: game.ID,
			Win:    game.Outcome == 1,
			Draw:   game.Outcome == -1,
			Loss:   game.Outcome == 2,
			Elo:    e1,
		})
		players[game.Player1ID] = player1
		mutex_p1.Unlock()

		mutex_p2.Lock()
		player2.GameHistory = append(player2.GameHistory, HistoryEntry{
			GameID: game.ID,
			Win:    game.Outcome == 2,
			Draw:   game.Outcome == -1,
			Loss:   game.Outcome == 1,
			Elo:    e2,
		})
		players[game.Player2ID] = player2
		mutex_p2.Unlock()

		// remove game from active games
		MutexActiveGames.Lock()
		for i, g := range ActiveGames {
			if g.ID == gameId {
				ActiveGames = append(ActiveGames[:i], ActiveGames[i+1:]...)
				break
			}
		}
		MutexActiveGames.Unlock()
	}

	// save the history and outcome
	mutexGames.Lock()
	games[gameId] = game
	mutexGames.Unlock()

	return true, ""

}

func servePerformActionBulk(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Token is required for authorization.", http.StatusUnauthorized)
		return
	}

	// get player
	player, err := GetPlayerByToken(token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
	}

	var actions []struct {
		GameID int  `json:"gameId"`
		Action Turn `json:"action"`
	}

	err = json.NewDecoder(r.Body).Decode(&actions)
	if err != nil {
		http.Error(w, "Invalid JSON input", http.StatusBadRequest)
		return
	}

	msgs := make(map[int]string, 0)

	for _, a := range actions {
		success, msg := applyAction(player, a.GameID, a.Action)
		if !success {
			msgs[a.GameID] = msg
		}
	}

}

func servePerformAction(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")
	id, err := strconv.Atoi(idString)

	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Token is required for authorization.", http.StatusUnauthorized)
		return
	}

	// get player
	player, err := GetPlayerByToken(token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
	}

	var action Turn
	err = json.NewDecoder(r.Body).Decode(&action)
	if err != nil {
		http.Error(w, "Invalid action (Parsing errors)", http.StatusBadRequest)
		return
	}

	success, msg := applyAction(player, id, action)
	if !success {
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func serveActiveGames(w http.ResponseWriter, _ *http.Request) {
	MutexActiveGames.Lock()
	defer MutexActiveGames.Unlock()

	w.Header().Set("Content-Type", "application/json")
	// TODO maybe only return the ID and the player names
	json.NewEncoder(w).Encode(ActiveGames)
}

func serveActiveGamesUser(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("userToken")
	if token == "" {
		http.Error(w, "Token is required for authorization.", http.StatusUnauthorized)
		return
	}

	// get player
	player, err := GetPlayerByToken(token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
	}

	myturn := make([]Game, 0)
	awating := make([]Game, 0)

	MutexActiveGames.Lock()

	for _, game := range ActiveGames {
		if (game.Player1ID == player.ID && game.GameState.NextPlayer() == 1) || (game.Player2ID == player.ID && game.GameState.NextPlayer() == 2) {
			myturn = append(myturn, game)
		} else if game.Player1ID == player.ID || game.Player2ID == player.ID {
			awating = append(awating, game)
		}
	}
	defer MutexActiveGames.Unlock()

	w.Header().Set("Content-Type", "application/json")
	// TODO maybe only return the ID and the player names
	json.NewEncoder(w).Encode(struct {
		MyTurn   []Game `json:"my_turn"`
		Awaiting []Game `json:"awaiting"`
	}{
		MyTurn:   myturn,
		Awaiting: awating,
	},
	)
}

func InitHttpHandler_Game_Handler() {

	http.HandleFunc("GET /game/{id}/state", func(w http.ResponseWriter, r *http.Request) {
		LogRequest(r)
		serveGameState(w, r)
	})

	http.HandleFunc("POST /game/{id}/action", func(w http.ResponseWriter, r *http.Request) {
		LogRequest(r)
		servePerformAction(w, r)
	})

	http.HandleFunc("POST /games/actions", func(w http.ResponseWriter, r *http.Request) {
		LogRequest(r)
		servePerformActionBulk(w, r)
	})

	http.HandleFunc("GET /games/active", func(w http.ResponseWriter, r *http.Request) {
		LogRequest(r)
		serveActiveGames(w, r)
	})

	http.HandleFunc("GET /games/active/{userToken}", func(w http.ResponseWriter, r *http.Request) {
		LogRequest(r)
		serveActiveGamesUser(w, r)
	})

}
