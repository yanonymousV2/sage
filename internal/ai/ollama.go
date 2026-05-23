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

const ollamaURL = "http://localhost:11434/api/generate"

var ollamaClient = &http.Client{
	Timeout: 120 * time.Second,
}

type OllamaRequest struct {
	Model       string  `json:"model"`
	Prompt      string  `json:"prompt"`
	Stream      bool    `json:"stream"`
	Temperature float64 `json:"temperature"`
}

type OllamaResponse struct {
	Response string `json:"response"`
	Error    string `json:"error,omitempty"`
}

func friendlyError(model, msg string) error {
	lower := strings.ToLower(msg)
	if strings.Contains(lower, "not found") || strings.Contains(lower, "unknown model") {
		return fmt.Errorf("model %q is not pulled — run: ollama pull %s", model, model)
	}
	return fmt.Errorf("ollama: %s", msg)
}

// AskOllama makes a blocking request and returns the full response.
// Used for structured calls (command planning, grading) that need complete output.
func AskOllama(model, prompt string) (string, error) {
	reqBody, err := json.Marshal(OllamaRequest{
		Model:       model,
		Prompt:      prompt,
		Stream:      false,
		Temperature: 0,
	})
	if err != nil {
		return "", fmt.Errorf("could not build request: %w", err)
	}

	resp, err := ollamaClient.Post(ollamaURL, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("could not reach Ollama — is it running? Start with: ollama serve")
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return "", fmt.Errorf("could not decode Ollama response")
	}
	if ollamaResp.Error != "" {
		return "", friendlyError(model, ollamaResp.Error)
	}

	return ollamaResp.Response, nil
}

// AskOllamaStream streams tokens via onToken as they arrive and returns the full accumulated response.
// Used for the explanation call so output appears live in the terminal.
func AskOllamaStream(model, prompt string, onToken func(string)) (string, error) {
	reqBody, err := json.Marshal(OllamaRequest{
		Model:       model,
		Prompt:      prompt,
		Stream:      true,
		Temperature: 0,
	})
	if err != nil {
		return "", fmt.Errorf("could not build request: %w", err)
	}

	resp, err := ollamaClient.Post(ollamaURL, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("could not reach Ollama — is it running? Start with: ollama serve")
	}
	defer resp.Body.Close()

	var chunk struct {
		Response string `json:"response"`
		Done     bool   `json:"done"`
		Error    string `json:"error"`
	}

	var full strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB — default 64KB too small for 14b model
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		if err := json.Unmarshal([]byte(line), &chunk); err != nil {
			continue
		}
		if chunk.Error != "" {
			return full.String(), friendlyError(model, chunk.Error)
		}
		if chunk.Response != "" {
			if onToken != nil {
				onToken(chunk.Response)
			}
			full.WriteString(chunk.Response)
		}
		if chunk.Done {
			break
		}
	}

	return full.String(), scanner.Err()
}
