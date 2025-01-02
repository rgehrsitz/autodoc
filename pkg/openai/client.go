package openai

import (
	"context"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/chat"
)

type Client struct {
	client *openai.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		client: openai.NewClient(apiKey),
	}
}

func (c *Client) AnalyzeCode(ctx context.Context, code string) (string, error) {
	resp, err := c.client.Chat().Create(ctx, &chat.Request{
		Model: "gpt-4",
		Messages: []chat.Message{
			{
				Role:    "user",
				Content: "Please analyze and document this code:\n" + code,
			},
		},
	})
	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}
