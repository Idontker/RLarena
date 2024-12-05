package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
)

type LFGameStruct struct {
	Player      *Player
	LfGameCount int
}

var (
	LookingForMatch = make([]LFGameStruct, 0)
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

	gameCountStr := r.URL.Query().Get("gameCount")

	if gameCountStr == "" {
		http.Error(w, "Missing gameCount", http.StatusBadRequest)
		return
	}
	gameCount, err := strconv.Atoi(gameCountStr)
	if err != nil {
		http.Error(w, "Invalid gameCount (should be an integer)", http.StatusBadRequest)
		return
	}

	mutexLfMatch.Lock()

	if len(LookingForMatch) > 0 {
		lfGame := LookingForMatch[0]
		opponent := *lfGame.Player

		if opponent.ID == player.ID {
			slog.Debug("STILL looking for match", "player", player.ID)
			fmt.Fprintf(w, "Still looking for match")
			w.WriteHeader(http.StatusOK)
			return
		}

		var createNgames int
		if lfGame.LfGameCount < gameCount {
			// store the number of games to create
			createNgames = lfGame.LfGameCount

			LookingForMatch = make([]LFGameStruct, 1)

			newLfGame := LFGameStruct{
				Player:      &player,
				LfGameCount: gameCount - lfGame.LfGameCount,
			}
			LookingForMatch = append(LookingForMatch, newLfGame)
		} else {
			// store the number of games to create
			createNgames = gameCount
			LookingForMatch[0] = LFGameStruct{
				Player:      LookingForMatch[0].Player,
				LfGameCount: LookingForMatch[0].LfGameCount - gameCount,
			}
		}

		// Remove opponent from LookingForMatch if he is satisfied with the number of games
		if lfGame.LfGameCount == gameCount {
			LookingForMatch = make([]LFGameStruct, 0)
		}
		mutexLfMatch.Unlock()

		// Add game to the list of active games
		MutexActiveGames.Lock()
		// Create a new games
		for i := 0; i < createNgames; i++ {
			game := createGame(player, opponent)
			ActiveGames = append(ActiveGames, game)
		}
		MutexActiveGames.Unlock()

		w.WriteHeader(http.StatusOK)
		slog.Debug("Match found!", "playerOne", player.ID, "playerTwo", opponent.ID, "gameCount", gameCount)
		fmt.Fprintf(w, "Match found! Player %d vs Player %d (in total %d games)", player.ID, opponent.ID, createNgames)
	} else {
		newLfGame := LFGameStruct{
			Player:      &player,
			LfGameCount: gameCount,
		}

		LookingForMatch = append(LookingForMatch, newLfGame)
		mutexLfMatch.Unlock()
		slog.Debug("Looing for match", "player", player.ID, "gameCount", gameCount)
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
