package service

import (
	"fmt"
	"net/http"
	"os"
	"strings"

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
	logger     logrus.FieldLogger
	collector  BotCollector
	dictionary operations.QuotesOperator
	discord    *client.Discord
	addr       string
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
		srv.dictionary.String(),
	}
}

func (srv *Service) Publish(channel, author string, message string) {
	defer srv.TrackRequest(channel, author)

	switch {
	case strings.Contains(message, "!help") || strings.Contains(message, "!command"):
		srv.discord.SendMessages(channel, author, srv.HelpMessage()...)
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

		srv.discord.SendMessages(channel, author, quotes...)
	}
}

// ListenAndServe listen on the tcp connection
func (srv *Service) ListenAndServe() error {
	http.HandleFunc("/internal/metrics", promhttp.Handler().ServeHTTP)
	http.HandleFunc("/internal/health", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})

	return http.ListenAndServe(srv.addr, nil)
}
