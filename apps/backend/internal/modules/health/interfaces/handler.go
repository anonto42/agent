// Package interfaces exposes the health module's HTTP handlers.
//
// This module is the reference for the per-module layout every domain follows:
//
//	modules/<domain>/
//	├── domain/          entity.go, repository.go (interface)
//	├── application/     dto.go, service.go (business logic)
//	├── infrastructure/  repository.go (GORM / external impls)
//	└── interfaces/      handler.go, routes.go (Gin)
package interfaces

import (
	"github.com/gin-gonic/gin"

	"github.com/levelaxis/charli/backend/pkg/response"
)

// Handler serves health / liveness checks.
type Handler struct{}

// NewHandler constructs the health handler.
func NewHandler() *Handler { return &Handler{} }

// Health reports service liveness.
func (h *Handler) Health(c *gin.Context) {
	response.OK(c, "ok", gin.H{"service": "charli"})
}
