package client

import (
	"fmt"
	"strings"
	"sync"

	discord "github.com/bwmarrin/discordgo"
)

type Discord struct {
	subscribers []Publisher
	session     *discord.Session
}

func NewDiscord(token string) (*Discord, error) {
	session, err := discord.New(fmt.Sprintf("Bot %s", token))
	if err != nil {
		return nil, fmt.Errorf("could not create new discord session: %w", err)
	}

	if err := session.Open(); err != nil {
		return nil, fmt.Errorf("could not open the new session: %w", err)
	}

	client := Discord{
		session:     session,
		subscribers: make([]Publisher, 0),
	}

	client.session.AddHandler(client.HandleMessageCreate)

	return &client, nil
}

func (client Discord) Close() error {
	if client.session != nil {
		return client.session.Close()
	}

	return nil
}

func (client *Discord) SendMessages(channelID, authorID string, messages ...string) {
	var wg sync.WaitGroup
	wg.Add(len(messages))

	for _, message := range messages {
		go func(content string) error {
			defer wg.Done()
			return client.SendMessage(channelID, authorID, content)
		}(message)
	}

	wg.Wait()
}

func (client *Discord) SendMessage(channelID, authorID, content string) error {
	if authorID == client.session.State.User.ID {
		return nil
	}

	_, err := client.session.ChannelMessageSend(channelID, content)
	return err
}

func (client *Discord) HandleMessageCreate(_ *discord.Session, m *discord.MessageCreate) {
	if m.Author.ID == client.session.State.User.ID {
		return
	}

	for _, subscriber := range client.subscribers {
		subscriber.
			Publish(m.ChannelID, m.Author.Username, strings.ToLower(m.Content))
	}
}

func (client *Discord) Author(id string) (string, error) {
	user, err := client.session.User(id)
	if err != nil {
		return "", err
	}
	return user.Username, nil
}

type Publisher interface {
	Publish(channelID, authorID string, message string)
}

func (client *Discord) SubscribeMessageEvents(publisher Publisher) {
	client.subscribers = append(client.subscribers, publisher)
}
