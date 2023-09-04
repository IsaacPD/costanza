package ttt

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"

	"github.com/isaacpd/costanza/pkg/cmd"
)

var (
	p1 = "❌"
	p2 = "⭕"

	playingTTT bool
	toe        *TicTacToe

	Coordinate = regexp.MustCompile("[0-2],[0-2]")
)

type TicTacToe struct {
	player1 string
	player2 string
	turn    bool
	grid    [3][3]string
}

func New(p1, p2 discordgo.User) *TicTacToe {
	var ttt TicTacToe
	ttt.player1 = p1.ID
	ttt.player2 = p2.ID
	return &ttt
}

func HandleTTT(c cmd.Context) (string, error) {
	if c.Message != nil {
		if Turn(c) {
			return "", nil
		}
		return Start(c)
	}

	option := c.Interaction.ApplicationCommandData().Options[0]
	if option.Name == "start" {
		return Start(c)
	} else {
		if Turn(c) {
			return "", nil
		}
		return "", fmt.Errorf("it is not your turn or the game has not started yet")
	}
}

func Turn(c cmd.Context) bool {
	logrus.Tracef("playingTTT: %v", playingTTT)
	if !playingTTT || !toe.IsPlaying(c.Author.ID) {
		return false
	}
	var x, y int64
	if c.Message != nil {
		if !Coordinate.MatchString(c.Arg) {
			return false
		}
		fmt.Sscanf(c.Arg, "%d,%d", &x, &y)
	} else {
		x = c.Interaction.ApplicationCommandData().Options[0].Options[0].IntValue()
		y = c.Interaction.ApplicationCommandData().Options[0].Options[1].IntValue()
	}
	result, finished := toe.Move(int(x), int(y), *c.Author)
	c.Send(result)
	playingTTT = !finished
	return true
}

func Start(c cmd.Context) (string, error) {
	if len(c.Args) != 1 {
		return "", fmt.Errorf("too many users mentioned, please specify only one person you would like to play against")
	}
	if c.Message != nil {
		toe = New(*c.Author, *c.Message.Mentions[0])
	} else {
		u := c.Interaction.ApplicationCommandData().Options[0].Options[0].UserValue(c.Session)
		toe = New(*c.Author, *u)
	}
	playingTTT = true
	return toe.String(), nil
}

func (t *TicTacToe) Move(x, y int, player discordgo.User) (string, bool) {
	if !(t.checkValid(x, y) && t.isTurn(player.ID)) {
		if !t.isTurn(player.ID) {
			return fmt.Sprintf("It is not %s's turn", player.Username), false
		}
		return fmt.Sprintf("%d, %d is not a valid move", x, y), false
	}

	if player.ID == t.player1 {
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

func (t *TicTacToe) IsPlaying(author string) bool {
	logrus.Tracef("author: %v, p1: %v p2: %v", author, t.player1, t.player2)
	return author == t.player1 || author == t.player2
}

func (t *TicTacToe) isTurn(author string) bool {
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
			if col == "" {
				fmt.Fprintf(&result, "%s | ", "🟥")
			} else {
				fmt.Fprintf(&result, "%s | ", col)
			}
		}
		fmt.Fprintln(&result)
	}
	return result.String()
}
