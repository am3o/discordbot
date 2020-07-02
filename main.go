package main

import (
	"context"
	"flag"

	"github.com/sirupsen/logrus"

	"github.com/am3o/discordbot/pkg/service"
)

func main() {
	logger := logrus.StandardLogger()

	var DictonaryPath = flag.String("dictionary", "./resources/dictonary.json", "path to the dictionary")
	var Token = flag.String("token", "", "discord bot token")
	flag.Parse()

	if *Token == "" {
		logger.Fatal("Could not start discord bot without any token")
	}

	bot, err := service.New(*Token, *DictonaryPath, logger)
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
