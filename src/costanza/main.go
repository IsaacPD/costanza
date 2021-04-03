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
	"github.com/isaacpd/costanza/pkg/sound"
	"github.com/isaacpd/costanza/pkg/util"
)

var (
	Isaac    = "217795612169601024"
	Costanza = "319980316309848064"

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
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot token")
	flag.StringVar(&Verbosity, "v", "info", "Verbosity level")
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
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

func parseMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	message := m.Content
	command := message[1:strings.Index(message, " ")]
	logrus.Infof("Command received %s\n", command)

	switch command {
	case "listen":
		// connectToFirstVoiceChannel(event.getGuild().getAudioManager(), true);
	case "back":
		// prevTrack(event.getChannel());
	case "play":
		if strings.Contains(message, "ttt") {
			if len(m.Mentions) != 1 {
				s.ChannelMessageSend(m.ChannelID, "Too many users mentioned, please specify only one person you would like to play against.")
				return
			}
			toe = ttt.New(*m.Author, *m.Mentions[0])
			PlayingTTT = true
			s.ChannelMessageSend(m.ChannelID, toe.String())
			return
		}
		sound.Play(s, m)
	case "skip":
		// skipTrack(event.getChannel());
	case "speak":
		// event.getChannel().sendMessage(afterCommand(message)).complete();
	case "send":
		filename := util.AfterCommand(message)
		f, err := os.Open(fmt.Sprintf("%s/%s", Images, filename))
		if err != nil {
			logrus.Warnf("Cannot open file %s\n", filename)
			return
		}
		r, err := s.ChannelFileSend(m.ChannelID, filename, f)
		if err != nil {
			logrus.Warnf("Couldn't send file %s, error : %s\n", filename, err)
		}
		logrus.Trace(r)
	case "translate":
		google.Translate(sendClosure(s, m), message)
	}
}

func sendClosure(s *discordgo.Session, m *discordgo.MessageCreate) func(string) {
	return func(send string) { s.ChannelMessageSend(m.ChannelID, send) }
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
	message := m.Content
	if m.Author.ID == Costanza {
		return
	}
	logrus.Tracef("Received Message {%s}\n", message)

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
		s.ChannelMessageSend(m.ChannelID, result)
		PlayingTTT = !finished
	}

	if message == "Goodbye Costanza" {
		s.ChannelMessageSend(m.ChannelID, "See you nerds later")
		os.Exit(1)
	}

	for _, p := range Profanity {
		if p.MatchString(message) {
			s.ChannelMessageSendTTS(m.ChannelID, "Watch your profanity "+m.Author.Username)
		}
	}

	if rand.Float64() < 0.001 {
		s.ChannelMessageSend(m.ChannelID, "Hello There @"+m.Author.Username)
	}
}
