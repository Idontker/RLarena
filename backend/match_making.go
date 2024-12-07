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

func serveLookingForMatch(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")

	player, err := DB_Get_Player_by_Token(token)
	if err != nil {
		http.Error(w, "Invalid token (error:"+err.Error()+")", http.StatusUnauthorized)
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
				Player:      player,
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
		// Create a new games
		created := 0
		msgs := make([]string, 0)
		for i := 0; i < createNgames; i++ {
			_, err := createGame(*player, opponent)
			if err != nil {
				msgs = append(msgs, err.Error())
			} else {
				created++
			}
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(struct {
			Created int      `json:"created"`
			Errors  []string `json:"errors"`
		}{
			Created: created,
			Errors:  msgs,
		})
		slog.Debug("Match found!", "playerOne", player.ID, "playerTwo", opponent.ID, "gameCount", gameCount)

	} else {
		newLfGame := LFGameStruct{
			Player:      player,
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

}
