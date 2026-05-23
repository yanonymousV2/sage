package ai

import (
	"fmt"
	"os"
)

// Backend is implemented by every AI provider.
type Backend interface {
	Complete(model, prompt string) (string, error)
	Stream(model, prompt string, onToken func(string)) (string, error)
}

// New returns the Backend for the given provider.
// API keys are read from the config value first, then from environment variables.
func New(provider, configAPIKey string) (Backend, error) {
	switch provider {
	case "ollama", "":
		return &OllamaBackend{}, nil

	case "claude":
		key := configAPIKey
		if key == "" {
			key = os.Getenv("ANTHROPIC_API_KEY")
		}
		if key == "" {
			return nil, fmt.Errorf("Claude requires an API key — set ANTHROPIC_API_KEY or run: sage config --api-key <key>")
		}
		return &ClaudeBackend{apiKey: key}, nil

	case "openai":
		key := configAPIKey
		if key == "" {
			key = os.Getenv("OPENAI_API_KEY")
		}
		if key == "" {
			return nil, fmt.Errorf("OpenAI requires an API key — set OPENAI_API_KEY or run: sage config --api-key <key>")
		}
		return &OpenAIBackend{apiKey: key}, nil

	default:
		return nil, fmt.Errorf("unknown provider %q — use: ollama, claude, openai", provider)
	}
}
