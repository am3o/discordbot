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
	client.session.AddHandler(client.HandleMessageUpdate)

	return &client, nil
}

func (client *Discord) Close() error {
	if client.session != nil {
		return client.session.Close()
	}

	return nil
}

func (client *Discord) Ping() bool {
	// todo: check the state of the discord session
	return true
}

func (client *Discord) SendMessages(channelID, authorID string, messages ...string) {
	var wg sync.WaitGroup
	wg.Add(len(messages))

	for _, message := range messages {
		go func(content string) error {
			defer wg.Done()
			return client.SendMessage(channelID, authorID, content)
		}(message) //nolint:errcheck
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

func (client *Discord) HandleMessageUpdate(_ *discord.Session, m *discord.MessageUpdate) {
	client.handleMessage(m.Author.ID, m.ChannelID, m.Author.Username, strings.ToLower(m.Content))
}

func (client *Discord) HandleMessageCreate(_ *discord.Session, m *discord.MessageCreate) {
	client.handleMessage(m.Author.ID, m.ChannelID, m.Author.Username, strings.ToLower(m.Content))
}

func (client *Discord) handleMessage(userID, channelID, username, content string) {
	if userID == client.session.State.User.ID {
		return
	}

	for _, subscriber := range client.subscribers {
		subscriber.Publish(channelID, username, content)
	}
}
func (client *Discord) Author(id string) (string, error) {
	user, err := client.session.User(id)
	if err != nil {
		return "", err
	}
	return user.Username, nil
}

func (client *Discord) GetPinned(channelID string) ([]string, error) {
	channelMessagesPinned, err := client.session.ChannelMessagesPinned(channelID)
	if err != nil {
		return nil, fmt.Errorf("no pinned messages: %w", err)
	}

	var pinnedMessages []string
	for _, channelMessagePinned := range channelMessagesPinned {
		if channelMessagePinned.Content != "" {
			pinnedMessage := fmt.Sprintf("> %v \n > - %v \n", channelMessagePinned.Content, channelMessagePinned.Author.Username)
			if strings.Contains(channelMessagePinned.Content, "http") {
				pinnedMessage = channelMessagePinned.Content
			}

			pinnedMessages = append(pinnedMessages, pinnedMessage)
		}
	}

	return pinnedMessages, nil
}

type Publisher interface {
	Publish(channelID, authorID string, message string)
}

func (client *Discord) SubscribeMessageEvents(publisher Publisher) {
	client.subscribers = append(client.subscribers, publisher)
}
