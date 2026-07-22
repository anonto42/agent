// Package interfaces exposes the chat module's HTTP surface:
//   - GET  /events  an SSE stream the client subscribes to (server -> client)
//   - POST /chat    send a user message           (client -> server)
//
// Together they give bidirectional realtime chat over plain HTTP.
package interfaces

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/levelaxis/charli/backend/internal/shared/infrastructure/llm"
	"github.com/levelaxis/charli/backend/internal/stream"
	"github.com/levelaxis/charli/backend/pkg/response"
)

const systemPrompt = "You are Charli, a helpful, concise browser assistant."

// Handler serves the chat SSE stream and the message endpoint.
type Handler struct {
	hub *stream.Hub
	llm llm.Client
	log *zap.Logger
}

// NewHandler constructs the chat handler.
func NewHandler(hub *stream.Hub, client llm.Client, log *zap.Logger) *Handler {
	return &Handler{hub: hub, llm: client, log: log}
}

// Events opens a Server-Sent Events stream for a session. The client keeps it
// open; the server pushes assistant replies (and, later, agent steps) down it.
func (h *Handler) Events(c *gin.Context) {
	session := c.Query("session")
	if session == "" {
		response.Error(c, http.StatusBadRequest, "missing session", nil)
		return
	}

	ch := h.hub.Subscribe(session)
	defer h.hub.Unsubscribe(session)

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// Flush headers immediately so the client's subscription is live before it
	// POSTs its first message.
	if _, err := c.Writer.Write([]byte(": connected\n\n")); err != nil {
		return
	}
	c.Writer.Flush()

	c.Stream(func(_ io.Writer) bool {
		select {
		case e, ok := <-ch:
			if !ok {
				return false
			}
			c.SSEvent("message", e)
			return true
		case <-c.Request.Context().Done():
			return false
		}
	})
}

// sendRequest is the POST /chat body. Mirrors contracts.ChatRequest.
type sendRequest struct {
	Session string `json:"session" binding:"required"`
	ID      string `json:"id" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// Send accepts a user message, acknowledges immediately, and produces the
// model's reply asynchronously — it arrives on the session's SSE stream.
func (h *Handler) Send(c *gin.Context) {
	var req sendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	go h.reply(req.Session, req.ID, req.Content)
	response.OK(c, "accepted", nil)
}

// reply calls the model and pushes the answer (or an error event) to the
// session's stream. Runs on its own goroutine, off the request context.
func (h *Handler) reply(session, id, content string) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	answer, err := h.llm.Complete(ctx, []llm.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: content},
	})
	if err != nil {
		h.log.Warn("llm complete failed", zap.Error(err))
		h.hub.Publish(session, stream.Event{Type: "error", ID: id, Content: "Charli could not answer right now."})
		return
	}
	h.hub.Publish(session, stream.Event{Type: "chat", ID: id, Content: answer})
}
