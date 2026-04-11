package dlp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ==========================================
// [ZERO-TRUST PATCH]: LOCAL LLM INTEGRATION
// ==========================================

type ollamaOptions struct {
	Temperature float64 `json:"temperature"`
}

// OllamaRequest is the request payload sent to the local Ollama API.
type OllamaRequest struct {
	Model   string        `json:"model"`
	Prompt  string        `json:"prompt"`
	Stream  bool          `json:"stream"`
	Options ollamaOptions `json:"options"` // Enforce execution parameters (e.g., zero creativity)
}

// OllamaResponse is the response payload received from the local Ollama API.
type OllamaResponse struct {
	Response string `json:"response"`
}

// IsOllamaHealthy performs a health check (ping) against the Ollama endpoint.
func IsOllamaHealthy(endpoint string) bool {
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(endpoint + "/api/tags")
	if err != nil || resp.StatusCode != http.StatusOK {
		return false
	}
	defer resp.Body.Close()
	return true
}

// FilterWithLocalLLM sends rawText to the local Ollama LLM for DLP filtering
// and returns the redacted output. (Layer 5: AI Output Guardrail)
func FilterWithLocalLLM(endpoint, rawText string) (string, error) {
	url := endpoint + "/api/generate"

	// Strict prompt to prevent the LLM from summarizing or hallucinating.
	// Fixed the Go backtick escape issue by using "triple backticks" text.
	prompt := fmt.Sprintf(`You are a strict Data Redaction pipeline.
Your ONLY job is to output the EXACT same text you receive, but replace any Passwords, API Keys, or Private Keys with "[REDACTED]".
CRITICAL RULES:
1. If there are no secrets, you MUST return the original text EXACTLY as it is.
2. DO NOT summarize the text. DO NOT skip any lines.
3. DO NOT add explanations, greetings, or markdown blocks (like triple backticks).

TEXT TO PROCESS:
%s`, rawText)

	reqBody := OllamaRequest{
		Model:  "qwen2.5:3b",
		Prompt: prompt,
		Stream: false,
		Options: ollamaOptions{
			Temperature: 0.0, // Absolute zero creativity, mechanical execution only
		},
	}

	jsonData, _ := json.Marshal(reqBody)
	// 30s timeout to accommodate large text logs processing
	client := http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	var ollamaResp OllamaResponse
	if err := json.Unmarshal(bodyBytes, &ollamaResp); err != nil {
		return "", err
	}

	return ollamaResp.Response, nil
}
