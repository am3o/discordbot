package main

import (
	"os"
	"time"

	"github.com/am3o/discordbot/pkg/collector"
	"github.com/sirupsen/logrus"

	"github.com/am3o/discordbot/pkg/service"
)

func main() {
	logger := logrus.StandardLogger()

	dictionary, _ := os.LookupEnv("DICTIONARY")
	token, ok := os.LookupEnv("DISCORD_TOKEN")
	if !ok {
		logger.Fatal("Token not found")
		panic("The token is required for the service")
	}

	bot, err := service.New(token,
		service.Dictionary(dictionary),
		service.Jokes(),
		service.PinnedMessages(token, time.Second*30),
		service.Collector(collector.New()),
		service.Logger(logger),
	)
	if err != nil {
		logger.WithError(err).Error("Could not initialize the bot")
		panic(err)
	}
	defer bot.Close()

	logrus.Info("Start the bot")
	if err := bot.ListenAndServe(); err != nil {
		logger.WithError(err).Error("Could not listen any more to the discord session ")
		return
	}

	logger.Info("Discord bot is stopped")
}
