package interfaces

import "github.com/gin-gonic/gin"

// RegisterRoutes mounts the chat routes onto the given router group.
func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	r.GET("/events", h.Events)    // SSE stream (server -> client)
	r.POST("/chat", h.Send)       // send a message (client -> server)
	r.POST("/confirm", h.Confirm) // approve/reject a proposed action
}
