package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaResponse struct {
	Response string `json:"response"`
}

func AskOllama(model string, prompt string) (string, error) {

	reqBody, err := json.Marshal(OllamaRequest{
		Model:  model,
		Prompt: prompt,
		Stream: false,
	})

	resp, err := http.Post(
		"http://localhost:11434/api/generate",
		"application/json",
		bytes.NewBuffer(reqBody),
	)

	if err != nil {
		return "", fmt.Errorf("could not reach Ollama — is it running? Start it with: ollama serve")
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return "", fmt.Errorf("could not decode Ollama response")
	}

	return ollamaResp.Response, nil
}
