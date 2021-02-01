package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type Joke struct {
	Content string `json:"joke"`
}

type Joker struct {
	url string
}

func NewJoker() Joker {
	return Joker{
		url: "https://icanhazdadjoke.com/",
	}
}

func (client Joker) GetRandomJoke(ctx context.Context) (Joke, error) {
	URL, err := url.Parse(client.url)
	if err != nil {
		return Joke{}, fmt.Errorf("could not parse joke api URL: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, URL.String(), nil)
	if err != nil {
		return Joke{}, fmt.Errorf("could not create joke api request: %w", err)
	}
	req.WithContext(ctx)
	req.Header.Add("Accept", "application/json")

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
