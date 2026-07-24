// Package interfaces exposes the Google integration's HTTP surface, under
// /api/v1/integrations/google:
//   - POST /connect  start an OAuth connection for a device (client -> server)
//   - GET  /callback Google's OAuth redirect target (server-to-server, via browser)
//   - GET  /status   whether a device is connected                 (client -> server)
//   - POST /append   perform a confirmed sheets_append action      (client -> server)
//
// /append is deliberately NOT part of the SSE chat protocol — it's what the
// extension's performAction calls to "execute" a sheets_append the same way
// it executes fill/click on the DOM (see apps/extension/shared/lib/performAction.ts).
package interfaces

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/levelaxis/charli/backend/internal/modules/google/application"
	"github.com/levelaxis/charli/backend/pkg/response"
)

// Handler serves the Google integration's connect/callback/status/append endpoints.
type Handler struct {
	svc *application.Service
}

// NewHandler constructs the Google integration handler.
func NewHandler(svc *application.Service) *Handler {
	return &Handler{svc: svc}
}

// connectRequest is the POST /connect body. Mirrors contracts.GoogleConnectRequest.
type connectRequest struct {
	DeviceID string `json:"deviceId" binding:"required"`
}

// Connect starts an OAuth attempt and returns the consent URL to open.
func (h *Handler) Connect(c *gin.Context) {
	var req connectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	authURL, err := h.svc.BeginConnect(req.DeviceID)
	if err != nil {
		response.Error(c, http.StatusServiceUnavailable, err.Error(), nil)
		return
	}
	response.OK(c, "ok", gin.H{"authUrl": authURL})
}

// Callback is Google's OAuth redirect target. It completes the flow and
// shows the user a small confirmation page in the tab Google redirected.
func (h *Handler) Callback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if err := h.svc.CompleteConnect(c.Request.Context(), code, state); err != nil {
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(
			`<p>Couldn't connect Google Sheets: `+err.Error()+`. You can close this tab and try again.</p>`))
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(
		`<p>Connected! You can close this tab.</p>`))
}

// Status reports whether a device has a completed Google connection.
func (h *Handler) Status(c *gin.Context) {
	deviceID := c.Query("deviceId")
	if deviceID == "" {
		response.Error(c, http.StatusBadRequest, "missing deviceId", nil)
		return
	}
	response.OK(c, "ok", gin.H{"connected": h.svc.IsConnected(deviceID)})
}

// appendRequest is the POST /append body. Mirrors contracts.GoogleAppendRequest.
// DeviceID is deliberately NOT required: an empty/unknown device just means
// "not connected," which AppendRow already reports as a clear failure detail
// rather than a validation error — a transient client-side identity glitch
// shouldn't turn into an opaque 400.
type appendRequest struct {
	DeviceID      string   `json:"deviceId"`
	SpreadsheetID string   `json:"spreadsheetId" binding:"required"`
	Values        []string `json:"values" binding:"required"`
}

// Append performs a confirmed sheets_append action.
func (h *Handler) Append(c *gin.Context) {
	var req appendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	if err := h.svc.AppendRow(c.Request.Context(), req.DeviceID, req.SpreadsheetID, req.Values); err != nil {
		response.OK(c, "ok", gin.H{"success": false, "detail": err.Error()})
		return
	}
	response.OK(c, "ok", gin.H{"success": true})
}
