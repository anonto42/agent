// Package llm is Charli's language-model client boundary. It defines the
// provider-agnostic Client interface the agent depends on, plus a factory
// that picks a concrete provider implementation (openai, google, deepseek)
// by name — the agent never knows which provider is behind it.
package llm

import (
	"context"

	"github.com/levelaxis/charli/backend/internal/shared/infrastructure/llm/deepseek"
	"github.com/levelaxis/charli/backend/internal/shared/infrastructure/llm/google"
	"github.com/levelaxis/charli/backend/internal/shared/infrastructure/llm/openai"
)

// Message is one turn sent to the model.
type Message struct {
	Role    string `json:"role"` // "system" | "user" | "assistant"
	Content string `json:"content"`
}

// Client is the minimal interface the agent depends on. Keeping it small is what
// makes the provider swappable (and easy to fake in tests).
type Client interface {
	Complete(ctx context.Context, messages []Message) (string, error)
}

// New builds a Client for the given provider ("openai", "google", or
// "deepseek"; defaults to "openai" for anything else, since that's also the
// shape Groq/Ollama/most self-hosted endpoints speak).
func New(provider, baseURL, apiKey, model string) Client {
	switch provider {
	case "google":
		return googleAdapter{google.New(baseURL, apiKey, model)}
	case "deepseek":
		return deepseekAdapter{deepseek.New(baseURL, apiKey, model)}
	default:
		return openaiAdapter{openai.New(baseURL, apiKey, model)}
	}
}

// Each provider package defines its own Message type (so it has zero
// dependency on this package and stays independently testable) — these
// adapters do the one-line conversion at the boundary.

type openaiAdapter struct{ c *openai.Client }

func (a openaiAdapter) Complete(ctx context.Context, messages []Message) (string, error) {
	out := make([]openai.Message, len(messages))
	for i, m := range messages {
		out[i] = openai.Message{Role: m.Role, Content: m.Content}
	}
	return a.c.Complete(ctx, out)
}

type deepseekAdapter struct{ c *deepseek.Client }

func (a deepseekAdapter) Complete(ctx context.Context, messages []Message) (string, error) {
	out := make([]deepseek.Message, len(messages))
	for i, m := range messages {
		out[i] = deepseek.Message{Role: m.Role, Content: m.Content}
	}
	return a.c.Complete(ctx, out)
}

type googleAdapter struct{ c *google.Client }

func (a googleAdapter) Complete(ctx context.Context, messages []Message) (string, error) {
	out := make([]google.Message, len(messages))
	for i, m := range messages {
		out[i] = google.Message{Role: m.Role, Content: m.Content}
	}
	return a.c.Complete(ctx, out)
}
