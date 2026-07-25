// Package google calls Gemini's native generateContent API. Unlike openai
// and deepseek, Gemini's wire format isn't OpenAI-compatible: the API key
// goes in a query param (not an Authorization header), there's no "system"
// role (system instructions are a separate top-level field), and replies
// come back as "candidates" of "parts", not "choices" of "message".
package google

import (
	"context"
	"fmt"
	"time"

	"github.com/levelaxis/charli/backend/internal/shared/infrastructure/llm/internal/httpjson"
)

// Message is one turn sent to the model.
type Message struct {
	Role    string `json:"role"` // "system" | "user" | "assistant"
	Content string `json:"content"`
}

// Client calls Gemini's v1beta generateContent endpoint.
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

type contentPart struct {
	Text string `json:"text"`
}

type content struct {
	Role  string        `json:"role"` // "user" | "model"
	Parts []contentPart `json:"parts"`
}

type systemInstruction struct {
	Parts []contentPart `json:"parts"`
}

type generateRequest struct {
	Contents          []content          `json:"contents"`
	SystemInstruction *systemInstruction `json:"systemInstruction,omitempty"`
}

type generateResponse struct {
	Candidates []struct {
		Content content `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Complete sends the conversation and returns the assistant's reply text.
// Gemini has no "system" role, so leading system messages are collected
// into systemInstruction and the rest are mapped to user/model turns.
func (c *Client) Complete(ctx context.Context, messages []Message) (string, error) {
	reqBody := toGenerateRequest(messages)

	var parsed generateResponse
	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", c.baseURL, c.model, c.apiKey)
	_, err := c.http.PostJSON(ctx, url, nil, reqBody, &parsed)
	if parsed.Error != nil {
		return "", fmt.Errorf("llm error: %s", parsed.Error.Message)
	}
	if err != nil {
		return "", err
	}
	if len(parsed.Candidates) == 0 || len(parsed.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("llm returned no candidates")
	}
	return parsed.Candidates[0].Content.Parts[0].Text, nil
}

// toGenerateRequest maps role-tagged messages onto Gemini's shape: system
// messages collapse into a single systemInstruction, everything else
// becomes a user/model turn.
func toGenerateRequest(messages []Message) generateRequest {
	var systemText string
	contents := make([]content, 0, len(messages))
	for _, m := range messages {
		switch m.Role {
		case "system":
			if systemText != "" {
				systemText += "\n\n"
			}
			systemText += m.Content
		case "assistant":
			contents = append(contents, content{Role: "model", Parts: []contentPart{{Text: m.Content}}})
		default:
			contents = append(contents, content{Role: "user", Parts: []contentPart{{Text: m.Content}}})
		}
	}

	reqBody := generateRequest{Contents: contents}
	if systemText != "" {
		reqBody.SystemInstruction = &systemInstruction{Parts: []contentPart{{Text: systemText}}}
	}
	return reqBody
}
