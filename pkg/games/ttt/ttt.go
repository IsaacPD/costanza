package ttt

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var (
	p1 = "❌"
	p2 = "⭕"
)

type TicTacToe struct {
	player1 discordgo.User
	player2 discordgo.User
	turn    bool
	grid    [3][3]string
}

func New(p1, p2 discordgo.User) *TicTacToe {
	var ttt TicTacToe
	ttt.player1 = p1
	ttt.player2 = p2
	return &ttt
}

func (t *TicTacToe) Move(x, y int, player discordgo.User) (string, bool) {
	if !(t.checkValid(x, y) && t.isTurn(player)) {
		if !t.isTurn(player) {
			return fmt.Sprintf("It is not %s's turn", player.Username), false
		}
		return fmt.Sprintf("%d, %d is not a valid move", x, y), false
	}

	if player == t.player1 {
		t.grid[x][y] = p1
	} else {
		t.grid[x][y] = p2
	}
	t.turn = !t.turn
	result := t.String()
	var finished bool
	if t.isDraw() {
		result += "\nIt's a draw!"
		finished = true
	} else if finished = t.checkPlayerWin(t.grid[x][y]); finished {
		result += "\n " + player.Username + " is the winner!"
	}
	return result, finished
}

func (t *TicTacToe) isDraw() bool {
	for i := range t.grid {
		for j := range t.grid[i] {
			if t.checkValid(i, j) {
				return false
			}
		}
	}
	return true
}

func (t *TicTacToe) IsPlaying(author discordgo.User) bool {
	return author == t.player1 || author == t.player2
}

func (t *TicTacToe) isTurn(author discordgo.User) bool {
	return (author == t.player1 && !t.turn) || (author == t.player2 && t.turn)
}

func (t *TicTacToe) checkValid(x, y int) bool {
	return t.grid[x][y] == ""
}

func (t *TicTacToe) checkPlayerWin(player string) bool {
	return (t.grid[0][0] == player && t.grid[0][1] == player && t.grid[0][2] == player) || // Check all rows.
		(t.grid[1][0] == player && t.grid[1][1] == player && t.grid[1][2] == player) ||
		(t.grid[2][0] == player && t.grid[2][1] == player && t.grid[2][2] == player) ||

		(t.grid[0][0] == player && t.grid[1][0] == player && t.grid[2][0] == player) || // Check all columns.
		(t.grid[0][1] == player && t.grid[1][1] == player && t.grid[2][1] == player) ||
		(t.grid[0][2] == player && t.grid[1][2] == player && t.grid[2][2] == player) ||

		(t.grid[0][0] == player && t.grid[1][1] == player && t.grid[2][2] == player) || // Check all diagonals.
		(t.grid[0][2] == player && t.grid[1][1] == player && t.grid[2][0] == player)
}

func (t *TicTacToe) String() string {
	var result strings.Builder
	for _, row := range t.grid {
		for _, col := range row {
			if col == " " {
				fmt.Fprintf(&result, "%s | ", "🟥")
			} else {
				fmt.Fprintf(&result, "%s | ", col)
			}
		}
		fmt.Fprintln(&result)
	}
	return result.String()
}