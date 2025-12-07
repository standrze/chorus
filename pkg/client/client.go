package client

import (
	"context"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type Client interface {
	ChatCompletion(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error)
}

type OpenAIClient struct {
	client *openai.Client
}

func NewClient(baseURL string) *OpenAIClient {
	client := openai.NewClient(option.WithBaseURL(baseURL))
	return &OpenAIClient{
		client: &client,
	}
}

func (c *OpenAIClient) ChatCompletion(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
	return c.client.Chat.Completions.New(ctx, params)
}
