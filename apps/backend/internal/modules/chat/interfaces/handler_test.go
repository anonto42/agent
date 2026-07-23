package interfaces

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/levelaxis/charli/backend/internal/shared/infrastructure/llm"
	"github.com/levelaxis/charli/backend/internal/stream"
)

// fakeLLM lets the test run offline (no real provider / API key).
type fakeLLM struct{ reply string }

func (f fakeLLM) Complete(_ context.Context, _ []llm.Message) (string, error) {
	return f.reply, nil
}

// capturingLLM records the messages it was asked to complete.
type capturingLLM struct {
	reply string
	got   chan []llm.Message
}

func (c *capturingLLM) Complete(_ context.Context, msgs []llm.Message) (string, error) {
	c.got <- msgs
	return c.reply, nil
}

func TestSendIncludesPageContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hub := stream.NewHub()
	cap := &capturingLLM{reply: "ok", got: make(chan []llm.Message, 1)}
	engine := gin.New()
	RegisterRoutes(engine.Group("/api/v1"), NewHandler(hub, cap, zap.NewNop()))

	srv := httptest.NewServer(engine)
	defer srv.Close()

	body := `{"session":"s1","id":"1","content":"summarize","page":"The secret word is platypus."}`
	resp, err := http.Post(srv.URL+"/api/v1/chat", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("post chat: %v", err)
	}
	_ = resp.Body.Close()

	select {
	case msgs := <-cap.got:
		var joined string
		for _, m := range msgs {
			joined += m.Content + "\n"
		}
		if !strings.Contains(joined, "platypus") {
			t.Fatalf("page context missing from prompt: %q", joined)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("llm was never called")
	}
}

func TestChatRoundTripSSE(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hub := stream.NewHub()
	engine := gin.New()
	handler := NewHandler(hub, fakeLLM{reply: "hello from charli"}, zap.NewNop())
	RegisterRoutes(engine.Group("/api/v1"), handler)

	srv := httptest.NewServer(engine)
	defer srv.Close()

	// Open the SSE stream.
	resp, err := http.Get(srv.URL + "/api/v1/events?session=s1")
	if err != nil {
		t.Fatalf("open sse: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	reader := bufio.NewReader(resp.Body)
	if _, err := reader.ReadString('\n'); err != nil { // consume ": connected"
		t.Fatalf("read connect line: %v", err)
	}

	// Send a message.
	body := `{"session":"s1","id":"1","content":"hi"}`
	pr, err := http.Post(srv.URL+"/api/v1/chat", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("post chat: %v", err)
	}
	_ = pr.Body.Close()

	// Read the stream until the model's reply arrives.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("read sse: %v", err)
		}
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		var e stream.Event
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if err := json.Unmarshal([]byte(payload), &e); err != nil {
			t.Fatalf("bad event json %q: %v", line, err)
		}
		if e.ID == "1" && e.Content == "hello from charli" {
			return // success: the (faked) model reply streamed back
		}
		t.Fatalf("unexpected event: %+v", e)
	}
	t.Fatal("timed out waiting for SSE event")
}
