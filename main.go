package main

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/am3o/discordbot/pkg/service"
	flag "github.com/spf13/pflag"
)

func main() {
	logger := logrus.StandardLogger()

	var DictonaryPath = flag.String("dictionary", "./resources/dictonary.json", "path to the dictionary")
	flag.Parse()

	token, exists := os.LookupEnv("DISCORD_BOT_TOKEN")
	if !exists {
		logger.Fatal("Could not start discord bot without any token")
	}

	bot, err := service.New(token, *DictonaryPath, logger)
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
