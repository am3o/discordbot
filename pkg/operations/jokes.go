package operations

import (
	"context"
	"fmt"

	"github.com/am3o/discordbot/pkg/client"
)

type JokesOperator struct {
	client client.Joker
}

func NewJokeOperator(client client.Joker) JokesOperator {
	return JokesOperator{
		client: client,
	}
}

func (operator *JokesOperator) Exec(ctx context.Context) (string, error) {
	resp, err := operator.client.GetRandomJoke(ctx)
	if err != nil {
		return "", fmt.Errorf("could not create new joke: %w", err)
	}

	return fmt.Sprintf("> %v \n || %v ||", resp.Setup, resp.Punchline), nil
}
