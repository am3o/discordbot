package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
)

type Joke struct {
	Content string `json:"joke"`
}

type Joker struct {
	url string
}

func NewJoker() Joker {
	return Joker{
		url: "https://official-joke-api.appspot.com/jokes",
	}
}

func (client Joker) GetRandomJoke(ctx context.Context) (Joke, error) {
	URL, err := url.Parse(client.url)
	if err != nil {
		return Joke{}, fmt.Errorf("could not parse joke api URL: %w", err)
	}
	URL.Path = path.Join(URL.Path, "random")

	req, err := http.NewRequest(http.MethodGet, URL.String(), nil)
	if err != nil {
		return Joke{}, fmt.Errorf("could not create joke api request: %w", err)
	}
	req.WithContext(ctx)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Joke{}, fmt.Errorf("could not request the joke api: %w", err)
	}
	defer resp.Body.Close()

	var joke Joke
	if err := json.NewDecoder(resp.Body).Decode(&joke); err != nil {
		return Joke{}, fmt.Errorf("could not unmarshal the response body: %w", err)
	}

	return joke, nil
}
