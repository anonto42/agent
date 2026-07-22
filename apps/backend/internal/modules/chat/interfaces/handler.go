// Package interfaces exposes the chat module's HTTP surface:
//   - GET  /events  an SSE stream the client subscribes to (server -> client)
//   - POST /chat    send a user message           (client -> server)
//
// Together they give bidirectional realtime chat over plain HTTP.
package interfaces

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/levelaxis/charli/backend/internal/stream"
	"github.com/levelaxis/charli/backend/pkg/response"
)

// Handler serves the chat SSE stream and the message endpoint.
type Handler struct {
	hub *stream.Hub
}

// NewHandler constructs the chat handler.
func NewHandler(hub *stream.Hub) *Handler { return &Handler{hub: hub} }

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

type sendRequest struct {
	Session string `json:"session" binding:"required"`
	ID      string `json:"id" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// Send accepts a user message and pushes the reply to the session's stream.
func (h *Handler) Send(c *gin.Context) {
	var req sendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	// Phase 0: echo. Phase 1+: hand to the agent loop and stream its steps.
	reply := stream.Event{Type: "chat", ID: req.ID, Content: "You said: " + req.Content}
	if !h.hub.Publish(req.Session, reply) {
		response.Error(c, http.StatusNotFound, "no active stream for session", nil)
		return
	}
	response.OK(c, "accepted", nil)
}
