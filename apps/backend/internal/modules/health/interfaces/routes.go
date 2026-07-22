package interfaces

import "github.com/gin-gonic/gin"

// RegisterRoutes mounts the health routes onto the given router group.
func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	r.GET("/health", h.Health)
}
