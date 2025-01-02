// autodoc/pkg/openai/client.go
package openai

import (
	"context"
	"fmt"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/rgehrsitz/AutoDoc/pkg/chunk"
)

// Client handles interactions with the OpenAI API.
type Client struct {
	client  *openai.Client
	chunker *chunk.Chunker
}

// NewClient initializes and returns a new OpenAI client.
func NewClient(apiKey string) *Client {
	return &Client{
		client: openai.NewClient(
			option.WithAPIKey(apiKey),
		),
		chunker: chunk.NewChunker(4000), // Reasonable chunk size for GPT-4
	}
}

// AnalyzeSource analyzes and documents the provided source code with language-specific prompts.
func (c *Client) AnalyzeSource(ctx context.Context, code string, language string) (string, error) {
	chunks := c.chunker.Split(code)
	var analyses []string

	for _, chunk := range chunks {
		prompt := "Please analyze and document this code segment"
		if language != "" {
			// Capitalize the first letter of the language for better readability
			prompt += " written in " + strings.Title(language) + "."
		}
		prompt += " (lines " + fmt.Sprintf("%d-%d", chunk.StartLine, chunk.EndLine) + "):\n" + chunk.Content

		resp, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
			Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
				openai.UserMessage(prompt),
			}),
			Model: openai.F(openai.ChatModelGPT4),
		})
		if err != nil {
			return "", err
		}
		analyses = append(analyses, resp.Choices[0].Message.Content)
	}

	return strings.Join(analyses, "\n\n"), nil
}
