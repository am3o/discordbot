package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
)

type JokeResponse struct {
	Setup     string `json:"setup"`
	Punchline string `json:"punchline"`
}

type Joker struct {
	url string
}

func NewJoker() Joker {
	return Joker{
		url: "https://official-joke-api.appspot.com/jokes",
	}
}

func (client Joker) GetRandomJoke(ctx context.Context) (JokeResponse, error) {
	url, err := url.Parse(client.url)
	if err != nil {
		return JokeResponse{}, fmt.Errorf("could not parse joke api url: %w", err)
	}
	url.Path = path.Join(url.Path, "random")

	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return JokeResponse{}, fmt.Errorf("could not create joke api request: %w", err)
	}
	req.WithContext(ctx)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return JokeResponse{}, fmt.Errorf("could not request the joke api: %w", err)
	}
	defer resp.Body.Close()

	var joke JokeResponse
	if err := json.NewDecoder(resp.Body).Decode(&joke); err != nil {
		return JokeResponse{}, fmt.Errorf("could not unmarshal the response body: %w", err)
	}

	return joke, nil
}
