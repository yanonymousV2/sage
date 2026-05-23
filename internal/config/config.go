package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const DefaultModel = "qwen2.5:14b"

type Config struct {
	Model         string `json:"model"`
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
		return Config{Model: DefaultModel}
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return Config{Model: DefaultModel}
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil || cfg.Model == "" {
		return Config{Model: DefaultModel}
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
