package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
)

var (
	LookingForMatch = make([]Player, 0)
	mutexLfMatch    sync.Mutex
)

func serveOpenMatchesByPlayer(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")

	player, err := GetPlayerByToken(token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	MutexActiveGames.Lock()
	defer MutexActiveGames.Unlock()
	myGamesIds := make([]int, 0)
	for _, game := range ActiveGames {
		if game.Player1ID == player.ID || game.Player2ID == player.ID {
			myGamesIds = append(myGamesIds, game.ID)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(myGamesIds)
}

func serveLookingForMatch(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")

	player, err := GetPlayerByToken(token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	mutexLfMatch.Lock()

	if len(LookingForMatch) > 0 {
		opponent := LookingForMatch[0]
		if opponent.ID == player.ID {

			slog.Debug("STILL looking for match", "player", player.ID)
			fmt.Fprintf(w, "Still looking for match")
			w.WriteHeader(http.StatusOK)
			return
		}

		// Remove opponent from LookingForMatch
		LookingForMatch = make([]Player, 0)
		mutexLfMatch.Unlock()

		// Create a new game
		game := createGame(player, opponent)

		// Add game to the list of active games (assuming you have a list of active games)
		MutexActiveGames.Lock()

		ActiveGames = append(ActiveGames, game)
		MutexActiveGames.Unlock()

		w.WriteHeader(http.StatusOK)
		slog.Debug("Match found!", "playerOne", player.ID, "playerTwo", opponent.ID)
		fmt.Fprintf(w, "Match found! Player %d vs Player %d", player.ID, opponent.ID)
	} else {
		LookingForMatch = append(LookingForMatch, player)
		mutexLfMatch.Unlock()
		slog.Debug("Looing for match", "player", player.ID)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Looking for match")
	}
}

func InitHttpHandler_Match_Making() {
	http.HandleFunc("GET /match/queueup/{token}", func(w http.ResponseWriter, r *http.Request) {
		LogRequest(r)
		serveLookingForMatch(w, r)
	})

	http.HandleFunc("GET /match/my/{token}", func(w http.ResponseWriter, r *http.Request) {
		LogRequest(r)
		serveOpenMatchesByPlayer(w, r)
	})
}
