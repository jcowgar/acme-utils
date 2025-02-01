package llm

import (
	"fmt"

	"github.com/jcowgar/acme-utils/internal/config"
	"github.com/jcowgar/acme-utils/internal/llm/providers/ollama"
	"github.com/jcowgar/acme-utils/internal/llm/providers/openai"
)

// NewProvider creates a new LLM provider based on the provider type
func NewProvider(providerType string, cfg config.ProviderConfig) (Provider, error) {
	switch providerType {
	case "ollama":
		model := cfg.Model
		params := cfg.Params
		return ollama.New(model, params)
	case "openai":
		model := cfg.Model
		params := cfg.Params
		return openai.New(model, params)
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", providerType)
	}
}
