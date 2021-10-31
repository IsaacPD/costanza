package router

import (
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"

	"github.com/isaacpd/costanza/pkg/cmd"
	"github.com/isaacpd/costanza/pkg/games/ttt"
	"github.com/isaacpd/costanza/pkg/google"
	"github.com/isaacpd/costanza/pkg/sound/player"
	"github.com/isaacpd/costanza/pkg/util"
)

var NewCmd = cmd.NewCmd

const (
	Isaac      = "217795612169601024"
	Images     = "/mnt/e/Desktop/Stuffs/Images"
	PREFIX     = "~"
	TimeLayout = "2006 Jan 2"
)

var (
	cmdMap = make(map[string]*cmd.Command)

	Profanity = []*regexp.Regexp{
		regexp.MustCompile(".*(f+|F+)(f*|F*)(u+|U+)(u*|U*)(c+|C+)(c*|C*)(k+|K+)(k*|K*).*"),
		regexp.MustCompile(".*(s+|S+)(s*|S*)(h+|H+)(h*|H*)(i+|I+)(i*|I*)(t+|T+)(t*|T*).*"),
		regexp.MustCompile(".*(b+|B+)(b*|B*)(i+|I+)(i*|I*)(t+|T+)(t*|T*)(c+|C+)(c*|C*)(h+|H+)(h*|H*).*"),
		regexp.MustCompile(".*(c+|C+)(c*|C*)(u+|U+)(u*|U*)(n+|N+)(n*|N*)(t+|T+)(t*|T*).*"),
	}

	Sc chan os.Signal
)

func AddCommand(cmd cmd.Command) {
	for _, a := range cmd.Names {
		cmdMap[a] = &cmd
	}
}

func HandleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	m.Content = strings.TrimSpace(m.Content)
	if m.Author.ID == s.State.User.ID {
		return
	}
	logrus.Tracef("Received Message {%s}", m.Content)

	var command string
	if strings.HasPrefix(m.Content, PREFIX) && len(m.Content) > len(PREFIX) {
		command = m.Content[len(PREFIX):]
		if i := strings.Index(command, util.ARG_IDENTIFIER); i != -1 {
			command = command[:i]
		}
		logrus.Infof("Command received %s", command)
	}

	c, ok := cmdMap[command]
	if !ok {
		logrus.Trace("Command not found in registered list")
		return
	}

	send := sendClosure(s, m)
	arg := strings.TrimSpace(util.AfterCommand(m.Content))
	ctx := cmd.Context{
		Cmd:       command,
		Arg:       arg,
		Args:      strings.Split(arg, " "),
		Session:   s,
		Message:   m,
		Author:    m.Author,
		ChannelID: m.ChannelID,
		GuildID:   m.GuildID,
		Send:      send,
		Log:       Log,
	}
	c.Handler(ctx)
}

func defaultCommand() cmd.Command {
	return cmd.NewCmd(cmd.Names{""}, func(c cmd.Context) {
		if handlePrivateMessage(c.Session, c.Message) {
			return
		}

		if ttt.HandleTTT(c) {
			return
		}

		if c.Message.Content == "Goodbye Costanza" {
			c.Send("See you nerds later")
			Sc <- os.Interrupt
			return
		}

		for _, p := range Profanity {
			if p.MatchString(c.Message.Content) {
				c.Send("Watch your profanity " + c.Author.Username)
			}
		}

		if rand.Float64() < 0.001 {
			c.Send("Hello There @" + c.Author.Username)
		}
	})
}

func helpCommand() cmd.Command {
	return NewCmd(cmd.Names{"help"}, func(c cmd.Context) {
		cmdSet := make(map[*cmd.Command]interface{})

		for _, v := range cmdMap {
			cmdSet[v] = true
		}

		var sb strings.Builder
		sb.WriteString("```\n")
		for command := range cmdSet {
			if command.Names[0] == "" {
				continue
			}
			fmt.Fprintf(&sb, "%s\n", command)
		}
		sb.WriteString("```")
		c.Send(sb.String())
	})
}

func RegisterCommands() {
	// Player commands
	AddCommand(NewCmd(cmd.Names{"back", "b"}, player.Previous))
	AddCommand(NewCmd(cmd.Names{"play", "p"}, player.QueueTrack))
	AddCommand(NewCmd(cmd.Names{"playnext", "pn"}, player.Play))
	AddCommand(NewCmd(cmd.Names{"queue", "q"}, player.PrintQueue))
	AddCommand(NewCmd(cmd.Names{"skip"}, player.Skip))
	AddCommand(NewCmd(cmd.Names{"pause"}, player.Pause))
	AddCommand(NewCmd(cmd.Names{"unpause"}, player.UnPause))
	AddCommand(NewCmd(cmd.Names{"tree"}, player.ListDir))

	// Misc
	AddCommand(NewCmd(cmd.Names{"translate"}, google.Translate))
	AddCommand(NewCmd(cmd.Names{"listen"}, nil))
	AddCommand(NewCmd(cmd.Names{"ttt", "tictactoe"}, func(c cmd.Context) {
		if len(c.Message.Mentions) != 1 {
			c.Send("Too many users mentioned, please specify only one person you would like to play against.")
			return
		}
		ttt.Start(c)
	}))
	AddCommand(NewCmd(cmd.Names{"send"}, func(c cmd.Context) {
		f, err := os.Open(fmt.Sprintf("%s/%s", Images, c.Arg))
		if err != nil {
			logrus.Warnf("Cannot open file %s", c.Arg)
			return
		}
		c.Log(c.Session.ChannelFileSend(c.ChannelID, c.Arg, f))
	}))
	AddCommand(NewCmd(cmd.Names{"speak"}, func(c cmd.Context) {
		if c.Author.ID != Isaac {
			return
		}
		err := c.Session.ChannelMessageDelete(c.ChannelID, c.Message.ID)
		if err != nil {
			logrus.Warnf("Couldn't delete message %s error: %s", c.Message.ID, err)
		}
		Log(c.Session.ChannelMessageSendTTS(c.ChannelID, c.Arg))
	}))
	AddCommand(NewCmd(cmd.Names{"themepicker", "tp"}, func(c cmd.Context) {
		if c.Author.ID != Isaac {
			return
		}

		t, err := time.Parse(TimeLayout, c.Arg)
		if err != nil {
			Log(c.Session.ChannelMessageSend(c.ChannelID, "Invalid date, Please format it as follows: "+TimeLayout))
			return
		}
		users := randomUserList(c, t.UnixNano())
		var message strings.Builder
		currentMonth := t.Month() + 1
		for _, u := range users {
			fmt.Fprintf(&message, "%s %s\n", currentMonth.String(), u.Mention())
			currentMonth = (currentMonth + 1) % 13
			if currentMonth == 0 {
				currentMonth = 1
			}
		}
		Log(c.Session.ChannelMessageSend(c.ChannelID, message.String()))
	}))
	AddCommand(helpCommand())
	AddCommand(defaultCommand())
}

func randomUserList(c cmd.Context, seed int64) []*discordgo.User {
	r := rand.New(rand.NewSource(seed))
	var userList []*discordgo.User
	members, _ := c.Session.GuildMembers(c.GuildID, "", 1000)
	for _, m := range members {
		if m.User.Bot {
			continue
		}
		userList = append(userList, m.User)
	}
	for i := range userList {
		j := r.Intn(i + 1)
		userList[i], userList[j] = userList[j], userList[i]
	}
	return userList
}

func sendClosure(s *discordgo.Session, m *discordgo.MessageCreate) func(string) {
	return func(send string) { Log(s.ChannelMessageSend(m.ChannelID, send)) }
}

func Log(m *discordgo.Message, err error) {
	if err != nil {
		logrus.Warnf("Error sending message: %s", err)
	} else {
		logrus.Tracef("Message sent %v", *m)
	}
}

func handlePrivateMessage(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	message := m.Content

	if m.GuildID != "" {
		return false
	}
	logrus.Infof("Received Private Message %s: {%s}\n", m.Author.Username, message)
	if m.Author.ID != Isaac {
		return false
	}

	userIDAndMessage := strings.Split(message, ":")
	dm, err := s.UserChannelCreate(userIDAndMessage[0])

	if err != nil {
		logrus.Errorf("Couldn't create channel %s", err)
		return false
	}
	if userIDAndMessage[1] == "" {
		messages, _ := s.ChannelMessages(dm.ID, 5, "", "", "")
		for _, m := range messages {
			logrus.Infof("%s: %s", m.Author.Username, m.Content)
		}
		return true
	}
	_, err = s.ChannelMessageSend(dm.ID, userIDAndMessage[1])
	if err != nil {
		logrus.Warnf("Error sending dm %s", err)
	}
	return true
}
