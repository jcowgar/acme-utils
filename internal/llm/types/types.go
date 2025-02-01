package types

import "context"

// Message represents a chat message with standardized roles
type Message struct {
	Role    string // "user" or "assistant"
	Content string
}

// Provider defines the interface that all LLM providers must implement
type Provider interface {
	// Chat sends a conversation to the LLM and returns the response
	Chat(ctx context.Context, messages []Message) (string, error)
	
	// Name returns the provider's name for identification
	Name() string
}
