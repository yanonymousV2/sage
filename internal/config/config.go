package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const DefaultModel = "qwen2.5:14b"

// DefaultModelFor returns a sensible default model for each provider.
var DefaultModelFor = map[string]string{
	"ollama": "qwen2.5:14b",
	"claude": "claude-sonnet-4-6",
	"openai": "gpt-4o-mini",
}

type Config struct {
	Provider      string `json:"provider"`
	Model         string `json:"model"`
	APIKey        string `json:"api_key,omitempty"`
	TrustReadOnly bool   `json:"trust_read_only"`
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".sage", "config.json"), nil
}

func Load() Config {
	p, err := configPath()
	if err != nil {
		return defaults()
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return defaults()
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return defaults()
	}
	if cfg.Model == "" {
		cfg.Model = DefaultModel
	}
	if cfg.Provider == "" {
		cfg.Provider = "ollama"
	}
	return cfg
}

func Save(cfg Config) error {
	p, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0644)
}

func Defaults() Config {
	return Config{
		Provider: "ollama",
		Model:    DefaultModel,
	}
}

func defaults() Config { return Defaults() }
