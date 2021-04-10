package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"

	"github.com/isaacpd/costanza/pkg/google"
	"github.com/isaacpd/costanza/pkg/router"
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
	discord, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("Error creating bot:", err)
	}

	discord.AddHandler(router.HandleMessage)
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
