package openai

import (
	"context"
	"fmt"

	"github.com/jcowgar/acme-utils/internal/config"
	"github.com/jcowgar/acme-utils/internal/llm/types"
	"github.com/sashabaranov/go-openai"
)

type Provider struct {
	client *openai.Client
	model  string
}

func New(model string, params map[string]interface{}) (*Provider, error) {
	apiKey, ok := params["api_key"].(string)
	if !ok {
		return nil, fmt.Errorf("api_key not found in config params")
	}
	apiKey = config.ExpandString(apiKey)

	openAiConfig := openai.DefaultConfig(apiKey)
	baseURL, ok := params["base_url"].(string)
	if ok {
		openAiConfig.BaseURL = baseURL
	}
	client := openai.NewClientWithConfig(openAiConfig)

	return &Provider{
		client: client,
		model:  model,
	}, nil
}

func (p *Provider) Name() string {
	return "openai"
}

func (p *Provider) Chat(ctx context.Context, messages []types.Message) (string, error) {
	// Convert messages to OpenAI format
	openaiMessages := make([]openai.ChatCompletionMessage, len(messages))
	for i, msg := range messages {
		role := msg.Role
		if role == "user" {
			role = "user"
		} else if role == "assistant" {
			role = "assistant"
		}

		openaiMessages[i] = openai.ChatCompletionMessage{
			Role:    role,
			Content: msg.Content,
		}
	}

	resp, err := p.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:    p.model,
			Messages: openaiMessages,
		},
	)
	if err != nil {
		return "", fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned")
	}

	return resp.Choices[0].Message.Content, nil
}
