package cmd

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type (
	Handler func(c Context)
	Names   []string

	Command struct {
		Names   []string
		Handler Handler
		Help    string
	}

	Context struct {
		Cmd  string
		Arg  string
		Args []string

		Session *discordgo.Session
		Message *discordgo.MessageCreate
		Author  *discordgo.User

		ChannelID string
		GuildID   string

		Send func(message string)
		Log  func(m *discordgo.Message, err error)
	}
)

func defaultHandler(c Context) {
	c.Send(fmt.Sprintf("%s is not implemented yet", c.Cmd))
}

func NewCmd(names Names, handler Handler) Command {
	if handler == nil {
		handler = defaultHandler
	}
	return Command{names, handler, ""}
}

func (c Command) String() string {
	var sb strings.Builder

	sb.WriteString("Name: " + c.Names[0])
	if len(c.Names) > 1 {
		fmt.Fprintf(&sb, " | Aliases: %s", c.Names[1:])
	}
	return sb.String()
}
