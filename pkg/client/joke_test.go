package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJoker_GetRandomJoke(t *testing.T) {
	joker := NewJoker()
	joke, err := joker.GetRandomJoke(context.Background())
	{
		assert.NoError(t, err)
		assert.True(t, joke.Content != "")
	}
}
