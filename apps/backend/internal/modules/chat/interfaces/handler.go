// Package interfaces exposes the chat module's HTTP surface:
//   - GET  /events    an SSE stream the client subscribes to (server -> client)
//   - POST /chat      send a user message            (client -> server)
//   - POST /confirm   approve/reject a proposed action (client -> server)
//   - POST /observe   report an executed action's result, continuing the loop (L3)
//   - POST /interrupt stop an in-progress multi-step task (L3 kill switch)
//
// Handlers only validate input and delegate to application.Service; they hold
// no business logic themselves.
package interfaces

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/levelaxis/charli/backend/internal/modules/chat/application"
	"github.com/levelaxis/charli/backend/internal/safety"
	"github.com/levelaxis/charli/backend/internal/shared/infrastructure/llm"
	"github.com/levelaxis/charli/backend/internal/stream"
	"github.com/levelaxis/charli/backend/internal/tools"
	"github.com/levelaxis/charli/backend/pkg/response"
)

// Handler serves the chat SSE stream and the message/confirm endpoints.
type Handler struct {
	hub *stream.Hub
	svc *application.Service
}

// NewHandler constructs the chat handler.
func NewHandler(hub *stream.Hub, client llm.Client, log *zap.Logger, registry *tools.Registry, safetyEngine *safety.Engine) *Handler {
	return &Handler{hub: hub, svc: application.NewService(client, log, registry, safetyEngine)}
}

// Events opens a Server-Sent Events stream for a session. The client keeps it
// open; the server pushes assistant replies and proposed/executed actions down it.
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
	Page    string `json:"page"` // text of the page the user is viewing (L1)
}

// Send accepts a user message, acknowledges immediately, and produces the
// model's reply asynchronously — it arrives on the session's SSE stream, either
// as a normal answer or a proposed action awaiting confirmation.
func (h *Handler) Send(c *gin.Context) {
	var req sendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		event := h.svc.Reply(ctx, req.Session, req.ID, req.Content, req.Page)
		if event.Type != "" {
			h.hub.Publish(req.Session, event)
		}
	}()
	response.OK(c, "accepted", nil)
}

// confirmRequest is the POST /confirm body. Mirrors contracts.ConfirmRequest.
type confirmRequest struct {
	Session  string `json:"session" binding:"required"`
	ID       string `json:"id" binding:"required"`
	Approved bool   `json:"approved"`
}

// Confirm resolves a previously proposed action (approve or reject). The
// outcome (execute/cancelled/denied) arrives on the session's SSE stream.
func (h *Handler) Confirm(c *gin.Context) {
	var req confirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	event, found := h.svc.Confirm(req.Session, req.ID, req.Approved)
	if !found {
		response.Error(c, http.StatusNotFound, "no pending action for that id", nil)
		return
	}

	h.hub.Publish(req.Session, event)
	response.OK(c, "accepted", nil)
}

// observeRequest is the POST /observe body. Mirrors contracts.ObserveRequest.
type observeRequest struct {
	Session string `json:"session" binding:"required"`
	ID      string `json:"id" binding:"required"`
	Success bool   `json:"success"`
	Detail  string `json:"detail"`
}

// Observe reports whether a previously executed action succeeded and
// continues the agent loop (L3): the next turn (another proposed action, or
// a final answer) arrives asynchronously on the session's SSE stream.
func (h *Handler) Observe(c *gin.Context) {
	var req observeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	if !h.svc.HasTask(req.Session, req.ID) {
		response.Error(c, http.StatusNotFound, "no in-progress task for that id", nil)
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		event := h.svc.Observe(ctx, req.Session, req.ID, req.Success, req.Detail)
		if event.Type != "" {
			h.hub.Publish(req.Session, event)
		}
	}()
	response.OK(c, "accepted", nil)
}

// interruptRequest is the POST /interrupt body. Mirrors contracts.InterruptRequest.
type interruptRequest struct {
	Session string `json:"session" binding:"required"`
	ID      string `json:"id" binding:"required"`
}

// Interrupt is the user's kill switch: it stops an in-progress multi-step
// task. The "interrupted" outcome arrives on the session's SSE stream.
func (h *Handler) Interrupt(c *gin.Context) {
	var req interruptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	event, found := h.svc.Interrupt(req.Session, req.ID)
	if !found {
		response.Error(c, http.StatusNotFound, "no in-progress task for that id", nil)
		return
	}

	h.hub.Publish(req.Session, event)
	response.OK(c, "accepted", nil)
}
