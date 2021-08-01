package operations

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/am3o/discordbot/pkg/client"
)

type PinnedMessagesOperator struct {
	sync.Mutex
	client *client.Discord
	cached map[string][]string
}

func NewPinnedMessagesOperator(client *client.Discord) *PinnedMessagesOperator {
	return &PinnedMessagesOperator{
		client: client,
		cached: make(map[string][]string),
	}
}

func (operator *PinnedMessagesOperator) Run(duration time.Duration) {
	ticker := time.NewTicker(duration)
	for ; ; <-ticker.C {
		for channelID := range operator.cached {
			pinnedMessages, err := operator.client.GetPinned(channelID)
			if err != nil {
				break
			}

			operator.Lock()
			operator.cached[channelID] = pinnedMessages
			operator.Unlock()
		}
	}
}

func (operator *PinnedMessagesOperator) Exec(channelID string) (string, error) {
	operator.Lock()
	defer operator.Unlock()

	_, exists := operator.cached[channelID]
	if !exists {
		pinnedMessages, err := operator.client.GetPinned(channelID)
		if err != nil {
			return "", err
		}

		if len(pinnedMessages) == 0 {
			return "", fmt.Errorf("no pinned messages detected")
		}

		operator.cached[channelID] = pinnedMessages
	}

	pinnedMessage := operator.cached[channelID][rand.Int()%len(operator.cached[channelID])]
	if pinnedMessage == "" {
		pinnedMessage = "undefined"
	}

	return pinnedMessage, nil
}
