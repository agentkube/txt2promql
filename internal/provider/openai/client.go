// internal/provider/openai/client.go
package openai

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

type OpenAIClient struct {
	openai_client *openai.Client
	model         string
	temp          float32
}

func NewClient(apiKey string, model string, temperature float32) *OpenAIClient {
	return &OpenAIClient{
		openai_client: openai.NewClient(apiKey),
		model:         model,
		temp:          temperature,
	}
}

func (c *OpenAIClient) Complete(ctx context.Context, prompt string) (string, error) {
	resp, err := c.openai_client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: c.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    "user",
					Content: prompt,
				},
			},
			Temperature: c.temp,
		},
	)
	if err != nil {
		return "", fmt.Errorf("AI completion error: %w", err)
	}
	return resp.Choices[0].Message.Content, nil
}
