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

const openaiURL = "https://api.openai.com/v1/chat/completions"

var openaiClient = &http.Client{Timeout: 120 * time.Second}

// OpenAIBackend implements Backend for the OpenAI API.
type OpenAIBackend struct {
	apiKey string
}

func (b *OpenAIBackend) Complete(model, prompt string) (string, error) {
	body, err := json.Marshal(map[string]any{
		"model":       model,
		"temperature": 0,
		"max_tokens":  2048,
		"messages":    []map[string]string{{"role": "user", "content": prompt}},
	})
	if err != nil {
		return "", fmt.Errorf("could not build request: %w", err)
	}

	req, _ := http.NewRequest("POST", openaiURL, bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+b.apiKey)
	req.Header.Set("content-type", "application/json")

	resp, err := openaiClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("could not reach OpenAI API: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", fmt.Errorf("could not decode OpenAI response")
	}
	if result.Error.Message != "" {
		return "", fmt.Errorf("openai: %s", result.Error.Message)
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("openai returned empty response")
	}

	return result.Choices[0].Message.Content, nil
}

func (b *OpenAIBackend) Stream(model, prompt string, onToken func(string)) (string, error) {
	body, err := json.Marshal(map[string]any{
		"model":       model,
		"temperature": 0,
		"max_tokens":  2048,
		"stream":      true,
		"messages":    []map[string]string{{"role": "user", "content": prompt}},
	})
	if err != nil {
		return "", fmt.Errorf("could not build request: %w", err)
	}

	req, _ := http.NewRequest("POST", openaiURL, bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+b.apiKey)
	req.Header.Set("content-type", "application/json")

	resp, err := openaiClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("could not reach OpenAI API: %w", err)
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
		if data == "[DONE]" || data == "" {
			break
		}

		var event struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}
		if event.Error.Message != "" {
			return full.String(), fmt.Errorf("openai: %s", event.Error.Message)
		}
		if len(event.Choices) > 0 && event.Choices[0].Delta.Content != "" {
			token := event.Choices[0].Delta.Content
			if onToken != nil {
				onToken(token)
			}
			full.WriteString(token)
		}
	}

	return full.String(), scanner.Err()
}
