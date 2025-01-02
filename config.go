// autodoc/config.go
package main

import (
	"errors"
	"os"
)

type Config struct {
	OpenAIKey string
}

func LoadConfig() (*Config, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, errors.New("OPENAI_API_KEY environment variable is not set")
	}

	return &Config{
		OpenAIKey: apiKey,
	}, nil
}
