package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"

	"github.com/isaacpd/costanza/pkg/games/ttt"
	"github.com/isaacpd/costanza/pkg/google"
	"github.com/isaacpd/costanza/pkg/sound/player"
	"github.com/isaacpd/costanza/pkg/util"
)

var (
	Isaac  = "217795612169601024"
	Images = "/mnt/e/Desktop/Stuffs/Images"

	Coordinate = regexp.MustCompile("[0-2],[0-2]")
	Profanity  = []*regexp.Regexp{
		regexp.MustCompile(".*(f+|F+)(f*|F*)(u+|U+)(u*|U*)(c+|C+)(c*|C*)(k+|K+)(k*|K*).*"),
		regexp.MustCompile(".*(s+|S+)(s*|S*)(h+|H+)(h*|H*)(i+|I+)(i*|I*)(t+|T+)(t*|T*).*"),
		regexp.MustCompile(".*(b+|B+)(b*|B*)(i+|I+)(i*|I*)(t+|T+)(t*|T*)(c+|C+)(c*|C*)(h+|H+)(h*|H*).*"),
		regexp.MustCompile(".*(c+|C+)(c*|C*)(u+|U+)(u*|U*)(n+|N+)(n*|N*)(t+|T+)(t*|T*).*"),
	}

	Token      string
	Verbosity  string
	PlayingTTT bool
	toe        *ttt.TicTacToe
	sc         chan os.Signal
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot token")
	flag.StringVar(&Verbosity, "v", "info", "Verbosity level")
	sc = make(chan os.Signal, 1)
}

func main() {
	flag.Parse()
	logrus.SetOutput(os.Stdout)
	google.InitializeServices(context.Background())

	lvl, err := logrus.ParseLevel(Verbosity)
	if err != nil {
		fmt.Printf("error parsing log level %s\n", err)
		logrus.SetLevel(logrus.TraceLevel)
	} else {
		logrus.SetLevel(lvl)
	}

	if Token == "" {
		Token = os.Getenv("COSTANZA_TOKEN")
	}
	discord, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("Error creating bot:", err)
	}

	discord.AddHandler(message)
	discord.Identify.Intents = discordgo.IntentsAll

	err = discord.Open()
	defer discord.Close()
	if err != nil {
		fmt.Println("error opening connection:", err)
		return
	}
	fmt.Println("Costanza is now running 😁")
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

func Log(m *discordgo.Message, err error) {
	if err != nil {
		logrus.Warnf("Error sending message: %s", err)
	} else {
		logrus.Tracef("Message sent %v", *m)
	}
}

func parseMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	message := m.Content
	command := strings.TrimSpace(message[1:])
	if i := strings.Index(message, " "); i != -1 {
		command = message[1:i]
	}
	logrus.Infof("Command received %s", command)
	send := sendClosure(s, m)

	switch command {
	case "listen":
		// connectToFirstVoiceChannel(event.getGuild().getAudioManager(), true);
	case "b", "back":
		player.Previous(m.GuildID)
	case "pn", "playnext":
		player.Play(s, m)
	case "p", "play":
		if strings.Contains(message, "ttt") {
			if len(m.Mentions) != 1 {
				send("Too many users mentioned, please specify only one person you would like to play against.")
				return
			}
			toe = ttt.New(*m.Author, *m.Mentions[0])
			PlayingTTT = true
			send(toe.String())
			return
		}
		player.QueueTrack(s, m)
	case "q", "queue":
		player.PrintQueue(m.ChannelID, m.GuildID)
	case "skip":
		player.Skip(s, m.GuildID)
	case "pause":
		player.Pause(m.GuildID)
	case "unpause":
		player.UnPause(m.GuildID)
	case "speak":
		if m.Author.ID != Isaac {
			return
		}
		err := s.ChannelMessageDelete(m.ChannelID, m.ID)
		if err != nil {
			logrus.Warnf("Couldn't delete message %s error: %s", m.ID, err)
		}
		Log(s.ChannelMessageSendTTS(m.ChannelID, util.AfterCommand(message)))
	case "send":
		filename := util.AfterCommand(message)
		f, err := os.Open(fmt.Sprintf("%s/%s", Images, filename))
		if err != nil {
			logrus.Warnf("Cannot open file %s", filename)
			return
		}
		r, err := s.ChannelFileSend(m.ChannelID, filename, f)
		if err != nil {
			logrus.Warnf("Couldn't send file %s, error : %s", filename, err)
		}
		logrus.Trace(r)
	case "translate":
		google.Translate(sendClosure(s, m), message)
	}
}

func sendClosure(s *discordgo.Session, m *discordgo.MessageCreate) func(string) {
	return func(send string) { Log(s.ChannelMessageSend(m.ChannelID, send)) }
}

func handlePrivateMessage(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	message := m.Content
	if m.GuildID != "" || m.Author.ID != Isaac {
		return false
	}

	logrus.Infof("Received Private Message {%s}\n", message)
	userIDAndMessage := strings.Split(util.AfterCommand(message), ":")
	dm, err := s.UserChannelCreate(userIDAndMessage[0])
	if err != nil {
		logrus.Errorf("Couldn't create channel %s", err)
		return false
	}
	_, err = s.ChannelMessageSend(dm.ID, userIDAndMessage[1])
	if err != nil {
		logrus.Warnf("Error sending dm %s", err)
	}
	return true
}

func message(s *discordgo.Session, m *discordgo.MessageCreate) {
	m.Content = strings.TrimSpace(m.Content)
	message := m.Content
	if m.Author.ID == s.State.User.ID {
		return
	}
	logrus.Tracef("Received Message {%s}", message)
	send := sendClosure(s, m)

	if handlePrivateMessage(s, m) {
		return
	}

	if strings.HasPrefix(message, "~") {
		parseMessage(s, m)
		return
	}

	if PlayingTTT && toe.IsPlaying(*m.Author) && Coordinate.MatchString(message) {
		var x, y int
		fmt.Sscanf(message, "%d,%d", &x, &y)
		result, finished := toe.Move(x, y, *m.Author)
		send(result)
		PlayingTTT = !finished
	}

	if message == "Goodbye Costanza" {
		send("See you nerds later")
		sc <- os.Interrupt
		return
	}

	for _, p := range Profanity {
		if p.MatchString(message) {
			send("Watch your profanity " + m.Author.Username)
		}
	}

	if rand.Float64() < 0.001 {
		send("Hello There @" + m.Author.Username)
	}
}
