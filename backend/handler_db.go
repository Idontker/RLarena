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
	slog.Info("Create User", "player1_id", player1_id, "player2_id", player2_id)
	db, err := Db_open()
	if err != nil {
		slog.Error("Error opening database during game creation", "error", err)
		return -1, err
	}
	defer db.Close()

	result, err := db.Exec("INSERT INTO Game (Player1ID, Player2ID, Outcome, Rows, Cols) VALUES (?, ?, ?, ?, ?)",
		player1_id,
		player2_id,
		0,
		rows,
		cols)
	if err != nil {
		slog.Error("Error inserting new game to db", "error", err)
		return -1, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		slog.Error("Error reading id after game insertion", "error", err)
		return -1, err
	}

	return int(id), nil
}

func reconstruct_game(db *sql.DB, db_game DB_Game) (*Game, error) {
	rows := db_game.Rows
	cols := db_game.Cols

	// load turns
	turnResults, err := db.Query("SELECT * FROM Turn WHERE GameID = ?", db_game.ID)
	if err != nil {
		slog.Error("Error querying turns", "gameID", db_game.ID, "error", err)
		return nil, err
	}

	history := []Turn{}
	for turnResults.Next() {
		db_turn := DB_Turn{}
		err = turnResults.Scan(&db_turn.GameID, &db_turn.ID, &db_turn.GameID, &db_turn.DestRow, &db_turn.DestCol, &db_turn.SourceRow, &db_turn.SourceCol, &db_turn.PlayerNum)
		if err != nil {
			slog.Error("Error scanning turn", "game id", db_game.ID, "error", err)
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
		History: make([]Turn, 0),
		Board:   board,
	}

	// apply all turns
	for _, turn := range history {
		valid := state.applyAction(turn)
		slog.Info("Reconstructing game", "turn", turn, "valid", valid)

		if !valid {
			boardState := ""
			for _, row := range state.Board {
				boardState += fmt.Sprintln(row)
			}
			return nil, fmt.Errorf("invalid turn within game history %v for game %v. Board state:\n%v", turn, db_game.ID, boardState)
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
		slog.Error("Error opening database during game lookup", "id", id, "error", err)
		return nil, err
	}
	defer db.Close()

	// load game data
	row := db.QueryRow("SELECT * FROM Game WHERE ID = ?", id)
	db_game := DB_Game{}
	err = row.Scan(&db_game.ID, &db_game.Player1ID, &db_game.Player2ID, &db_game.Outcome, &db_game.Rows, &db_game.Cols)
	if err != nil {
		slog.Error("Error querying game by id", "id", id, "error", err)
		return nil, err
	}

	game, err := reconstruct_game(db, db_game)
	if err != nil {
		slog.Error("Error during reconstruction of game", "id", id, "error", err)
		return nil, err
	}

	return game, nil
}

func DB_apply_action(action Turn, game *Game) error {
	db, err := Db_open()
	if err != nil {
		slog.Error("Error starting opening game (apply action)", "error", err)
		return err
	}
	defer db.Close()

	transaction, err := db.Begin()
	if err != nil {
		slog.Error("Error starting transaction (apply action)", "error", err)
		return err
	}

	// add turn
	_, err = transaction.Exec("INSERT INTO Turn (GameID, TurnID, DestRow, DestCol, SourceRow, SourceCol, PlayerNum) VALUES (?, ?, ?, ?, ?, ?, ?)",
		game.ID, action.TurnID, action.DestRow, action.DestCol, action.SourceRow, action.SourceCol, action.Player)
	if err != nil {
		slog.Error("Error inserting turn (apply action)", "gameID", game.ID, "turnID", action.TurnID, "SourceCol", action.SourceCol, "SourceRow", action.SourceRow, "DestRow", action.DestRow, "DestCol", action.DestCol, "error", err)
		transaction.Rollback()
		return err
	}

	// update game outcome
	_, err = transaction.Exec("UPDATE Game SET Outcome = ? WHERE ID = ?", game.Outcome, game.ID)
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
		slog.Error("Error opening database during active game lookup", "error", err)
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM Game WHERE Outcome = 0")
	if err != nil {
		slog.Error("Error querying actives games", "error", err)
		return nil, err
	}

	db_games := make([]DB_Game, 0)
	for rows.Next() {
		db_game := DB_Game{}
		err = rows.Scan(&db_game.ID, &db_game.Player1ID, &db_game.Player2ID, &db_game.Outcome, &db_game.Rows, &db_game.Cols)
		if err != nil {
			slog.Error("Error during reading games (get active game query)", "error", err)
			return nil, err
		}
		db_games = append(db_games, db_game)

	}

	games := make([]Game, 0)
	for _, db_game := range db_games {
		game, err := reconstruct_game(db, db_game)
		if err != nil {
			slog.Error("Error during reconstruction of game", "id", db_game.ID, "error", err)
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

	rows, err := db.Query("SELECT * FROM Game WHERE Outcome = 0 AND (Player1ID = ? OR Player2ID = ?)", player.ID, player.ID)
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
			slog.Error("Error during reconstruction of game", "id", db_game.ID, "error", err)
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
	slog.Info("Create User", "name", name)
	secretToken := generateToken()

	db, err := Db_open()
	defer db.Close()
	if err != nil {
		slog.Error("Error opening database during player creation", "error", err)
		return "", err
	}

	_, err = db.Exec("INSERT INTO Player (Name, SecretToken, Elo) VALUES (?, ?, ?)", name, secretToken, 1000)
	if err != nil {
		slog.Error("Error inserting new player to db", "error", err)
		return "", err
	}

	return secretToken, nil
}

func reconstruct_history(db *sql.DB, playerID int) ([]HistoryEntry, error) {
	history := []HistoryEntry{}
	historyResults, err := db.Query("SELECT GameID, Win, Draw, Loss, Elo FROM HistoryEntry WHERE PlayerID = ?", playerID)
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

	row := db.QueryRow("SELECT * FROM Player WHERE ID = ?", id)

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

	rows, err := db.Query("SELECT * FROM Player")
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
	slog.Debug("Get Player by Token", "token", token)
	db, err := Db_open()
	if err != nil {
		slog.Error("Error opening database during player lookup (by token)", "error", err)
		return nil, err
	}
	defer db.Close()

	row := db.QueryRow("SELECT * FROM Player WHERE SecretToken = ?", token)
	db_player := DB_Player{}
	err = row.Scan(&db_player.ID, &db_player.Name, &db_player.SecretToken, &db_player.CurrentElo)
	if err != nil {
		slog.Error("Error opening database during player lookup (by token)", "token", token, "error", err)
		return nil, err
	}

	// reconstruct game history
	history := []HistoryEntry{}
	historyResults, err := db.Query("SELECT GameID, Win, Draw, Loss, Elo FROM HistoryEntry WHERE PlayerID = ?", db_player.ID)
	if err != nil {
		slog.Error("Error opening database during history lookup", "playerID", db_player.ID, "error", err)
		return nil, err
	}
	for historyResults.Next() {
		db_history := HistoryEntry{}
		err = historyResults.Scan(&db_history.GameID, &db_history.Win, &db_history.Draw, &db_history.Loss, &db_history.Elo)
		if err != nil {
			slog.Error("Error during history query (during scanning scan)", "playerID", db_player.ID, "error", err)
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
	row := transaction.QueryRow("SELECT Elo FROM Player WHERE ID = ?", playerOneID)
	err = row.Scan(&currentElo_1)
	if err != nil {
		transaction.Rollback()
		return err
	}

	// lookup elo of player 1
	row = transaction.QueryRow("SELECT Elo FROM Player WHERE ID = ?", playerTwoID)
	err = row.Scan(&currentElo_2)
	if err != nil {
		transaction.Rollback()
		return err
	}

	_, e1, e2 := CalculateEloUpdate(currentElo_1, currentElo_2, outcome)
	// assert.True(success, "Elo should be updated")

	// update elo player 1
	_, err = transaction.Exec("UPDATE Player SET Elo = ? WHERE ID = ?", e1, playerOneID)
	if err != nil {
		transaction.Rollback()
		return err
	}

	// update elo player 2
	_, err = transaction.Exec("UPDATE Player SET Elo = ? WHERE ID = ?", e2, playerTwoID)
	if err != nil {
		transaction.Rollback()
		return err
	}

	// update history player 1
	_, err = transaction.Exec("INSERT INTO HistoryEntry (GameID, PlayerID, Win, Draw, Loss, Elo) VALUES (?, ?, ?, ?, ?, ?)", hist1.GameID, playerOneID, hist1.Win, hist1.Draw, hist1.Loss, e1)
	if err != nil {
		transaction.Rollback()
		return err
	}

	// update history player 2
	_, err = transaction.Exec("INSERT INTO HistoryEntry (GameID, PlayerID, Win, Draw, Loss, Elo) VALUES (?, ?, ?, ?, ?, ?)", hist2.GameID, playerTwoID, hist2.Win, hist2.Draw, hist2.Loss, e2)
	if err != nil {
		transaction.Rollback()
		return err
	}

	transaction.Commit()
	return nil
}
