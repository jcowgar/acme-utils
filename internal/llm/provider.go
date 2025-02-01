package llm

import (
	"github.com/jcowgar/acme-utils/internal/llm/types"
)

// Config holds the common configuration for any provider
type Config struct {
	Type   string                 `yaml:"type"`   // e.g., "ollama", "openai", etc.
	Model  string                 `yaml:"model"`
	Params map[string]interface{} `yaml:"params"` // Provider-specific parameters
}

// For convenience, expose the Message type from types package
type Message = types.Message

// For convenience, expose the Provider interface from types package
type Provider = types.Provider
