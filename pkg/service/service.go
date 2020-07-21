package service

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"

	discord "github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
	"gopkg.in/square/go-jose.v2/json"
)

type Dictionary map[string][]string

var DEFAULT_DICTONARY = map[string][]string{"status": {"offline"}}

type Service struct {
	logger    logrus.FieldLogger
	dictonary Dictionary
	session   *discord.Session
	addr      string
}

type Option func(*Service) error

func Logger(log logrus.FieldLogger) Option {
	return func(service *Service) error {
		service.logger = log
		return nil
	}
}

func Dictonary(path string) Option {
	return func(service *Service) error {
		reader, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("could not read the dictonary: %w", err)
		}

		if err := json.NewDecoder(reader).Decode(&service.dictonary); err != nil {
			return fmt.Errorf("could not unmarshal the dictonary: %w", err)
		}

		return nil
	}
}

func Address(addr string) Option {
	return func(service *Service) error {
		service.addr = addr
		return nil
	}
}

// New creates a new instance of the service, which creates a new discord session and manage them.
func New(token string, options ...Option) (Service, error) {
	session, err := discord.New(fmt.Sprintf("Bot %s", token))
	if err != nil {
		return Service{}, fmt.Errorf("could not create new session: %w", err)
	}

	if err := session.Open(); err != nil {
		return Service{}, fmt.Errorf("could not open the new session: %w", err)
	}

	var service = Service{
		logger:    logrus.StandardLogger(),
		dictonary: DEFAULT_DICTONARY,
		session:   session,
		addr:      ":8080",
	}

	for _, option := range options {
		if err := option(&service); err != nil {
			return Service{}, fmt.Errorf("could noit execute the option: %w", err)
		}
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
	message := strings.ToLower(m.Content)
	switch {
	case strings.Contains(message, "!help") || strings.Contains(message, "!command"):
		message := service.HelpMessage()
		service.sendMessages(s, m, message...)
	default:
		if strings.Contains(message, "!") {
			quotes, err := service.QuoteMessage(message)
			if err != nil {
				service.logger.WithError(err).Error("Could not find any quote")
				return
			}

			service.sendMessages(s, m, quotes...)
		}
	}
}

// QuoteMessage returns famous words of some persons out of the dictonary
func (service *Service) QuoteMessage(message string) ([]string, error) {
	var quotes = make([]string, 0)
	for buzzword, values := range service.dictonary {
		if strings.Contains(message, fmt.Sprintf("!%v", strings.ToLower(buzzword))) {
			quotes = append(quotes, fmt.Sprintf("> %v \n > - %v", values[rand.Int()%len(values)], buzzword))
		}
	}

	if len(quotes) == 0 {
		return nil, fmt.Errorf("could not find any qoutes")
	}

	return quotes, nil
}

// HelpMessage returns all commands of the bot
func (service *Service) HelpMessage() []string {
	var commands = []string{
		"The current buzzwords can be used by the bot",
	}

	for command := range service.dictonary {
		commands = append(commands, command)
	}

	return commands
}

func (service *Service) sendMessages(s *discord.Session, m *discord.MessageCreate, messages ...string) {
	for _, message := range messages {
		go func() {
			if m.Author.ID == s.State.User.ID {
				return
			}

			_, err := s.ChannelMessage(m.ChannelID, message)
			if err != nil {
				service.logger.WithError(err).Error("Could not send message")
			}
		}()
	}
}

// ListenAndServe listen on the tcp connection
func (s *Service) ListenAndServe() error {
	return http.ListenAndServe(s.addr, nil)
}
