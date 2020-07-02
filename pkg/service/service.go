package service

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"

	discord "github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
	"gopkg.in/square/go-jose.v2/json"
)

type Dictionary map[string][]string

type Service struct {
	logger    logrus.FieldLogger
	dictonary Dictionary
	session   *discord.Session
}

// New creates a new instance of the service, which creates a new discord session and manage them.
func New(token string, path string, logger logrus.FieldLogger) (Service, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return Service{}, fmt.Errorf("could not read the dictonary: %w", err)
	}

	var dictonary Dictionary
	if err := json.Unmarshal(data, &dictonary); err != nil {
		return Service{}, fmt.Errorf("could not unmarshal the dictonary: %w", err)
	}

	session, err := discord.New(fmt.Sprintf("Bot %s", token))
	if err != nil {
		return Service{}, fmt.Errorf("could not create new session: %w", err)
	}

	var service = Service{
		logger:    logger,
		dictonary: dictonary,
		session:   session,
	}

	if err := service.session.Open(); err != nil {
		return Service{}, fmt.Errorf("could not open the new session: %w", err)
	}

	service.session.AddHandler(service.HandleMessageCreate)

	return service, nil
}

// Close shut down the current discord session
func (service *Service) Close() {
	if err := service.session.Close(); err != nil {
		service.logger.WithError(err).Error("Could not successfully close the session")
		return
	}
	service.logger.Info("session successfully closed")
}

// HandleMessageCreate is the handler of a discord message event.
func (service *Service) HandleMessageCreate(s *discord.Session, m *discord.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	message := strings.ToLower(m.Content)
	switch {
	case strings.Contains(message, "!help") || strings.Contains(message, "!command"):
		service.HelpMessage(s, m)
	default:
		if strings.Contains(message, "!") {
			quote, err := service.QuoteMessage(message)
			if err != nil {
				service.logger.WithError(err).Error("Could not find any quote")
				return
			}

			if _, err := s.ChannelMessageSend(m.ChannelID, quote); err != nil {
				service.logger.WithError(err).WithField("channel", m.ChannelID).Error("Could not send message")
				return
			}
		}
	}
}

// QuoteMessage returns famous words of some persons out of the dictonary
func (service *Service) QuoteMessage(message string) (string, error) {
	for buzzword, values := range service.dictonary {
		if strings.Contains(message, fmt.Sprintf("!%v", strings.ToLower(buzzword))) {
			return fmt.Sprintf("> %v \n > - %v", values[rand.Int()%len(values)], buzzword), nil
		}
	}
	return "", fmt.Errorf("could not find any qoutes")
}

// HelpMessage returns all commands of the bot
func (service *Service) HelpMessage(s *discord.Session, m *discord.MessageCreate) {
	for buzzword := range service.dictonary {
		if _, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("- %v", buzzword)); err != nil {
			service.logger.WithError(err).WithField("channel", m.ChannelID).Error("Could not send message")
		}
	}
}

// ListenAndServe
func (Service) ListenAndServe(ctx context.Context) error {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	select {
	case <-ctx.Done():
		return fmt.Errorf("context closed")
	case <-sc:
		return fmt.Errorf("program interupted and killed")
	}
}
