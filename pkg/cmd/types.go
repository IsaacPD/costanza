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
		Names   Names
		Handler Handler
		Help    string
	}

	Context struct {
		Cmd  string
		Arg  string
		Args []string

		Session     *discordgo.Session
		Interaction *discordgo.InteractionCreate
		Author      *discordgo.User

		ChannelID string
		GuildID   string

		Send func(message string)
		Ack  func()
		Log  func(m *discordgo.Message, err error)
	}
)

func defaultHandler(c Context) {
	c.Send(fmt.Sprintf("%s is not implemented yet", c.Cmd))
}

func NewCmd(names Names, handler Handler, help string) Command {
	if handler == nil {
		handler = defaultHandler
	}
	return Command{names, handler, help}
}

func (c Command) String() string {
	var sb strings.Builder

	sb.WriteString("Name: " + c.Names[0])
	if len(c.Names) > 1 {
		fmt.Fprintf(&sb, " | Aliases: %s", c.Names[1:])
	}
	if c.Help != "" {
		sb.WriteString(". Help: " + c.Help)
	}
	return sb.String()
}

func (c Command) ApplicationCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        c.Names[0],
		Description: c.Help,
		Type:        discordgo.ChatApplicationCommand,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "input",
				Description: "input",
				Required:    true,
			},
		},
	}
}
