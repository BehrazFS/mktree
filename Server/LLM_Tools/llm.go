package llms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// LLM represents the language model configuration
type LLM struct {
	Model             string
	SystemPrompt      string
	Stream            bool
	MaxTokens         int
	Temperature       float64
	TopP              float64
	ReasoningEffort   string
	SupportsReasoning bool // new field to track if model supports ReasoningEffort
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest represents the request payload
type ChatRequest struct {
	Model       string    `json:"model"`
	Stream      bool      `json:"stream"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float64   `json:"temperature"`
	TopP        float64   `json:"top_p"`
	Messages    []Message `json:"messages"`

	// Only include if supported
	ReasoningEffort *string `json:"reasoning_effort,omitempty"`
}

// ChatResponse represents the API response
type ChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// NewLLM creates a new LLM instance
func NewLLM(model, systemPrompt string, supportsReasoning bool) *LLM {
	re := "low"
	return &LLM{
		Model:             model,
		SystemPrompt:      systemPrompt,
		Stream:            false,
		MaxTokens:         65536,
		Temperature:       0.5,
		TopP:              1.0,
		ReasoningEffort:   re,
		SupportsReasoning: supportsReasoning,
	}
}

// Call sends a query to the Cerebras API and returns the response
func (l *LLM) Call(query string) (string, error) {
	apiKey := os.Getenv("CEREBRAS_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("CEREBRAS_API_KEY environment variable not set")
	}

	messages := []Message{
		{Role: "system", Content: l.SystemPrompt},
		{Role: "user", Content: query},
	}

	reqPayload := ChatRequest{
		Model:       l.Model,
		Stream:      l.Stream,
		MaxTokens:   l.MaxTokens,
		Temperature: l.Temperature,
		TopP:        l.TopP,
		Messages:    messages,
	}

	// Only set ReasoningEffort if the model supports it
	if l.SupportsReasoning {
		reqPayload.ReasoningEffort = &l.ReasoningEffort
	}

	jsonData, err := json.Marshal(reqPayload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.cerebras.ai/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var chatResponse ChatResponse
	err = json.Unmarshal(body, &chatResponse)
	if err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	if len(chatResponse.Choices) > 0 {
		return chatResponse.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no response choices received")
}
