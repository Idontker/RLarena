package main

import (
	"database/sql"
	"fmt"
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

func DB_Create_Game(player1_id int, player2_id int, rows int, cols int) (int, error) {
	db, err := Db_open()
	if err != nil {
		return -1, err
	}

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

	db.Close()
	return id, nil
}

func DB_Get_Game(id int) (*Game, error) {
	db, err := Db_open()
	if err != nil {
		return nil, err
	}

	// load game data
	row := db.QueryRow("SELECT * FROM Games WHERE ID = ?", id)
	db_game := DB_Game{}
	err = row.Scan(&db_game.ID, &db_game.Player1ID, &db_game.Player2ID, &db_game.Outcome, &db_game.Rows, &db_game.Cols)
	if err != nil {
		return nil, err
	}

	rows := db_game.Rows
	cols := db_game.Cols

	// load turns
	turnResults, err := db.Query("SELECT * FROM Turns WHERE ID = ?", id)
	if err != nil {
		return nil, err
	}

	db.Close()

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

func DB_apply_action(action Turn, game *Game) error {
	db, err := Db_open()
	defer db.Close()
	if err != nil {
		return err
	}

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

func DB_update_Elo_and_History(playerOneID int, playerTwoID int, outcome int, hist1 *HistoryEntry, hist2 *HistoryEntry) error {
	db, err := Db_open()
	defer db.Close()
	if err != nil {
		return err
	}

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

func DB_Get_Player(id int) (*Player, error) {
	db, err := Db_open()
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
	history := []HistoryEntry{}
	historyResults, err := db.Query("SELECT GameID, Win, Draw, Loss, Elo FROM History WHERE PlayerID = ?", id)
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
