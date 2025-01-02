package main

import (
	"errors"
	"os"
)

type Config struct {
	OpenAIAPIKey string
}

func LoadConfig() (*Config, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, errors.New("OPENAI_API_KEY environment variable is not set")
	}

	return &Config{
		OpenAIAPIKey: apiKey,
	}, nil
}
