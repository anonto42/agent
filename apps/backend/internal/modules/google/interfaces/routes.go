package interfaces

import "github.com/gin-gonic/gin"

// RegisterRoutes mounts the Google integration routes onto the given router
// group (expected to be the /integrations/google sub-group).
func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	r.POST("/connect", h.Connect)
	r.GET("/callback", h.Callback)
	r.GET("/status", h.Status)
	r.POST("/append", h.Append)
}
