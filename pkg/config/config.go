// autodoc/pkg/config/config.go

package config

import (
	"fmt"
	"os"
	"strings"
)

// Config holds the configuration settings for AutoDoc.
type Config struct {
	OpenAIKey    string
	ProjectName  string
	ProjectURL   string
	Theme        string
	CustomStyles map[string]string
}

// LoadConfig loads configuration from environment variables or a config file.
// For simplicity, this example uses environment variables.
func LoadConfig() (*Config, error) {
	openAIKey := os.Getenv("AUTODOC_OPENAI_KEY")
	if openAIKey == "" {
		return nil, fmt.Errorf("AUTODOC_OPENAI_KEY environment variable is not set")
	}

	projectName := os.Getenv("AUTODOC_PROJECT_NAME")
	if projectName == "" {
		projectName = "AutoDoc Project" // Default value
	}

	projectURL := os.Getenv("AUTODOC_PROJECT_URL")
	if projectURL == "" {
		projectURL = "https://example.com/project" // Default value
	}

	theme := os.Getenv("AUTODOC_THEME")
	if theme == "" {
		theme = "light" // Default theme
	}

	// Example for custom styles; this can be extended as needed
	customStyles := make(map[string]string)
	if styles := os.Getenv("AUTODOC_CUSTOM_STYLES"); styles != "" {
		// Expecting styles in key1=value1;key2=value2 format
		pairs := splitAndTrim(styles, ";")
		for _, pair := range pairs {
			kv := splitAndTrim(pair, "=")
			if len(kv) == 2 {
				customStyles[kv[0]] = kv[1]
			}
		}
	}

	return &Config{
		OpenAIKey:    openAIKey,
		ProjectName:  projectName,
		ProjectURL:   projectURL,
		Theme:        theme,
		CustomStyles: customStyles,
	}, nil
}

// splitAndTrim splits a string by the given separator and trims each part.
func splitAndTrim(s, sep string) []string {
	parts := []string{}
	for _, part := range strings.Split(s, sep) {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}
