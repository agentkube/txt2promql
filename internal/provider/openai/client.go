package openai

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/sashabaranov/go-openai"
)

const (
	maxToken         = 512
	presencePenalty  = 0.0
	frequencyPenalty = 0.0
)

type OpenAIClient struct {
	client      *openai.Client
	model       string
	temperature float32
	topP        float32
}

func NewClient(cfg IAIConfig) (*OpenAIClient, error) {
	config := openai.DefaultConfig(cfg.GetPassword())

	// Configure base URL if provided
	if baseURL := cfg.GetBaseURL(); baseURL != "" {
		config.BaseURL = baseURL
	}

	// Configure proxy if provided
	transport := &http.Transport{}
	if proxyEndpoint := cfg.GetProxyEndpoint(); proxyEndpoint != "" {
		proxyUrl, err := url.Parse(proxyEndpoint)
		if err != nil {
			return nil, fmt.Errorf("parsing proxy URL: %w", err)
		}
		transport.Proxy = http.ProxyURL(proxyUrl)
	}

	// Configure organization ID if provided
	if orgID := cfg.GetOrganizationId(); orgID != "" {
		config.OrgID = orgID
	}

	// Configure custom headers if provided
	customHeaders := cfg.GetCustomHeaders()
	config.HTTPClient = &http.Client{
		Transport: &headerTransport{
			origin:  transport,
			headers: customHeaders,
		},
	}

	client := openai.NewClientWithConfig(config)
	if client == nil {
		return nil, errors.New("failed to create OpenAI client")
	}

	return &OpenAIClient{
		client:      client,
		model:       cfg.GetModel(),
		temperature: cfg.GetTemperature(),
		topP:        cfg.GetTopP(),
	}, nil
}

func (c *OpenAIClient) Complete(ctx context.Context, prompt string) (string, error) {
	resp, err := c.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: c.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Temperature:      c.temperature,
			MaxTokens:        maxToken,
			PresencePenalty:  presencePenalty,
			FrequencyPenalty: frequencyPenalty,
			TopP:             c.topP,
		},
	)
	if err != nil {
		return "", fmt.Errorf("AI completion error: %w", err)
	}
	return resp.Choices[0].Message.Content, nil
}

// headerTransport adds custom headers to requests
type headerTransport struct {
	origin  http.RoundTripper
	headers []http.Header
}

func (t *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clonedReq := req.Clone(req.Context())
	for _, header := range t.headers {
		for key, values := range header {
			for _, value := range values {
				clonedReq.Header.Add(key, value)
			}
		}
	}
	if t.origin == nil {
		t.origin = http.DefaultTransport
	}
	return t.origin.RoundTrip(clonedReq)
}
