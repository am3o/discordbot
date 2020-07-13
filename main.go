package main

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/am3o/discordbot/pkg/service"
)

func main() {
	logger := logrus.StandardLogger()

	dictionary, ok := os.LookupEnv("DICTIONARY")
	if !ok {
		logger.Fatal("Resources not found")
	}

	token, ok := os.LookupEnv("TOKEN")
	if !ok {
		logger.Fatal("Token not found")
	}

	bot, err := service.New(token, dictionary, logger)
	if err != nil {
		logger.WithError(err).Error("Could not initialize the bot")
	}
	defer bot.Close()

	if err := bot.ListenAndServe(context.Background()); err != nil {
		logger.WithError(err).Error("Could not listen any more to the discord session ")
		return
	}

	logger.Info("Discord bot is stopped")
}
