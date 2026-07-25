// Package httpjson is the shared "POST JSON, check status, decode JSON"
// plumbing every LLM provider client needs. Providers differ in request/
// response shape and auth placement, not in this mechanical part.
package httpjson

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client does POST-JSON-get-JSON calls with a shared timeout.
type Client struct {
	HTTP *http.Client
}

// New builds a Client with the given timeout.
func New(timeout time.Duration) *Client {
	return &Client{HTTP: &http.Client{Timeout: timeout}}
}

// PostJSON marshals reqBody, POSTs it to url with headers applied, and
// unmarshals the response body into respBody. Returns the raw response body
// alongside any error so callers can surface API-specific error payloads.
func (c *Client) PostJSON(ctx context.Context, url string, headers map[string]string, reqBody, respBody any) ([]byte, error) {
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call llm: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	// Decode even on a non-2xx status: most providers put a usable error
	// message in the same JSON shape as a success response (e.g. an
	// "error.message" field), so callers can inspect respBody either way.
	_ = json.Unmarshal(raw, respBody)

	if resp.StatusCode != http.StatusOK {
		return raw, fmt.Errorf("llm status %d: %s", resp.StatusCode, string(raw))
	}
	return raw, nil
}
