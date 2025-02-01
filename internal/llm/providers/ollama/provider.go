package ollama

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/jcowgar/acme-utils/internal/llm/types"
	ollamaapi "github.com/ollama/ollama/api"
)

type Provider struct {
	client *ollamaapi.Client
	model  string
}

func New(model string, params map[string]interface{}) (*Provider, error) {
	baseURL, ok := params["base_url"].(string)
	if !ok {
		return nil, fmt.Errorf("base_url not found in config params")
	}

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	client := ollamaapi.NewClient(parsedURL, http.DefaultClient)

	return &Provider{
		client: client,
		model:  model,
	}, nil
}

func (p *Provider) Name() string {
	return "ollama"
}

func (p *Provider) Chat(ctx context.Context, messages []types.Message) (string, error) {
	// Convert messages to Ollama format
	ollamaMessages := make([]ollamaapi.Message, len(messages))
	for i, msg := range messages {
		ollamaMessages[i] = ollamaapi.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	stream := false
	req := &ollamaapi.ChatRequest{
		Model:    p.model,
		Messages: ollamaMessages,
		Stream:   &stream,
		Options:  map[string]interface{}{"num_ctx": 8192},
	}

	var response *ollamaapi.ChatResponse
	responseHandler := func(r ollamaapi.ChatResponse) error {
		response = &r
		return nil
	}

	if err := p.client.Chat(ctx, req, responseHandler); err != nil {
		return "", fmt.Errorf("ollama chat failed: %w", err)
	}

	return response.Message.Content, nil
}
