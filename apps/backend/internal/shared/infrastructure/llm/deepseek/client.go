// Package deepseek calls DeepSeek's /chat/completions endpoint. DeepSeek's
// API is wire-compatible with OpenAI's today, so this delegates to the
// openai package rather than duplicating its request/response plumbing —
// it stays its own provider/package (own import path, own Message type,
// selectable via LLM_PROVIDER) so DeepSeek-specific behavior (e.g. its
// reasoning-model fields) can be layered in later without touching openai.
package deepseek

import (
	"context"

	"github.com/levelaxis/charli/backend/internal/shared/infrastructure/llm/openai"
)

// Message is one turn sent to the model.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Client calls DeepSeek's /chat/completions endpoint.
type Client struct {
	inner *openai.Client
}

// New builds a client for the given base URL, API key, and model.
func New(baseURL, apiKey, model string) *Client {
	return &Client{inner: openai.New(baseURL, apiKey, model)}
}

// Complete sends the conversation and returns the assistant's reply text.
func (c *Client) Complete(ctx context.Context, messages []Message) (string, error) {
	out := make([]openai.Message, len(messages))
	for i, m := range messages {
		out[i] = openai.Message{Role: m.Role, Content: m.Content}
	}
	return c.inner.Complete(ctx, out)
}
