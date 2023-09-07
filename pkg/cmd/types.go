package cmd

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

type (
	Handler func(c Context) (string, error)
	Names   []string

	Command struct {
		Names   Names
		Handler Handler
		Help    string
		Inputs  []*discordgo.ApplicationCommandOption
	}

	Context struct {
		Cmd  string
		Arg  string
		Args []string

		Session     *discordgo.Session
		Interaction *discordgo.InteractionCreate
		Message     *discordgo.MessageCreate
		Author      *discordgo.User

		ChannelID string
		GuildID   string

		Options map[string]interface{}

		SendEphemeral func(message string, isEphemeral bool)
		Ack           func()
		Defer         func()
		Followup      func(message string)
		Log           func(m *discordgo.Message, err error)
	}
)

func GetOption[T any](c *Context, name string) T {
	v, present := c.Options[name]
	if !present {
		logrus.Panicf("%v option not present.", name)
	}
	return v.(T)
}

func GetOptionWithDefault[T any](c *Context, name string, def T) T {
	v, present := c.Options[name]
	if present {
		return v.(T)
	}
	return def
}

func (c *Context) Send(message string) {
	c.SendEphemeral(message, false)
}

func defaultHandler(c Context) (string, error) {
	return "", fmt.Errorf("%s is not implemented yet", c.Cmd)
}

func NewCmd(names Names, handler Handler, help string) Command {
	if handler == nil {
		handler = defaultHandler
	}
	return Command{names, handler, help, nil}
}

func NewCmdWithOptions(names Names, handler Handler, help string, options ...*discordgo.ApplicationCommandOption) Command {
	if handler == nil {
		handler = defaultHandler
	}
	return Command{names, handler, help, options}
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
		Options:     c.Inputs,
	}
}
