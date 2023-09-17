package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"

	"github.com/isaacpd/costanza/pkg/chat"
	"github.com/isaacpd/costanza/pkg/google"
	"github.com/isaacpd/costanza/pkg/router"
	"github.com/isaacpd/costanza/pkg/sound/player"
)

var (
	Token     string
	Verbosity string
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot token")
	flag.StringVar(&Verbosity, "v", "info", "Verbosity level")
	router.Sc = make(chan os.Signal, 1)
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
	err = chat.Init()
	if err != nil {
		logrus.Fatalf("Could not connect to ChatService: %v", err)
	}

	if Token == "" {
		Token = os.Getenv("COSTANZA_TOKEN")
	}
	Token = strings.TrimSpace(Token)
	discord, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("Error creating bot:", err)
	}
	discord.State.MaxMessageCount = 10

	router.RegisterCommands(discord)
	discord.AddHandler(router.HandleMessage)
	discord.AddHandler(player.MessageReact)
	discord.AddHandler(player.MessageRemove)
	discord.AddHandler(router.HandleCommand)
	discord.Identify.Intents = discordgo.IntentsAll

	err = discord.Open()
	defer discord.Close()
	if err != nil {
		logrus.Fatalf("error opening connection: %v", err)
		return
	}
	fmt.Println("Costanza is now running 😁")
	signal.Notify(router.Sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-router.Sc
}
