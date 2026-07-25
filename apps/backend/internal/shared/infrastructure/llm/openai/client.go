// Package openai calls any OpenAI-compatible /chat/completions endpoint.
// This is also what Groq, Ollama, and most self-hosted OpenAI-shaped
// endpoints speak, so it doubles as the default/fallback provider.
package openai

import (
	"context"
	"fmt"
	"time"

	"github.com/levelaxis/charli/backend/internal/shared/infrastructure/llm/internal/httpjson"
)

// Message is one turn sent to the model.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Client calls an OpenAI-compatible /chat/completions endpoint.
type Client struct {
	baseURL string
	apiKey  string
	model   string
	http    *httpjson.Client
}

// New builds a client for the given base URL, API key, and model.
func New(baseURL, apiKey, model string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		model:   model,
		http:    httpjson.New(60 * time.Second),
	}
}

type chatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type chatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Complete sends the conversation and returns the assistant's reply text.
func (c *Client) Complete(ctx context.Context, messages []Message) (string, error) {
	var parsed chatResponse
	headers := map[string]string{"Authorization": "Bearer " + c.apiKey}
	_, err := c.http.PostJSON(ctx, c.baseURL+"/chat/completions", headers, chatRequest{Model: c.model, Messages: messages}, &parsed)
	if parsed.Error != nil {
		return "", fmt.Errorf("llm error: %s", parsed.Error.Message)
	}
	if err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("llm returned no choices")
	}
	return parsed.Choices[0].Message.Content, nil
}
