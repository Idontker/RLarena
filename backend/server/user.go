package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

const K = 32

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

var playerMutexes = make(map[int]*sync.Mutex)

func getPlayerMutex(playerID int) *sync.Mutex {
	if _, exists := playerMutexes[playerID]; !exists {
		playerMutexes[playerID] = &sync.Mutex{}
	}
	return playerMutexes[playerID]
}

var (
	players          = make(map[int]Player)
	playerTokensToID = make(map[string]int)
	playerNames      = make(map[string]bool)
	mutex            sync.Mutex
	nextPlayerID     = 1
)

func GetPlayerByToken(token string) (Player, error) {
	mutex.Lock()
	defer mutex.Unlock()

	playerID, exists := playerTokensToID[token]
	if !exists {
		return Player{}, errors.New("player not found")
	}
	player := players[playerID]

	return player, nil
}

func generateToken() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 32)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}

func UpdateElo(playerOneId int, playerTwoId int, outcome int) (bool, int, int) {
	playerOneMutex := getPlayerMutex(playerOneId)
	playerTwoMutex := getPlayerMutex(playerTwoId)

	playerOneMutex.Lock()
	defer playerOneMutex.Unlock()

	playerTwoMutex.Lock()
	defer playerTwoMutex.Unlock()

	playerOne, playerOneExists := players[playerOneId]
	playerTwo, playerTwoExists := players[playerTwoId]

	if !playerOneExists || !playerTwoExists {
		if !playerOneExists {
			fmt.Println("[ERROR]: playerOne not found. Id:", playerOneId)
		}
		if !playerTwoExists {
			fmt.Println("[ERROR]: playerTwo not found. Id:", playerTwoId)
		}
		return false, 0, 0
	}

	e1 := 1 / (1. + math.Pow(10, (float64(playerTwo.CurrentElo)-float64(playerOne.CurrentElo))/400))
	e2 := 1 / (1. + math.Pow(10, (float64(playerOne.CurrentElo)-float64(playerTwo.CurrentElo))/400))

	var s1, s2 float64
	if outcome == 1 {
		s1, s2 = 1, 0
	} else if outcome == -1 {
		s1, s2 = 0.5, 0.5
	} else if outcome == 2 {
		s1, s2 = 0, 1
	} else {
		fmt.Println("[ERROR]: this outcome is impossible. Match between", playerOneId, playerTwoId)
		return false, 0, 0
	}

	playerOne.CurrentElo = playerOne.CurrentElo + int(float64(K)*(s1-float64(e1)))
	playerTwo.CurrentElo = playerTwo.CurrentElo + int(float64(K)*(s2-float64(e2)))

	players[playerOneId] = playerOne
	players[playerTwoId] = playerTwo

	return true, playerOne.CurrentElo, playerTwo.CurrentElo
}

func serveDisplayPlayers(w http.ResponseWriter, _ *http.Request) {
	mutex.Lock()

	var playerList []Player
	for _, player := range players {
		playerList = append(playerList, player)
	}

	mutex.Unlock()

	json.NewEncoder(w).Encode(playerList)
}

func serveSignUp(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	if _, exists := playerNames[name]; exists {
		http.Error(w, "Name already taken", http.StatusBadRequest)
		return
	}

	token := generateToken()
	player := Player{
		ID:          nextPlayerID,
		Name:        name,
		CurrentElo:  1000,
		GameHistory: []HistoryEntry{},
		SecretToken: token,
	}

	players[nextPlayerID] = player
	playerTokensToID[token] = nextPlayerID
	playerNames[name] = true
	nextPlayerID++

	response := map[string]string{"token": token}
	json.NewEncoder(w).Encode(response)
}

func serveGetPlayerByToken(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	if token == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}
	player, err := GetPlayerByToken(token)
	if err != nil {
		http.Error(w, "Player not found", http.StatusUnauthorized)
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
