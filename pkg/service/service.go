package service

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"

	discord "github.com/bwmarrin/discordgo"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"gopkg.in/square/go-jose.v2/json"
)

type Dictionary map[string][]string

var DefaultDictionary = map[string][]string{"foo": {"bar"}}

type BotCollector interface {
	prometheus.Collector
	TrackMessage(string, string)
}

type Service struct {
	logger     logrus.FieldLogger
	collector  BotCollector
	dictionary Dictionary
	session    *discord.Session
	addr       string
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
			return fmt.Errorf("could not read the dictionary: %w", err)
		}
		defer reader.Close()

		if err := json.NewDecoder(reader).Decode(&service.dictionary); err != nil {
			return fmt.Errorf("could not unmarshal the dictionary: %w", err)
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

func Collector(collector BotCollector) Option {
	return func(service *Service) error {
		service.collector = collector
		return prometheus.Register(collector)
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
		logger:     logrus.StandardLogger(),
		dictionary: DefaultDictionary,
		session:    session,
		addr:       ":8080",
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
	defer service.TrackRequest(s, m)

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

func (service *Service) TrackRequest(s *discord.Session, m *discord.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	var channelName = "undefined"
	defer func() {
		service.collector.TrackMessage(channelName, m.Author.Username)
	}()

	channel, err := s.Channel(m.ChannelID)
	if err != nil {
		return
	}
	channelName = channel.Name
}

// QuoteMessage returns famous words of some persons out of the dictionary
func (service *Service) QuoteMessage(message string) ([]string, error) {
	var quotes = make([]string, 0)
	for buzzword, values := range service.dictionary {
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

	for command := range service.dictionary {
		commands = append(commands, command)
	}

	return commands
}

func (service *Service) sendMessages(s *discord.Session, m *discord.MessageCreate, messages ...string) {
	for _, message := range messages {
		go func(content string) {
			if m.Author.ID == s.State.User.ID {
				return
			}

			_, err := s.ChannelMessageSend(m.ChannelID, message)
			if err != nil {
				service.logger.WithError(err).Error("Could not send message")
			}
		}(message)
	}
}

// ListenAndServe listen on the tcp connection
func (s *Service) ListenAndServe() error {
	http.HandleFunc("/internal/metrics", promhttp.Handler().ServeHTTP)
	http.HandleFunc("/internal/health", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})

	s.logger.WithField("address", s.addr).Info("Service is still running")
	return http.ListenAndServe(s.addr, nil)
}
