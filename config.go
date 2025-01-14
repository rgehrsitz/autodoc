// autodoc/config.go

package main

import (
	"errors"
	"os"
)

type Config struct {
	OpenAIKey    string
	ProjectName  string
	ProjectURL   string
	Theme        string
	CustomStyles map[string]string
}

func LoadConfig() (*Config, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, errors.New("OPENAI_API_KEY environment variable is not set")
	}

	return &Config{
		OpenAIKey:    apiKey,
		ProjectName:  "AutoDoc",                              // Default value or load from environment
		ProjectURL:   "https://github.com/rgehrsitz/AutoDoc", // Default value or load from environment
		Theme:        "light",                                // Default value
		CustomStyles: map[string]string{},                    // Initialize as needed
	}, nil
}
