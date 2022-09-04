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
)

var NewCmd = cmd.NewCmd

const (
	Isaac      = "217795612169601024"
	Images     = "/mnt/e/Desktop/Stuffs/Images"
	TimeLayout = "2006 Jan 2"
)

var (
	cmdMap             = make(map[string]*cmd.Command)
	registeredCommands []*discordgo.ApplicationCommand

	Profanity = []*regexp.Regexp{
		regexp.MustCompile(".*(f+|F+)(f*|F*)(u+|U+)(u*|U*)(c+|C+)(c*|C*)(k+|K+)(k*|K*).*"),
		regexp.MustCompile(".*(s+|S+)(s*|S*)(h+|H+)(h*|H*)(i+|I+)(i*|I*)(t+|T+)(t*|T*).*"),
		regexp.MustCompile(".*(b+|B+)(b*|B*)(i+|I+)(i*|I*)(t+|T+)(t*|T*)(c+|C+)(c*|C*)(h+|H+)(h*|H*).*"),
		regexp.MustCompile(".*(c+|C+)(c*|C*)(u+|U+)(u*|U*)(n+|N+)(n*|N*)(t+|T+)(t*|T*).*"),
	}

	appID = "319980316309848064"
	Sc    chan os.Signal
)

func AddCommand(cmd cmd.Command, s *discordgo.Session) {
	for _, a := range cmd.Names {
		cmdMap[a] = &cmd
	}
	c, err := s.ApplicationCommandCreate(appID, "", cmd.ApplicationCommand())
	if err != nil {
		logrus.Panicf("Cannot create '%v' command: %v", cmd.Names[0], err)
	}
	registeredCommands = append(registeredCommands, c)
}

func HandleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	input := strings.TrimSpace(data.Options[0].StringValue())
	logrus.Tracef("Received Command {%+v}", data)
	logrus.Infof("Command received %s", data.Name)

	c, ok := cmdMap[data.Name]
	if !ok {
		logrus.Trace("Command not found in registered list")
		return
	}

	send := sendClosure(s, i)
	arg := strings.TrimSpace(input)
	ctx := cmd.Context{
		Cmd:         data.Name,
		Arg:         arg,
		Args:        strings.Split(arg, " "),
		Session:     s,
		Interaction: i,
		Author:      i.Member.User,
		ChannelID:   i.ChannelID,
		GuildID:     i.GuildID,
		Send:        send,
		Ack:         sendAck(s, i),
		Log:         Log,
	}
	c.Handler(ctx)
}

func HandleMessageDelete(s *discordgo.Session, m *discordgo.MessageDelete) {
	var sb strings.Builder
	if m.BeforeDelete != nil && !m.BeforeDelete.Author.Bot {
		fmt.Fprintf(&sb, "%s deleted the following message \n\n %s \n\n What are you trying to hide? 🤨🤔", m.BeforeDelete.Author.Mention(), m.BeforeDelete.Content)
		Log(s.ChannelMessageSend(m.ChannelID, sb.String()))
	}
}

func HandleMessageUpdate(s *discordgo.Session, m *discordgo.MessageUpdate) {
	var sb strings.Builder
	if m.BeforeUpdate != nil && !m.BeforeUpdate.Author.Bot {
		fmt.Fprintf(&sb, "%s updated the following message \n\n`Before:`\n%s\n\n`After:`\n%s\n\n What are you trying to hide? 🤨🤔", m.BeforeUpdate.Author.Mention(), m.BeforeUpdate.Content, m.Content)
		Log(s.ChannelMessageSend(m.ChannelID, sb.String()))
	}
}

func defaultCommand() cmd.Command {
	return cmd.NewCmd(cmd.Names{""}, func(c cmd.Context) {
		// if handlePrivateMessage(c.Session, c.Message) {
		// 	return
		// }

		if ttt.HandleTTT(c) {
			return
		}

		if c.Arg == "Goodbye Costanza" {
			c.Send("See you nerds later")
			Sc <- os.Interrupt
			return
		}

		for _, p := range Profanity {
			if p.MatchString(c.Arg) {
				c.Send("Watch your profanity " + c.Author.Username)
			}
		}

		if rand.Float64() < 0.001 {
			c.Send("Hello There @" + c.Author.Username)
		}
	}, "default")
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
	}, "Help")
}

func RegisterCommands(s *discordgo.Session) {
	// Player commands
	AddCommand(NewCmd(cmd.Names{"back", "b"}, player.Previous, "Go to previous track"), s)
	AddCommand(NewCmd(cmd.Names{"play", "p"}, player.QueueTrack, "Queue a track"), s)
	AddCommand(NewCmd(cmd.Names{"playnext", "pn"}, player.Play, "Play a track"), s)
	AddCommand(NewCmd(cmd.Names{"queue", "q"}, player.PrintQueue, "Show the queue"), s)
	AddCommand(NewCmd(cmd.Names{"skip"}, player.Skip, "Skip a track in the queue"), s)
	AddCommand(NewCmd(cmd.Names{"pause"}, player.Pause, "Pause the current track"), s)
	AddCommand(NewCmd(cmd.Names{"unpause"}, player.UnPause, "Unpause the current track"), s)
	AddCommand(NewCmd(cmd.Names{"tree"}, player.ListDir, "Print out the directory"), s)

	// Misc
	AddCommand(NewCmd(cmd.Names{"themepicker", "tp"}, func(c cmd.Context) {
		if c.Author.ID != Isaac {
			return
		}

		t, err := time.Parse(TimeLayout, c.Arg)
		if err != nil {
			c.Send("Invalid date, Please format it as follows: " + TimeLayout)
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
		c.Send(message.String())
	}, "Display the themepickers for the following months"), s)
	AddCommand(NewCmd(cmd.Names{"translate"}, google.Translate, "Translate some text"), s)
	AddCommand(NewCmd(cmd.Names{"listen"}, nil, "Listen to voice"), s)
	AddCommand(NewCmd(cmd.Names{"ttt", "tictactoe"}, ttt.Start, "Play tic tac toe"), s)
	AddCommand(NewCmd(cmd.Names{"send"}, func(c cmd.Context) {
		f, err := os.Open(fmt.Sprintf("%s/%s", Images, c.Arg))
		if err != nil {
			logrus.Warnf("Cannot open file %s", c.Arg)
			return
		}
		c.Ack()
		c.Log(c.Session.ChannelFileSend(c.ChannelID, c.Arg, f))
	}, "Send a file that costanza has"), s)
	AddCommand(NewCmd(cmd.Names{"speak"}, func(c cmd.Context) {
		if c.Author.ID != Isaac {
			return
		}
		c.Ack()
		Log(c.Session.ChannelMessageSendTTS(c.ChannelID, c.Arg))
	}, "Make costanza speak"), s)
	AddCommand(helpCommand(), s)
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

func sendClosure(s *discordgo.Session, i *discordgo.InteractionCreate) func(string) {
	return func(send string) {
		LogError(s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: send,
			},
		}))
	}
}

func sendAck(s *discordgo.Session, i *discordgo.InteractionCreate) func() {
	return func() {
		LogError(s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponsePong,
		}))
	}
}

func LogError(err error) {
	if err != nil {
		logrus.Warnf("Error sending message: %s", err)
	}
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
