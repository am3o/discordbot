package service

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/am3o/discordbot/pkg/client"
	"github.com/am3o/discordbot/pkg/operations"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"gopkg.in/square/go-jose.v2/json"
)

type BotCollector interface {
	prometheus.Collector
	TrackMessage(channel, userHandle string)
	TrackBotUsage(channel, userHandle string)
}

type Service struct {
	logger         logrus.FieldLogger
	collector      BotCollector
	dictionary     operations.QuotesOperator
	jokes          operations.JokesOperator
	pinnedMessages *operations.PinnedMessagesOperator
	discord        *client.Discord
	addr           string
}

// Option is an optional setting for the Service
type Option func(*Service) error

// Logger option for the service
func Logger(log logrus.FieldLogger) Option {
	return func(service *Service) error {
		service.logger = log
		return nil
	}
}

// Dictionary option for the service
func Dictionary(path string) Option {
	return func(service *Service) error {
		reader, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("could not read the dictionary: %w", err)
		}
		defer reader.Close()

		var entries map[string][]string
		if err := json.NewDecoder(reader).Decode(&entries); err != nil {
			return fmt.Errorf("could not unmarshal the dictionary: %w", err)
		}

		service.dictionary = operations.NewQuotesOperator(entries)
		return nil
	}
}

func PinnedMessages(token string, update time.Duration) Option {
	return func(service *Service) error {
		client, err := client.NewDiscord(token)
		if err != nil {
			return fmt.Errorf("could not initialize discord client for pinned messages: %w", err)
		}

		service.pinnedMessages = operations.NewPinnedMessagesOperator(client)
		go service.pinnedMessages.Run(update)
		return nil
	}
}

// Address option for the service
func Address(addr string) Option {
	return func(service *Service) error {
		service.addr = addr
		return nil
	}
}

// Collector option for the service
func Collector(collector BotCollector) Option {
	return func(service *Service) error {
		service.collector = collector
		return prometheus.Register(collector)
	}
}

// Jokes option for the service
func Jokes() Option {
	return func(service *Service) error {
		joker := client.NewJoker()
		service.jokes = operations.NewJokeOperator(joker)
		return nil
	}
}

// New creates a new instance of the service, which creates a new discord session and manage them.
func New(token string, options ...Option) (Service, error) {
	discord, err := client.NewDiscord(token)
	if err != nil {
		return Service{}, fmt.Errorf("could not create new session: %w", err)
	}

	service := Service{
		logger:  logrus.StandardLogger(),
		discord: discord,
		addr:    ":8080",
	}

	for _, option := range options {
		if err := option(&service); err != nil {
			return Service{}, fmt.Errorf("could noit execute the option: %w", err)
		}
	}

	discord.SubscribeMessageEvents(&service)

	return service, nil
}

// Close shut down the current discord session
func (srv *Service) Close() {
	if srv.discord != nil {
		if err := srv.discord.Close(); err != nil {
			srv.logger.WithError(err).Error("Cloud not shut down the discord session")
		}
	}
}

func (srv *Service) TrackRequest(channel, authorID string) {
	author, _ := srv.discord.Author(authorID)

	srv.collector.TrackMessage(channel, author)
	srv.collector.TrackBotUsage(channel, author)
}

// HelpMessage returns all commands of the bot
func (srv *Service) HelpMessage() []string {
	return []string{
		"The current buzzwords can be used by the bot",
		"joke",
		srv.dictionary.String(),
	}
}

func (srv *Service) Publish(channel, author string, message string) {
	defer srv.TrackRequest(channel, author)

	var response []string
	switch {
	case strings.Contains(message, "!help") || strings.Contains(message, "!command"):
		response = append(response, srv.HelpMessage()...)
	case strings.Contains(message, "!joke"):
		joke, err := srv.jokes.Exec(context.Background())
		if err != nil {
			srv.logger.WithFields(logrus.Fields{
				"message": message,
				"author":  author,
				"channel": channel,
			}).WithError(err).Error("could not create joke")
			return
		}
		response = append(response, joke)
	case strings.Contains(message, "!pin"):
		pinnedMessage, err := srv.pinnedMessages.Exec(channel)
		if err != nil {
			srv.logger.WithFields(logrus.Fields{
				"message": message,
				"author":  author,
				"channel": channel,
			}).WithError(err).Error("could not create joke")
			return
		}
		response = append(response, pinnedMessage)
	default:
		quotes := srv.dictionary.Exec(message)

		if len(quotes) == 0 {
			srv.logger.WithFields(logrus.Fields{
				"message": message,
				"author":  author,
				"channel": channel,
			}).Error("could not detect some quotes")
			return
		}
		response = append(response, quotes...)
	}

	srv.discord.SendMessages(channel, author, response...)
}

// ListenAndServe listen on the tcp connection
func (srv *Service) ListenAndServe() error {
	http.HandleFunc("/internal/metrics", promhttp.Handler().ServeHTTP)
	http.HandleFunc("/internal/health", func(writer http.ResponseWriter, request *http.Request) {
		if srv.discord.Ping() {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusOK)
	})

	return http.ListenAndServe(srv.addr, nil)
}
