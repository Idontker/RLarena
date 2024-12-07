package main

import (
	"encoding/json"
	"fmt"
)

type Turn struct {
	TurnID    int `json:"turnID"`
	DestRow   int `json:"destRow"`
	DestCol   int `json:"destCol"`
	SourceRow int `json:"sourceRow"`
	SourceCol int `json:"sourceCol"`
	Player    int `json:"player"`
}

func (t *Turn) Eq(other Turn) bool {
	return t.DestCol == other.DestCol && t.DestRow == other.DestRow && t.SourceCol == other.SourceCol && t.SourceRow == other.SourceRow && t.Player == other.Player
}

type GameState struct {
	Rows    int     `json:"rows"`
	Cols    int     `json:"cols"`
	History []Turn  `json:"history"`
	Board   [][]int `json:"board"`
}

// Implement the json.Marshaler interface
func (g GameState) MarshalJSON() ([]byte, error) {
	// Create an alias to avoid recursion
	type Alias GameState

	// Embed the original struct with derived fields
	return json.Marshal(&struct {
		Alias
		GameOver      bool   `json:"gameOver"`
		Winner        int    `json:"winner"`
		MoveOptions   []Turn `json:"moveOptions"`
		CurrentPlayer int    `json:"currentPlayer"`
	}{
		Alias:         Alias(g),
		GameOver:      g.IsEnd(),
		Winner:        g.GetWinner(),
		MoveOptions:   g.PossibleMoves(),
		CurrentPlayer: g.NextPlayer(),
	})
}

func (g *GameState) Display() {
	fmt.Printf("Rows: %d, Cols: %d\n", g.Rows, g.Cols)
	fmt.Println("Board:")
	for _, row := range g.Board {
		fmt.Println(row)
	}
	fmt.Println("History:")
	for _, turn := range g.History {
		fmt.Printf("Player %d moved (%d, %d) to (%d, %d)\n", turn.Player, turn.SourceRow, turn.SourceCol, turn.DestRow, turn.DestCol)
	}
}

func (g *GameState) NextPlayer() int {
	return len(g.History)%2 + 1
}

func (g *GameState) IsEnd() bool {
	return g.GetWinner() != 0 || len(g.PossibleMoves()) == 0
}

func (g *GameState) GetWinner() int {
	for i := 0; i < g.Cols; i++ {
		if g.Board[0][i] == 2 {
			return 2
		} else if g.Board[g.Rows-1][i] == 1 {
			return 1
		}
	}
	// draw
	if len(g.PossibleMoves()) == 0 {
		return -1
	}
	return 0
}

func (g *GameState) PossibleMoves() []Turn {
	np := g.NextPlayer()
	nextMoves := make([]Turn, 0)

	for col := 0; col < g.Cols; col++ {
		for row := 0; row < g.Rows; row++ {
			if g.Board[row][col] == np {
				var dy int
				if np == 1 {
					dy = 1
				} else {
					dy = -1
				}

				for _, dx := range []int{-1, 0, 1} {
					x := col + dx
					y := row + dy

					if x < 0 || x >= g.Cols || y < 0 || y >= g.Rows {
						continue
					}

					if dx != 0 && g.Board[y][x] != np && g.Board[y][x] != 0 {
						nextMoves = append(nextMoves, Turn{
							DestRow:   y,
							DestCol:   x,
							SourceRow: row,
							SourceCol: col,
							Player:    np,
						})
					} else if dx == 0 && g.Board[y][x] == 0 {
						nextMoves = append(nextMoves, Turn{
							DestRow:   y,
							DestCol:   x,
							SourceRow: row,
							SourceCol: col,
							Player:    np,
						})
					}
				}
			}
		}
	}
	return nextMoves
}

func (g *GameState) applyAction(action Turn) bool {
	moves := g.PossibleMoves()

	for _, move := range moves {
		if move.Eq(action) {
			drow, dcol := action.DestRow, action.DestCol
			srow, scol := action.SourceRow, action.SourceCol
			g.Board[drow][dcol] = action.Player
			g.Board[srow][scol] = 0

			g.History = append(g.History, action)
			return true
		}
	}
	return false
}
