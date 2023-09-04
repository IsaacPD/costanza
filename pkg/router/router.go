package router

import (
	"errors"
	"flag"
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
var NewCmdWithOptions = cmd.NewCmdWithOptions

const (
	Isaac      = "217795612169601024"
	Ariana     = "427657255090126848"
	Images     = "/mnt/e/Desktop/Stuffs/Images"
	PREFIX     = "~"
	TimeLayout = "2006 Jan 2"
)

var (
	Zero               = 0.0
	cmdMap             = make(map[string]*cmd.Command)
	registeredCommands []*discordgo.ApplicationCommand

	Profanity = []*regexp.Regexp{
		regexp.MustCompile(".*(f+|F+)(f*|F*)(u+|U+)(u*|U*)(c+|C+)(c*|C*)(k+|K+)(k*|K*).*"),
		regexp.MustCompile(".*(s+|S+)(s*|S*)(h+|H+)(h*|H*)(i+|I+)(i*|I*)(t+|T+)(t*|T*).*"),
		regexp.MustCompile(".*(b+|B+)(b*|B*)(i+|I+)(i*|I*)(t+|T+)(t*|T*)(c+|C+)(c*|C*)(h+|H+)(h*|H*).*"),
		regexp.MustCompile(".*(c+|C+)(c*|C*)(u+|U+)(u*|U*)(n+|N+)(n*|N*)(t+|T+)(t*|T*).*"),
	}

	appID = "319980316309848064"

	Sc chan os.Signal

	StringOption = discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionString,
		Name:        "input",
		Description: "input",
		Required:    true,
	}
	UserOption = discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionUser,
		Name:        "user",
		Description: "user",
		Required:    true,
	}

	SkipRegister bool

	ErrInvalidPermission error = errors.New("invalid permission")
	ErrInvalidTimeFormat error = errors.New("invalid date, please format it as follows: " + TimeLayout)
)

func init() {
	flag.BoolVar(&SkipRegister, "s", false, "Skip registering commands with discord. Should only be done if there are no changes with the commands.")
}

func AddCommand(cmd cmd.Command, s *discordgo.Session) {
	for _, a := range cmd.Names {
		cmdMap[a] = &cmd
	}
	if cmd.Names[0] != "" && !SkipRegister {
		c, err := s.ApplicationCommandCreate(appID, "", cmd.ApplicationCommand())
		if err != nil {
			logrus.Panicf("Cannot create '%v' command: %v", cmd.Names[0], err)
		}
		registeredCommands = append(registeredCommands, c)
	}
}

func HandleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	logrus.Tracef("Received Message {%s}", m.Content)
	m.Content = strings.TrimSpace(m.Content)
	if m.Author.ID == s.State.User.ID {
		return
	}

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

	send := sendClosureMessage(s, m)
	arg := strings.TrimSpace(util.AfterCommand(m.Content))
	ctx := cmd.Context{
		Cmd:           command,
		Arg:           arg,
		Args:          strings.Split(arg, " "),
		Message:       m,
		Session:       s,
		Author:        m.Author,
		ChannelID:     m.ChannelID,
		GuildID:       m.GuildID,
		SendEphemeral: send,
		Log:           Log,
	}
	finalize(c, ctx)
}

func HandleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	var input string
	if data.Options[0].Type == discordgo.ApplicationCommandOptionString {
		input = strings.TrimSpace(data.Options[0].StringValue())
	}
	logrus.Tracef("Received Command {%+v}", data)
	logrus.Infof("Command received %s", data.Name)

	c, ok := cmdMap[data.Name]
	if !ok {
		logrus.Trace("Command not found in registered list")
		return
	}

	var author *discordgo.User
	if i.Member != nil {
		author = i.Member.User
	} else {
		author = i.User
	}

	send := sendClosure(s, i)
	arg := strings.TrimSpace(input)
	ctx := cmd.Context{
		Cmd:           data.Name,
		Arg:           arg,
		Args:          strings.Split(arg, " "),
		Session:       s,
		Interaction:   i,
		Author:        author,
		ChannelID:     i.ChannelID,
		GuildID:       i.GuildID,
		SendEphemeral: send,
		Ack:           sendAck(s, i),
		Defer:         sendDefer(s, i),
		Followup:      sendFollowup(s, i),
		Log:           Log,
	}
	finalize(c, ctx)
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
	return cmd.NewCmd(cmd.Names{""}, func(c cmd.Context) (string, error) {
		// if handlePrivateMessage(c.Session, c.Message) {
		// 	return
		// }

		if ttt.Turn(c) {
			return "", nil
		}

		if c.Arg == "Goodbye Costanza" {
			c.Send("See you nerds later")
			Sc <- os.Interrupt
			return "", nil
		}

		// for _, p := range Profanity {
		// 	if p.MatchString(c.Arg) {
		// 		c.Send("Watch your profanity " + c.Author.Username)
		// 	}
		// }

		if rand.Float64() < 0.001 {
			c.Send("Hello There @" + c.Author.Username)
		}
		return "", nil
	}, "default")
}

func helpCommand() cmd.Command {
	return NewCmd(cmd.Names{"help"}, func(c cmd.Context) (string, error) {
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
		return sb.String(), nil
	}, "Help")
}

func tttCommand() cmd.Command {
	return NewCmdWithOptions(cmd.Names{"ttt", "tictactoe"}, ttt.HandleTTT, "Play tic tac toe", &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Name:        "start",
		Description: "Start a new game of tic tac toe with the specified user",
		Options:     []*discordgo.ApplicationCommandOption{&UserOption},
	}, &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Name:        "turn",
		Description: "Take your turn in tic tac toe",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "row",
				Description: "The row where you will take your turn.",
				MinValue:    &Zero,
				MaxValue:    2,
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "col",
				Description: "The row where you will take your turn.",
				MinValue:    &Zero,
				MaxValue:    2,
			},
		},
	})
}

func RegisterCommands(s *discordgo.Session) {
	addCommand := func(c cmd.Command) {
		AddCommand(c, s)
	}

	// Player commands
	addCommand(NewCmd(cmd.Names{"back", "b"}, player.Previous, "Go to previous track"))
	addCommand(NewCmdWithOptions(cmd.Names{"play", "p"}, player.QueueTrack, "Queue a track", &StringOption))
	addCommand(NewCmdWithOptions(cmd.Names{"playnext", "pn"}, player.Play, "Play a track", &StringOption))
	addCommand(NewCmd(cmd.Names{"queue", "q"}, player.PrintQueue, "Show the queue"))
	addCommand(NewCmd(cmd.Names{"skip"}, player.Skip, "Skip a track in the queue"))
	addCommand(NewCmd(cmd.Names{"pause"}, player.Pause, "Pause the current track"))
	addCommand(NewCmd(cmd.Names{"unpause"}, player.UnPause, "Unpause the current track"))
	addCommand(NewCmd(cmd.Names{"tree"}, player.ListDir, "Print out the directory"))

	// Misc
	addCommand(NewCmdWithOptions(cmd.Names{"themepicker", "tp"}, func(c cmd.Context) (string, error) {
		if c.Author.ID != Isaac {
			return "", ErrInvalidPermission
		}

		t, err := time.Parse(TimeLayout, c.Arg)
		if err != nil {
			return "", ErrInvalidTimeFormat
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
		return message.String(), nil
	}, "Display the themepickers for the following months", &StringOption))
	addCommand(NewCmdWithOptions(cmd.Names{"translate"}, google.Translate, "Translate some text", &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionString,
		Name:        "text",
		Description: "The text that will be translated. Use semicolons to separate multiple strings to translate.",
		Required:    true,
	}, &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionString,
		Name:        "target",
		Description: "The language to translate it into (also accepts language codes e.g. `fr` for french)",
    Required:     true,
	}))
	addCommand(NewCmd(cmd.Names{"listen"}, nil, "Listen to voice"))
	addCommand(tttCommand())
	addCommand(NewCmdWithOptions(cmd.Names{"send"}, func(c cmd.Context) (string, error) {
		f, err := os.Open(fmt.Sprintf("%s/%s", Images, c.Arg))
		if err != nil {
			logrus.Warnf("Cannot open file %s", c.Arg)
			return "", fmt.Errorf("cannot open file")
		}
		c.Log(c.Session.ChannelFileSend(c.ChannelID, c.Arg, f))
		return "Done", nil
	}, "Send a file that costanza has", &StringOption))
	addCommand(NewCmdWithOptions(cmd.Names{"speak"}, func(c cmd.Context) (string, error) {
		if c.Author.ID != Isaac && c.Author.ID != Ariana {
			return "", ErrInvalidPermission
		}
		Log(c.Session.ChannelMessageSendTTS(c.ChannelID, c.Arg))
		if c.Interaction != nil {
			c.SendEphemeral("Done", true)
		}
		return "", nil
	}, "Make costanza speak", &StringOption))
	addCommand(NewCmdWithOptions(cmd.Names{"archive"}, util.Archive, "Archives a channel", &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionChannel,
		Name:        "channel",
		Description: "the channel to archive",
		Required:    true,
	}))
	addCommand(helpCommand())
	addCommand(defaultCommand())
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

func finalize(c *cmd.Command, ctx cmd.Context) {
	message, err := c.Handler(ctx)
	if err != nil {
		ctx.Send(fmt.Sprintf("error: %v", err))
	} else if message != "" {
		ctx.Send(message)
	}
}

func sendClosureMessage(s *discordgo.Session, m *discordgo.MessageCreate) func(string, bool) {
	return func(send string, _ bool) { Log(s.ChannelMessageSend(m.ChannelID, send)) }
}

func sendClosure(s *discordgo.Session, i *discordgo.InteractionCreate) func(string, bool) {
	return func(send string, isEphemeral bool) {
		var flags discordgo.MessageFlags
		if isEphemeral {
			flags = discordgo.MessageFlagsEphemeral
		}
		LogError(s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: send,
				Flags:   flags,
			},
		}))
	}
}

func sendAck(s *discordgo.Session, i *discordgo.InteractionCreate) func() {
	return func() {
		LogError(s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		}))
	}
}

func sendDefer(s *discordgo.Session, i *discordgo.InteractionCreate) func() {
	return func() {
		LogError(s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		}))
	}
}

func sendFollowup(s *discordgo.Session, i *discordgo.InteractionCreate) func(string) {
	return func(send string) {
		_, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: send,
		})
		LogError(err)
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
