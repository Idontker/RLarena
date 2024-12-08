package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"strconv"
)

var PAGE_SIZE = 10

type Game struct {
	ID        int        `json:"id"`
	Player1ID int        `json:"player1_id"`
	Player2ID int        `json:"player2_id"`
	Outcome   int        `json:"outcome"`    // -1: draw, 0: ongoing, 1: win playerOne, 2: win playerTwo
	GameState *GameState `json:"game_state"` // Additional field to store the state of the game
}

func createGame(p1 Player, p2 Player) (*Game, error) {
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

	var id1, id2 int
	switch rand.Intn(2) {
	case 0:
		id1, id2 = p1.ID, p2.ID
	case 1:
		id2, id1 = p1.ID, p2.ID
	}

	id, err := DB_Create_Game(id1, id2, rows, cols)
	if err != nil {
		slog.Error("Error creating game", "error", err)
		return nil, err
	}
	game, err := DB_Get_Game(id)
	if err != nil {
		slog.Error("Error creating game", "id", id, "error", err)
		return nil, err
	}

	return game, err
}

func serveGames(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	if pageStr == "" {
		pageStr = "1"
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		http.Error(w, "Invalid page number", http.StatusBadRequest)
		return
	}

	startIdx := (page - 1) * PAGE_SIZE
	endIdx := page * PAGE_SIZE

	games, err := DB_Get_Games(startIdx, endIdx)
	if err != nil {
		slog.Error("Error getting game state", "page", page, "error", err)
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(games)
}

func serveGameState(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")
	w.Header().Set("Content-Type", "application/json")

	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
	}

	game, err := DB_Get_Game(id)
	if err != nil {
		slog.Error("Error getting game state", "id", id, "error", err)
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	slog.Debug("game pulled", "id", id, "game", game)
	game.GameState.Display()

	json.NewEncoder(w).Encode(game)
}

func applyAction(player Player, gameId int, action Turn) (bool, string) {
	game, err := DB_Get_Game(gameId)
	if err != nil {
		return false, err.Error()
	}

	if (game.GameState.NextPlayer() == 1 && game.Player1ID != player.ID) ||
		(game.GameState.NextPlayer() == 2 && game.Player2ID != player.ID) {
		return false, "not your turn"
	}

	valid := game.GameState.applyAction(action)
	if !valid {
		return false, "invalid action (against the rules)"
	}

	// update game outcome
	if game.GameState.IsEnd() {
		game.Outcome = game.GameState.GetWinner()
	}

	err = DB_apply_action(action, game)
	if err != nil {
		return false, err.Error()
	}

	if game.GameState.IsEnd() {
		// update player elo and game histories
		hist1 := &HistoryEntry{
			GameID: game.ID,
			Win:    game.Outcome == 1,
			Draw:   game.Outcome == -1,
			Loss:   game.Outcome == 2,
			Elo:    0,
		}
		hist2 := &HistoryEntry{
			GameID: game.ID,
			Win:    game.Outcome == 2,
			Draw:   game.Outcome == -1,
			Loss:   game.Outcome == 1,
			Elo:    0,
		}
		DB_update_Elo_and_History(game.Player1ID, game.Player2ID, game.Outcome, hist1, hist2)
	}

	return true, ""

}

func servePerformActionBulk(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Token is required for authorization.", http.StatusUnauthorized)
		return
	}

	// get player
	player, err := DB_Get_Player_by_Token(token)
	if err != nil {
		msg := fmt.Sprint("Invalid token (%s)", err.Error())
		http.Error(w, msg, http.StatusUnauthorized)
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
		success, msg := applyAction(*player, a.GameID, a.Action)
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
	player, err := DB_Get_Player_by_Token(token)
	if err != nil {
		msg := fmt.Sprint("Invalid token (%s)", err.Error())
		http.Error(w, msg, http.StatusUnauthorized)
	}

	var action Turn
	err = json.NewDecoder(r.Body).Decode(&action)
	if err != nil {
		http.Error(w, "Invalid action (Parsing errors)", http.StatusBadRequest)
		return
	}

	success, msg := applyAction(*player, id, action)
	if !success {
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func serveActiveGames(w http.ResponseWriter, _ *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	activeGames, err := DB_Get_Active_Games()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(activeGames)
}

func serveActiveGamesUser(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("userToken")
	if token == "" {
		http.Error(w, "Token is required for authorization.", http.StatusUnauthorized)
		return
	}

	// get player
	player, err := DB_Get_Player_by_Token(token)
	if err != nil {
		msg := fmt.Sprint("Invalid token (%s)", err.Error())
		http.Error(w, msg, http.StatusUnauthorized)
		return
	}

	myturn := make([]Game, 0)
	awating := make([]Game, 0)

	activeGames, err := DB_Get_Active_Games_By_Player(player)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, game := range activeGames {
		if (game.Player1ID == player.ID && game.GameState.NextPlayer() == 1) || (game.Player2ID == player.ID && game.GameState.NextPlayer() == 2) {
			myturn = append(myturn, game)
		} else {
			awating = append(awating, game)
		}
	}

	w.Header().Set("Content-Type", "application/json")
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

	http.HandleFunc("GET /games/all", func(w http.ResponseWriter, r *http.Request) {
		LogRequest(r)
		serveGames(w, r)
	})

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
