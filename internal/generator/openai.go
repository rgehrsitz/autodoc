// autodoc/internal/generator/openai.go

package generator

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// OpenAIClient handles interactions with the OpenAI API.
type OpenAIClient struct {
	client *openai.Client
}

// NewOpenAIClient initializes and returns a new OpenAI client.
func NewOpenAIClient(apiKey string) *OpenAIClient {
	return &OpenAIClient{
		client: openai.NewClient(
			option.WithAPIKey(apiKey),
		),
	}
}

// AnalyzeSource analyzes and documents the provided source code.
func (c *OpenAIClient) AnalyzeSource(ctx context.Context, code string, language string) (string, error) {
	prompt := fmt.Sprintf("Please analyze this %s code and provide comprehensive documentation:\n\n%s",
		language, code)

	resp, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		}),
		Model: openai.F(openai.ChatModelChatgpt4oLatest),
	})
	if err != nil {
		return "", fmt.Errorf("failed to analyze code: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no completion choices returned")
	}

	return resp.Choices[0].Message.Content, nil
}
