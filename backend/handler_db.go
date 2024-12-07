package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"math/rand"
	"time"
)

type DB_Game struct {
	ID        int
	Player1ID int
	Player2ID int
	Outcome   int
	Rows      int
	Cols      int
}

type DB_Turn struct {
	ID        int
	TurnID    int
	GameID    int
	DestRow   int
	DestCol   int
	SourceRow int
	SourceCol int
	PlayerNum int
}

type DB_Player struct {
	ID          int
	Name        string
	SecretToken string
	CurrentElo  int
}

var db *sql.DB

func Db_open() (*sql.DB, error) {
	return sql.Open("sqlite3", "./app.db")
}

// ------------------------------
// Game Functions
// ------------------------------

func DB_Create_Game(player1_id int, player2_id int, rows int, cols int) (int, error) {
	db, err := Db_open()
	if err != nil {
		return -1, err
	}
	defer db.Close()

	row := db.QueryRow("INSERT INTO Games (Player1ID, Player2ID, Outcome, Rows, Cols) VALUES (?, ?, ?, ?)",
		player1_id,
		player2_id,
		0,
		rows,
		cols)
	var id int
	err = row.Scan(&id)
	if err != nil {
		return -1, err
	}

	return id, nil
}

func reconstruct_game(db *sql.DB, db_game DB_Game) (*Game, error) {
	rows := db_game.Rows
	cols := db_game.Cols

	// load turns
	turnResults, err := db.Query("SELECT * FROM Turns WHERE ID = ?", db_game.ID)
	if err != nil {
		return nil, err
	}

	history := []Turn{}
	for turnResults.Next() {
		db_turn := DB_Turn{}
		err = turnResults.Scan(&db_turn.ID, &db_turn.GameID, &db_turn.DestRow, &db_turn.DestCol, &db_turn.SourceRow, &db_turn.SourceCol, &db_turn.PlayerNum)
		if err != nil {
			return nil, err
		}
		turn := Turn{
			TurnID:    db_turn.ID,
			DestRow:   db_turn.DestRow,
			DestCol:   db_turn.DestCol,
			SourceRow: db_turn.SourceRow,
			SourceCol: db_turn.SourceCol,
			Player:    db_turn.PlayerNum,
		}
		history = append(history, turn)
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

	state := GameState{
		Rows:    rows,
		Cols:    cols,
		History: history,
		Board:   board,
	}

	// apply all turns
	for _, turn := range history {
		valid := state.applyAction(turn)
		if !valid {
			return nil, fmt.Errorf("invalid turn within game history")
		}
	}

	game := Game{
		ID:        db_game.ID,
		Player1ID: db_game.Player1ID,
		Player2ID: db_game.Player2ID,
		Outcome:   db_game.Outcome,
		GameState: &state,
	}

	return &game, nil
}

func DB_Get_Game(id int) (*Game, error) {
	db, err := Db_open()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// load game data
	row := db.QueryRow("SELECT * FROM Games WHERE ID = ?", id)
	db_game := DB_Game{}
	err = row.Scan(&db_game.ID, &db_game.Player1ID, &db_game.Player2ID, &db_game.Outcome, &db_game.Rows, &db_game.Cols)
	if err != nil {
		return nil, err
	}

	game, err := reconstruct_game(db, db_game)
	return game, nil

}

func DB_apply_action(action Turn, game *Game) error {
	db, err := Db_open()
	if err != nil {
		return err
	}
	defer db.Close()

	transaction, err := db.Begin()
	if err != nil {
		return err
	}

	// add turn
	_, err = transaction.Exec("INSERT INTO Turns (GameID, DestRow, DestCol, SourceRow, SourceCol, PlayerNum) VALUES (?, ?, ?, ?, ?, ?)",
		game.ID, action.DestRow, action.DestCol, action.SourceRow, action.SourceCol, action.Player)
	if err != nil {
		transaction.Rollback()
		return err
	}

	// update game outcome
	_, err = transaction.Exec("UPDATE Games SET Outcome = ? WHERE ID = ?", game.Outcome, game.ID)
	if err != nil {
		transaction.Rollback()
		return err
	}

	transaction.Commit()
	return nil
}

func DB_Get_Active_Games() ([]Game, error) {
	db, err := Db_open()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM Games WHERE Outcome = 0")
	if err != nil {
		return nil, err
	}

	db_games := make([]DB_Game, 0)
	for rows.Next() {
		db_game := DB_Game{}
		err = rows.Scan(&db_game.ID, &db_game.Player1ID, &db_game.Player2ID, &db_game.Outcome, &db_game.Rows, &db_game.Cols)
		if err != nil {
			return nil, err
		}
		db_games = append(db_games, db_game)

	}

	games := make([]Game, 0)
	for _, db_game := range db_games {
		game, err := reconstruct_game(db, db_game)
		if err != nil {
			return nil, err
		}
		games = append(games, *game)
	}
	return games, nil
}

func DB_Get_Active_Games_By_Player(player *Player) ([]Game, error) {
	db, err := Db_open()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM Games WHERE Outcome = 0 AND (Player1ID = ? OR Player2ID = ?)", player.ID, player.ID)
	if err != nil {
		return nil, err
	}

	db_games := make([]DB_Game, 0)
	for rows.Next() {
		db_game := DB_Game{}
		err = rows.Scan(&db_game.ID, &db_game.Player1ID, &db_game.Player2ID, &db_game.Outcome, &db_game.Rows, &db_game.Cols)
		if err != nil {
			return nil, err
		}
		db_games = append(db_games, db_game)

	}

	games := make([]Game, 0)
	for _, db_game := range db_games {
		game, err := reconstruct_game(db, db_game)
		if err != nil {
			return nil, err
		}
		games = append(games, *game)
	}
	return games, nil
}

// ------------------------------
// Player Functions
// ------------------------------

func generateToken() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 32)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}

func DB_Create_Player(name string) (string, error) {
	secretToken := generateToken()

	db, err := Db_open()
	defer db.Close()
	if err != nil {
		return "", err
	}

	row := db.QueryRow("INSERT INTO Players (Name, SecretToken, CurrentElo) VALUES (?, ?, ?)", name, secretToken, 1000)
	var id int
	err = row.Scan(&id)
	if err != nil {
		return "", err
	}

	return secretToken, nil
}
func reconstruct_history(db *sql.DB, playerID int) ([]HistoryEntry, error) {
	history := []HistoryEntry{}
	historyResults, err := db.Query("SELECT GameID, Win, Draw, Loss, Elo FROM History WHERE PlayerID = ?", playerID)
	if err != nil {
		return nil, err
	}
	for historyResults.Next() {
		db_history := HistoryEntry{}
		err = historyResults.Scan(&db_history.GameID, &db_history.Win, &db_history.Draw, &db_history.Loss, &db_history.Elo)
		if err != nil {
			return nil, err
		}
		history = append(history, db_history)
	}
	return history, nil
}

func DB_Get_Player(id int) (*Player, error) {
	db, err := Db_open()
	defer db.Close()
	if err != nil {
		return nil, err
	}

	row := db.QueryRow("SELECT * FROM Players WHERE ID = ?", id)

	db_player := DB_Player{}
	err = row.Scan(&db_player.ID, &db_player.Name, &db_player.SecretToken, &db_player.CurrentElo)
	if err != nil {
		return nil, err
	}

	// reconstruct game history
	history, err := reconstruct_history(db, db_player.ID)
	if err != nil {
		return nil, err
	}

	player := Player{
		ID:          db_player.ID,
		Name:        db_player.Name,
		SecretToken: db_player.SecretToken,
		CurrentElo:  db_player.CurrentElo,
		GameHistory: history,
	}
	return &player, nil
}

func DB_Get_Players() ([]Player, error) {
	db, err := Db_open()
	defer db.Close()
	if err != nil {
		return nil, err
	}

	rows, err := db.Query("SELECT * FROM Players")
	if err != nil {
		return nil, err
	}

	players := make([]Player, 0)
	for rows.Next() {
		player := Player{}
		err = rows.Scan(&player.ID, &player.Name, &player.SecretToken, &player.CurrentElo)
		if err != nil {
			slog.Error("Error scanning player", "error", err)
		}

		// reconstruct game history
		history, err := reconstruct_history(db, player.ID)
		if err != nil {
			slog.Error("Error reconstructing history", "error", err)
		}
		player.GameHistory = history
		players = append(players, player)
	}

	return players, nil
}

func DB_Get_Player_by_Token(token string) (*Player, error) {
	db, err := Db_open()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	row := db.QueryRow("SELECT ID FROM Players WHERE SecretToken = ?", token)
	db_player := DB_Player{}
	err = row.Scan(&db_player.ID, &db_player.Name, &db_player.SecretToken, &db_player.CurrentElo)
	if err != nil {
		return nil, err
	}

	// reconstruct game history
	history := []HistoryEntry{}
	historyResults, err := db.Query("SELECT GameID, Win, Draw, Loss, Elo FROM History WHERE PlayerID = ?", db_player.ID)
	if err != nil {
		return nil, err
	}
	for historyResults.Next() {
		db_history := HistoryEntry{}
		err = historyResults.Scan(&db_history.GameID, &db_history.Win, &db_history.Draw, &db_history.Loss, &db_history.Elo)
		if err != nil {
			return nil, err
		}
		history = append(history, db_history)
	}

	player := Player{
		ID:          db_player.ID,
		Name:        db_player.Name,
		SecretToken: db_player.SecretToken,
		CurrentElo:  db_player.CurrentElo,
		GameHistory: history,
	}
	return &player, nil
}

func DB_update_Elo_and_History(playerOneID int, playerTwoID int, outcome int, hist1 *HistoryEntry, hist2 *HistoryEntry) error {
	db, err := Db_open()
	if err != nil {
		return err
	}
	defer db.Close()

	transaction, err := db.Begin()
	if err != nil {
		return err
	}
	var currentElo_1, currentElo_2 int

	// lookup elo of player 1
	row := transaction.QueryRow("SELECT CurrentElo FROM Players WHERE ID = ?", playerOneID)
	err = row.Scan(&currentElo_1)
	if err != nil {
		transaction.Rollback()
		return err
	}

	// lookup elo of player 1
	row = transaction.QueryRow("SELECT CurrentElo FROM Players WHERE ID = ?", playerTwoID)
	err = row.Scan(&currentElo_2)
	if err != nil {
		transaction.Rollback()
		return err
	}

	_, e1, e2 := CalculateEloUpdate(currentElo_1, currentElo_2, outcome)
	// assert.True(success, "Elo should be updated")

	// update elo player 1
	_, err = transaction.Exec("UPDATE Players SET CurrentElo = ? WHERE ID = ?", e1, playerOneID)
	if err != nil {
		transaction.Rollback()
		return err
	}

	// update elo player 2
	_, err = transaction.Exec("UPDATE Players SET CurrentElo = ? WHERE ID = ?", e2, playerTwoID)
	if err != nil {
		transaction.Rollback()
		return err
	}

	// update history player 1
	_, err = transaction.Exec("INSERT INTO History (GameID, PlayerID, Win, Draw, Loss, Elo) VALUES (?, ?, ?, ?, ?, ?)", hist1.GameID, playerOneID, hist1.Win, hist1.Draw, hist1.Loss, e1)
	if err != nil {
		transaction.Rollback()
		return err
	}

	// update history player 2
	_, err = transaction.Exec("INSERT INTO History (GameID, PlayerID, Win, Draw, Loss, Elo) VALUES (?, ?, ?, ?, ?, ?)", hist2.GameID, playerTwoID, hist2.Win, hist2.Draw, hist2.Loss, e2)
	if err != nil {
		transaction.Rollback()
		return err
	}

	transaction.Commit()
	return nil
}
