package interfaces

import (
	"bufio"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/levelaxis/charli/backend/internal/stream"
)

func TestChatRoundTripSSE(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hub := stream.NewHub()
	engine := gin.New()
	RegisterRoutes(engine.Group("/api/v1"), NewHandler(hub))

	srv := httptest.NewServer(engine)
	defer srv.Close()

	// Open the SSE stream.
	resp, err := http.Get(srv.URL + "/api/v1/events?session=s1")
	if err != nil {
		t.Fatalf("open sse: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	reader := bufio.NewReader(resp.Body)
	// Consume the ": connected" comment; the subscription is now live.
	if _, err := reader.ReadString('\n'); err != nil {
		t.Fatalf("read connect line: %v", err)
	}

	// Send a message.
	body := `{"session":"s1","id":"1","content":"hi"}`
	pr, err := http.Post(srv.URL+"/api/v1/chat", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("post chat: %v", err)
	}
	_ = pr.Body.Close()

	// Read the stream until the reply arrives.
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
		if e.ID == "1" && e.Content == "You said: hi" {
			return // success
		}
		t.Fatalf("unexpected event: %+v", e)
	}
	t.Fatal("timed out waiting for SSE event")
}
