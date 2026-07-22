// Package websocket hosts Charli's realtime gateway: one goroutine per session,
// streaming replies to the extension and receiving user/page messages.
//
// Phase 0: echo each chat message back as an assistant reply. Phase 1+ hands the
// message to the agent loop instead of echoing.
package websocket

import (
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Message is the Phase 0 chat envelope exchanged over the socket. `id`
// correlates a request with its reply so the client can resolve the right turn.
type Message struct {
	Type    string `json:"type"`
	ID      string `json:"id"`
	Content string `json:"content"`
}

// Handler upgrades HTTP connections to websockets and runs one session loop
// per connection.
type Handler struct {
	log *zap.Logger
}

// NewHandler constructs the websocket handler.
func NewHandler(log *zap.Logger) *Handler { return &Handler{log: log} }

// Serve accepts a websocket connection and processes messages until the client
// disconnects. Each connection is handled on its own goroutine (Gin/net-http).
func (h *Handler) Serve(c *gin.Context) {
	conn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
		// Dev: the extension connects from a chrome-extension:// origin. Replace
		// with an explicit OriginPatterns allowlist before production.
		InsecureSkipVerify: true,
	})
	if err != nil {
		h.log.Warn("ws accept failed", zap.Error(err))
		return
	}
	defer func() { _ = conn.CloseNow() }()

	ctx := c.Request.Context()
	h.log.Info("ws session opened")

	for {
		var msg Message
		if err := wsjson.Read(ctx, conn, &msg); err != nil {
			h.log.Info("ws session closed", zap.Error(err))
			return
		}

		// Phase 0: echo. Phase 1+: dispatch to the agent loop and stream steps.
		reply := Message{Type: "chat", ID: msg.ID, Content: "You said: " + msg.Content}
		if err := wsjson.Write(ctx, conn, reply); err != nil {
			h.log.Info("ws write failed", zap.Error(err))
			return
		}
	}
}
