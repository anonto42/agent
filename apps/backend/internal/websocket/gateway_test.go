package websocket

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func TestHandlerEchoesChat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/ws", NewHandler(zap.NewNop()).Serve)

	srv := httptest.NewServer(engine)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	url := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	conn, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer func() { _ = conn.CloseNow() }()

	if err := wsjson.Write(ctx, conn, Message{Type: "chat", ID: "1", Content: "hi"}); err != nil {
		t.Fatalf("write: %v", err)
	}

	var reply Message
	if err := wsjson.Read(ctx, conn, &reply); err != nil {
		t.Fatalf("read: %v", err)
	}

	if reply.ID != "1" || reply.Content != "You said: hi" {
		t.Fatalf("unexpected reply: %+v", reply)
	}
}
