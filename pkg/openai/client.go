// autodoc/pkg/openai/client.go
package openai

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
	"github.com/rgehrsitz/AutoDoc/pkg/chunk"
)

// Client handles interactions with the OpenAI API.
type Client struct {
	client  *openai.Client
	Chunker *chunk.Chunker
}

// NewClient initializes and returns a new OpenAI client.
func NewClient(apiKey string) *Client {
	return &Client{
		client: openai.NewClient(
			option.WithAPIKey(apiKey),
		),
		Chunker: chunk.NewChunker(4000), // Reasonable chunk size for GPT-4
	}
}

// AnalyzeSource analyzes and documents the provided source code with language-specific prompts.
func (c *Client) AnalyzeSource(ctx context.Context, code string, language string) (string, error) {
	chunks := c.Chunker.Split(code)
	var analyses []string

	for _, ck := range chunks {
		prompt := "Please analyze and document this code segment"
		if language != "" {
			// Capitalize the first letter of the language for better readability
			prompt += " written in " + strings.Title(language) + "."
		}
		prompt += " (lines " + fmt.Sprintf("%d-%d", ck.StartLine, ck.EndLine) + "):\n" + ck.Content

		resp, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
			Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
				openai.UserMessage(prompt),
			}),
			Model: openai.F(openai.ChatModelChatgpt4oLatest),
		})
		if err != nil {
			if strings.Contains(err.Error(), "rate limit") {
				log.Println("Rate limit reached. Please try again later.")
			}
			return "", err
		}
		analyses = append(analyses, resp.Choices[0].Message.Content)
	}

	return strings.Join(analyses, "\n\n"), nil
}

// GenerateEmbedding generates an embedding for the given text using the latest openai-go library.
func (c *Client) GenerateEmbedding(ctx context.Context, text string) ([]float64, error) {
	resp, err := c.client.Embeddings.New(ctx, openai.EmbeddingNewParams{
		// Because the library defines a “Field” type for Model, we wrap
		// the string constant in openai.F(...).
		Model: openai.F(openai.EmbeddingModelTextEmbeddingAda002),

		// Input must be a union type. We convert our string to a union
		// with shared.UnionString(...), then wrap it in openai.F[...]() again.
		Input: openai.F[openai.EmbeddingNewParamsInputUnion](
			shared.UnionString(text),
		),

		// Optionally set other fields (like Dimensions, EncodingFormat, etc.)
		// Dimensions:     openai.F(int64(1)),
		// EncodingFormat: openai.F(openai.EmbeddingNewParamsEncodingFormatFloat),
		// User:           openai.F("user-1234"),
	})
	if err != nil {
		return nil, fmt.Errorf("error calling Embeddings.New: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data returned")
	}

	// Return the embedding for the first (and only) input text
	return resp.Data[0].Embedding, nil
}
