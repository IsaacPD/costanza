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
	router.RegisterCommands()

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
	Token = strings.TrimSpace(Token)
	discord, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("Error creating bot:", err)
	}

	discord.AddHandler(router.HandleMessage)
	discord.AddHandler(player.MessageReact)
	discord.AddHandler(player.MessageRemove)
	discord.Identify.Intents = discordgo.IntentsAll

	err = discord.Open()
	defer discord.Close()
	if err != nil {
		fmt.Println("error opening connection:", err)
		return
	}
	fmt.Println("Costanza is now running 😁")
	signal.Notify(router.Sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-router.Sc
}
