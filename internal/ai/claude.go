package ai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const claudeURL = "https://api.anthropic.com/v1/messages"
const claudeVersion = "2023-06-01"

var claudeClient = &http.Client{Timeout: 120 * time.Second}

// ClaudeBackend implements Backend for the Anthropic Claude API.
type ClaudeBackend struct {
	apiKey string
}

func (b *ClaudeBackend) Complete(model, prompt string) (string, error) {
	body, err := json.Marshal(map[string]any{
		"model":      model,
		"max_tokens": 2048,
		"messages":   []map[string]string{{"role": "user", "content": prompt}},
	})
	if err != nil {
		return "", fmt.Errorf("could not build request: %w", err)
	}

	req, _ := http.NewRequest("POST", claudeURL, bytes.NewBuffer(body))
	req.Header.Set("x-api-key", b.apiKey)
	req.Header.Set("anthropic-version", claudeVersion)
	req.Header.Set("content-type", "application/json")

	resp, err := claudeClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("could not reach Claude API: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", fmt.Errorf("could not decode Claude response")
	}
	if result.Error.Message != "" {
		return "", fmt.Errorf("claude: %s", result.Error.Message)
	}
	if len(result.Content) == 0 {
		return "", fmt.Errorf("claude returned empty response")
	}

	return result.Content[0].Text, nil
}

func (b *ClaudeBackend) Stream(model, prompt string, onToken func(string)) (string, error) {
	body, err := json.Marshal(map[string]any{
		"model":      model,
		"max_tokens": 2048,
		"stream":     true,
		"messages":   []map[string]string{{"role": "user", "content": prompt}},
	})
	if err != nil {
		return "", fmt.Errorf("could not build request: %w", err)
	}

	req, _ := http.NewRequest("POST", claudeURL, bytes.NewBuffer(body))
	req.Header.Set("x-api-key", b.apiKey)
	req.Header.Set("anthropic-version", claudeVersion)
	req.Header.Set("content-type", "application/json")

	resp, err := claudeClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("could not reach Claude API: %w", err)
	}
	defer resp.Body.Close()

	var full strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "" {
			continue
		}

		var event struct {
			Type  string `json:"type"`
			Delta struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"delta"`
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}
		if event.Error.Message != "" {
			return full.String(), fmt.Errorf("claude: %s", event.Error.Message)
		}
		if event.Type == "content_block_delta" && event.Delta.Type == "text_delta" {
			if onToken != nil {
				onToken(event.Delta.Text)
			}
			full.WriteString(event.Delta.Text)
		}
		if event.Type == "message_stop" {
			break
		}
	}

	return full.String(), scanner.Err()
}
