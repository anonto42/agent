// Package app wires Charli's dependencies and HTTP router together.
// All construction happens here; leaf packages stay wiring-free.
package app

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	chat "github.com/levelaxis/charli/backend/internal/modules/chat/interfaces"
	health "github.com/levelaxis/charli/backend/internal/modules/health/interfaces"
	"github.com/levelaxis/charli/backend/internal/shared/config"
	"github.com/levelaxis/charli/backend/internal/shared/infrastructure/llm"
	"github.com/levelaxis/charli/backend/internal/shared/middleware"
	"github.com/levelaxis/charli/backend/internal/stream"
)

// App holds the top-level application dependencies.
type App struct {
	Config *config.Config
	Logger *zap.Logger
	Engine *gin.Engine
}

// New builds the application: router, middleware, and module routes.
func New(cfg *config.Config, log *zap.Logger) *App {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(gin.Recovery(), middleware.CORS())

	api := engine.Group("/api/v1")
	health.RegisterRoutes(api, health.NewHandler())

	// Realtime chat over SSE (stream down) + POST (send up), answered by the LLM.
	hub := stream.NewHub()
	llmClient := llm.New(cfg.LLMBaseURL, cfg.LLMAPIKey, cfg.LLMModel)
	chat.RegisterRoutes(api, chat.NewHandler(hub, llmClient, log))

	return &App{Config: cfg, Logger: log, Engine: engine}
}

// Run starts the HTTP server and blocks.
func (a *App) Run() error {
	a.Logger.Info("charli backend listening", zap.String("port", a.Config.Port))
	return a.Engine.Run(":" + a.Config.Port)
}
