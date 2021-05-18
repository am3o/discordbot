package operations

import (
	"fmt"
	"math/rand"

	"github.com/am3o/discordbot/pkg/client"
)

type PinnedMessagesOperator struct {
	client *client.Discord
}

func NewPinnedMessagesOperator(client *client.Discord) PinnedMessagesOperator {
	return PinnedMessagesOperator{
		client: client,
	}
}

func (operator *PinnedMessagesOperator) Exec(channelID string) (string, error) {
	pinnedMessages, err := operator.client.GetPinned(channelID)
	if err != nil {
		return "", err
	}

	if len(pinnedMessages) == 0 {
		return "", fmt.Errorf("no pinned messages detected")
	}

	return pinnedMessages[rand.Int()%len(pinnedMessages)], nil
}
