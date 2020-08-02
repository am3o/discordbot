package service

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"

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
	TrackMessage(channel, userHandle string)
	TrackBotUsage(channel, userHandle string)
}

type TextFormatter interface {
	Format(quote, source string) string
}

type Service struct {
	logger     logrus.FieldLogger
	collector  BotCollector
	formatter  TextFormatter
	dictionary Dictionary
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

// Dictonary option for the service
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
		message := srv.HelpMessage()
		srv.sendMessages(s, m, message...)
	default:
		if strings.Contains(message, "!") {
			quotes, err := srv.QuoteMessage(message)
			if err != nil {
				srv.logger.WithError(err).Error("Could not find any quote")
				return
			}

			srv.sendMessages(s, m, quotes...)
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

// QuoteMessage returns famous words of some persons out of the dictionary
func (srv *Service) QuoteMessage(message string) ([]string, error) {
	var quotes = make([]string, 0)
	for buzzword, values := range srv.dictionary {
		if strings.Contains(message, fmt.Sprintf("!%v", strings.ToLower(buzzword))) {
			quote := values[rand.Int()%len(values)]
			quotes = append(quotes, srv.formatter.Format(quote, buzzword))
		}
	}

	if len(quotes) == 0 {
		return nil, fmt.Errorf("could not find any qoutes")
	}

	return quotes, nil
}

// HelpMessage returns all commands of the bot
func (srv *Service) HelpMessage() []string {
	var commands = []string{
		"The current buzzwords can be used by the bot",
	}

	for command := range srv.dictionary {
		commands = append(commands, command)
	}

	return commands
}

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
