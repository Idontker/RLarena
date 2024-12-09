package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
)

const K = 32 // constant for Elo calculation

type HistoryEntry struct {
	GameID int  `json:"id"`
	Win    bool `json:"win"`
	Draw   bool `json:"draw"`
	Loss   bool `json:"loss"`
	Elo    int  `json:"elo"`
}

type Player struct {
	ID          int            `json:"id"`
	Name        string         `json:"name"`
	CurrentElo  int            `json:"current_elo"`
	GameHistory []HistoryEntry `json:"game_history"`
	SecretToken string         `json:"-"`
}

// func (p Player) MarshalJSON() ([]byte, error) {
// 	type Alias Player

// 	history = make([]HistoryEntry, len(p.GameHistory))

// }

func CalculateEloUpdate(current_elo1 int, current_elo2 int, outcome int) (bool, int, int) {
	e1 := 1 / (1. + math.Pow(10, (float64(current_elo2)-float64(current_elo1))/400))
	e2 := 1 / (1. + math.Pow(10, (float64(current_elo1)-float64(current_elo2))/400))

	var s1, s2 float64
	if outcome == 1 {
		s1, s2 = 1, 0
	} else if outcome == -1 {
		s1, s2 = 0.5, 0.5
	} else if outcome == 2 {
		s1, s2 = 0, 1
	} else {
		fmt.Println("[ERROR]: this outcome is impossible.")
		return false, 0, 0
	}

	new_elo1 := current_elo1 + int(float64(K)*(s1-float64(e1)))
	new_elo2 := current_elo2 + int(float64(K)*(s2-float64(e2)))

	return true, new_elo1, new_elo2
}

func serveDisplayPlayers(w http.ResponseWriter, _ *http.Request) {
	players, err := DB_Get_Players()

	if err != nil {
		http.Error(w, "Error fetching players", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(players)
}

func serveSignUp(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}
	token, err := DB_Create_Player(name)

	if err != nil {
		http.Error(w, "Error creating player", http.StatusInternalServerError)
		return
	}

	err = ensureGamesAreRunning()
	if err != nil {
		slog.Error("Error ensuring games are running", "error", err)
	}

	response := map[string]string{"token": token}
	json.NewEncoder(w).Encode(response)
}

func serveGetPlayerByToken(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	if token == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}

	player, err := DB_Get_Player_by_Token(token)
	if err != nil {
		msg := fmt.Sprint("Player not found (%s)", err.Error())
		http.Error(w,
			msg,
			http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(player)
}

func InitHttpHandler_Users() {
	http.HandleFunc("GET /users", func(w http.ResponseWriter, r *http.Request) {
		LogRequest(r)
		serveDisplayPlayers(w, r)
	})

	http.HandleFunc("GET /user/signup", func(w http.ResponseWriter, r *http.Request) {
		LogRequest(r)
		serveSignUp(w, r)
	})

	http.HandleFunc("GET /user/{token}", func(w http.ResponseWriter, r *http.Request) {
		LogRequest(r)
		serveGetPlayerByToken(w, r)
	})
}
