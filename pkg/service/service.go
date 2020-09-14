package service

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/am3o/discordbot/pkg/message"
	discord "github.com/bwmarrin/discordgo"
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

type TextFormatter interface {
	Format(quote, source string) string
}

type Entry struct {
	keyword  string
	detector message.KeywordDetector
	quotes   []string
}

type Service struct {
	logger     logrus.FieldLogger
	collector  BotCollector
	formatter  TextFormatter
	dictionary []Entry
	session    *discord.Session
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

		var dictonary []Entry
		for key, entry := range entries {
			dictonary = append(dictonary, Entry{
				detector: message.NewKeywordDetector(key),
				keyword:  key,
				quotes:   entry,
			})
		}

		service.dictionary = dictonary
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

// MessageFormatter option for the service
func MessageFormatter(formatter TextFormatter) Option {
	return func(service *Service) error {
		service.formatter = formatter
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
		formatter: message.DefaultTextFormatter,
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
func (srv *Service) Close() {
	if err := srv.session.Close(); err != nil {
		srv.logger.WithError(err).Error("Could not successfully close the session")
		return
	}

	srv.logger.Info("session successfully closed")
}

// HandleMessageCreate is the handler of a discord message event.
func (srv *Service) HandleMessageCreate(s *discord.Session, m *discord.MessageCreate) {
	defer srv.TrackRequest(s, m)

	message := strings.ToLower(m.Content)
	switch {
	case strings.Contains(message, "!help") || strings.Contains(message, "!command"):
		srv.sendMessages(s, m, srv.HelpMessage()...)
	default:
		if strings.Contains(message, "!") {
			var quotes = make([]string, 0)
			for _, entry := range srv.dictionary {
				if entry.detector.IsKeywordIncluded(message) {
					quote := entry.quotes[rand.Int()%len(entry.quotes)]
					quotes = append(quotes, srv.formatter.Format(quote, entry.keyword))
				}
			}

			if len(quotes) > 0 {
				srv.sendMessages(s, m, quotes...)
			}
		}
	}
}

func (srv *Service) TrackRequest(s *discord.Session, m *discord.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	var channelName = "undefined"
	defer func() {
		srv.collector.TrackMessage(channelName, m.Author.Username)
		srv.collector.TrackBotUsage(channelName, m.Author.Username)
	}()

	channel, err := s.Channel(m.ChannelID)
	if err != nil {
		return
	}
	channelName = channel.Name
}

// HelpMessage returns all commands of the bot
func (srv *Service) HelpMessage() []string {
	var message = []string{
		"The current buzzwords can be used by the bot",
	}

	for _, entry := range srv.dictionary {
		message = append(message, entry.keyword)
	}

	return message
}

// sendMessage sends the message to the discord server with the active session
func (srv *Service) sendMessages(s *discord.Session, m *discord.MessageCreate, messages ...string) {
	var wg sync.WaitGroup
	wg.Add(len(messages))

	for _, message := range messages {
		var content = message
		go func() {
			defer wg.Done()
			if m.Author.ID == s.State.User.ID {
				return
			}

			_, err := s.ChannelMessageSend(m.ChannelID, content)
			if err != nil {
				srv.logger.WithError(err).Error("Could not send message")
			}
		}()
	}

	wg.Wait()
}

// ListenAndServe listen on the tcp connection
func (srv *Service) ListenAndServe() error {
	http.HandleFunc("/internal/metrics", promhttp.Handler().ServeHTTP)
	http.HandleFunc("/internal/health", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})

	srv.logger.WithField("address", srv.addr).Info("Service is still running")
	return http.ListenAndServe(srv.addr, nil)
}
