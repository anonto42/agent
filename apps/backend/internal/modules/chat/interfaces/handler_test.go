package interfaces

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/levelaxis/charli/backend/internal/safety"
	"github.com/levelaxis/charli/backend/internal/shared/infrastructure/llm"
	"github.com/levelaxis/charli/backend/internal/stream"
	"github.com/levelaxis/charli/backend/internal/tools"
)

// newTestHandler wires a Handler with the default tool registry and a safety
// engine backed by it — the same wiring internal/app performs.
func newTestHandler(hub *stream.Hub, client llm.Client) *Handler {
	registry := tools.Default()
	return NewHandler(hub, client, zap.NewNop(), registry, safety.NewEngine(registry))
}

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

// sequenceLLM returns one reply per call, in order; calls past the end of
// the slice repeat the last reply. Safe for the one-call-at-a-time pattern
// these HTTP round-trip tests use (each request is awaited via SSE before
// the next is sent).
type sequenceLLM struct {
	mu      sync.Mutex
	replies []string
	calls   int
}

func (s *sequenceLLM) Complete(_ context.Context, _ []llm.Message) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	i := s.calls
	if i >= len(s.replies) {
		i = len(s.replies) - 1
	}
	s.calls++
	return s.replies[i], nil
}

func TestSendIncludesPageContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hub := stream.NewHub()
	cap := &capturingLLM{reply: "ok", got: make(chan []llm.Message, 1)}
	engine := gin.New()
	RegisterRoutes(engine.Group("/api/v1"), newTestHandler(hub, cap))

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
	handler := newTestHandler(hub, fakeLLM{reply: "hello from charli"})
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

	e := readNextEvent(t, reader)
	if e.ID != "1" || e.Content != "hello from charli" {
		t.Fatalf("unexpected event: %+v", e)
	}
}

// TestActionProposalAndConfirm exercises the full L2 loop over real HTTP + SSE:
// the model proposes an action -> client sees "action" -> confirms -> client
// sees "execute" with the same action.
func TestActionProposalAndConfirm(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hub := stream.NewHub()
	reply := `{"action":{"kind":"click","target":"Submit"},"message":"Click submit?"}`
	engine := gin.New()
	RegisterRoutes(engine.Group("/api/v1"), newTestHandler(hub, fakeLLM{reply: reply}))

	srv := httptest.NewServer(engine)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/events?session=s1")
	if err != nil {
		t.Fatalf("open sse: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	reader := bufio.NewReader(resp.Body)
	if _, err := reader.ReadString('\n'); err != nil {
		t.Fatalf("read connect line: %v", err)
	}

	body := `{"session":"s1","id":"1","content":"click submit"}`
	pr, err := http.Post(srv.URL+"/api/v1/chat", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("post chat: %v", err)
	}
	_ = pr.Body.Close()

	proposed := readNextEvent(t, reader)
	if proposed.Type != "action" || proposed.Action == nil || proposed.Action.Kind != "click" {
		t.Fatalf("expected a proposed click action, got %+v", proposed)
	}

	confirmBody := `{"session":"s1","id":"1","approved":true}`
	cr, err := http.Post(srv.URL+"/api/v1/confirm", "application/json", strings.NewReader(confirmBody))
	if err != nil {
		t.Fatalf("post confirm: %v", err)
	}
	if cr.StatusCode != http.StatusOK {
		t.Fatalf("confirm status: %d", cr.StatusCode)
	}
	_ = cr.Body.Close()

	executed := readNextEvent(t, reader)
	if executed.Type != "execute" || executed.Action == nil || executed.Action.Target != "Submit" {
		t.Fatalf("expected an execute event for Submit, got %+v", executed)
	}
}

// TestConfirmUnknownIDReturns404 checks that confirming a never-proposed (or
// already-resolved) id is rejected rather than silently accepted.
func TestConfirmUnknownIDReturns404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hub := stream.NewHub()
	engine := gin.New()
	RegisterRoutes(engine.Group("/api/v1"), newTestHandler(hub, fakeLLM{reply: "n/a"}))

	srv := httptest.NewServer(engine)
	defer srv.Close()

	body := `{"session":"s1","id":"does-not-exist","approved":true}`
	resp, err := http.Post(srv.URL+"/api/v1/confirm", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("post confirm: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

// TestObserveContinuesLoopThenFinishes exercises the full L3 loop over real
// HTTP + SSE: propose -> confirm -> execute -> observe -> the model's next
// turn is a plain final answer.
func TestObserveContinuesLoopThenFinishes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hub := stream.NewHub()
	llmClient := &sequenceLLM{replies: []string{
		`{"action":{"kind":"click","target":"Submit"},"message":"Click submit?"}`,
		"All done!",
	}}
	engine := gin.New()
	RegisterRoutes(engine.Group("/api/v1"), newTestHandler(hub, llmClient))

	srv := httptest.NewServer(engine)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/events?session=s1")
	if err != nil {
		t.Fatalf("open sse: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	reader := bufio.NewReader(resp.Body)
	if _, err := reader.ReadString('\n'); err != nil {
		t.Fatalf("read connect line: %v", err)
	}

	body := `{"session":"s1","id":"1","content":"click submit"}`
	pr, err := http.Post(srv.URL+"/api/v1/chat", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("post chat: %v", err)
	}
	_ = pr.Body.Close()

	proposed := readNextEvent(t, reader)
	if proposed.Type != "action" {
		t.Fatalf("expected a proposed action, got %+v", proposed)
	}

	confirmBody := `{"session":"s1","id":"1","approved":true}`
	cr, err := http.Post(srv.URL+"/api/v1/confirm", "application/json", strings.NewReader(confirmBody))
	if err != nil {
		t.Fatalf("post confirm: %v", err)
	}
	_ = cr.Body.Close()

	executed := readNextEvent(t, reader)
	if executed.Type != "execute" {
		t.Fatalf("expected execute, got %+v", executed)
	}

	observeBody := `{"session":"s1","id":"1","success":true}`
	or, err := http.Post(srv.URL+"/api/v1/observe", "application/json", strings.NewReader(observeBody))
	if err != nil {
		t.Fatalf("post observe: %v", err)
	}
	if or.StatusCode != http.StatusOK {
		t.Fatalf("observe status: %d", or.StatusCode)
	}
	_ = or.Body.Close()

	final := readNextEvent(t, reader)
	if final.Type != "chat" || final.Content != "All done!" {
		t.Fatalf("expected the loop to continue and finish, got %+v", final)
	}
}

// TestObserveUnknownIDReturns404 checks that observing a never-executed (or
// already-finished) task is rejected rather than silently accepted.
func TestObserveUnknownIDReturns404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hub := stream.NewHub()
	engine := gin.New()
	RegisterRoutes(engine.Group("/api/v1"), newTestHandler(hub, fakeLLM{reply: "n/a"}))

	srv := httptest.NewServer(engine)
	defer srv.Close()

	body := `{"session":"s1","id":"does-not-exist","success":true}`
	resp, err := http.Post(srv.URL+"/api/v1/observe", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("post observe: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

// TestInterruptStopsInProgressTask covers the kill switch over real HTTP +
// SSE: interrupting a task with a pending action publishes "interrupted",
// and the action can no longer be confirmed afterward.
func TestInterruptStopsInProgressTask(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hub := stream.NewHub()
	reply := `{"action":{"kind":"click","target":"Submit"},"message":"Click submit?"}`
	engine := gin.New()
	RegisterRoutes(engine.Group("/api/v1"), newTestHandler(hub, fakeLLM{reply: reply}))

	srv := httptest.NewServer(engine)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/events?session=s1")
	if err != nil {
		t.Fatalf("open sse: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	reader := bufio.NewReader(resp.Body)
	if _, err := reader.ReadString('\n'); err != nil {
		t.Fatalf("read connect line: %v", err)
	}

	body := `{"session":"s1","id":"1","content":"click submit"}`
	pr, err := http.Post(srv.URL+"/api/v1/chat", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("post chat: %v", err)
	}
	_ = pr.Body.Close()

	proposed := readNextEvent(t, reader)
	if proposed.Type != "action" {
		t.Fatalf("expected a proposed action, got %+v", proposed)
	}

	interruptBody := `{"session":"s1","id":"1"}`
	ir, err := http.Post(srv.URL+"/api/v1/interrupt", "application/json", strings.NewReader(interruptBody))
	if err != nil {
		t.Fatalf("post interrupt: %v", err)
	}
	if ir.StatusCode != http.StatusOK {
		t.Fatalf("interrupt status: %d", ir.StatusCode)
	}
	_ = ir.Body.Close()

	stopped := readNextEvent(t, reader)
	if stopped.Type != "interrupted" || stopped.Content != "Stopped." {
		t.Fatalf("expected an interrupted event, got %+v", stopped)
	}

	confirmBody := `{"session":"s1","id":"1","approved":true}`
	cr, err := http.Post(srv.URL+"/api/v1/confirm", "application/json", strings.NewReader(confirmBody))
	if err != nil {
		t.Fatalf("post confirm: %v", err)
	}
	defer func() { _ = cr.Body.Close() }()
	if cr.StatusCode != http.StatusNotFound {
		t.Fatalf("expected confirming an interrupted action to 404, got %d", cr.StatusCode)
	}
}

// TestInterruptUnknownIDReturns404 checks that interrupting a session with no
// in-progress task is rejected rather than silently accepted.
func TestInterruptUnknownIDReturns404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hub := stream.NewHub()
	engine := gin.New()
	RegisterRoutes(engine.Group("/api/v1"), newTestHandler(hub, fakeLLM{reply: "n/a"}))

	srv := httptest.NewServer(engine)
	defer srv.Close()

	body := `{"session":"s1","id":"does-not-exist"}`
	resp, err := http.Post(srv.URL+"/api/v1/interrupt", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("post interrupt: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

// readNextEvent reads SSE lines until it finds the next `data:` payload and
// decodes it as a stream.Event.
func readNextEvent(t *testing.T, reader *bufio.Reader) stream.Event {
	t.Helper()
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
		return e
	}
	t.Fatal("timed out waiting for SSE event")
	return stream.Event{}
}
