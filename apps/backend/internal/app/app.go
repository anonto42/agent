// Package app wires Charli's dependencies and HTTP router together.
// All construction happens here; leaf packages stay wiring-free.
package app

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	auditapp "github.com/levelaxis/charli/backend/internal/modules/audit/application"
	auditdomain "github.com/levelaxis/charli/backend/internal/modules/audit/domain"
	auditinfra "github.com/levelaxis/charli/backend/internal/modules/audit/infrastructure"
	chat "github.com/levelaxis/charli/backend/internal/modules/chat/interfaces"
	health "github.com/levelaxis/charli/backend/internal/modules/health/interfaces"
	"github.com/levelaxis/charli/backend/internal/safety"
	"github.com/levelaxis/charli/backend/internal/shared/config"
	"github.com/levelaxis/charli/backend/internal/shared/infrastructure/database"
	"github.com/levelaxis/charli/backend/internal/shared/infrastructure/llm"
	"github.com/levelaxis/charli/backend/internal/shared/middleware"
	"github.com/levelaxis/charli/backend/internal/stream"
	"github.com/levelaxis/charli/backend/internal/tools"
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
	registry := tools.Default()
	safetyEngine := safety.NewEngine(registry)
	auditService := auditapp.NewService(newAuditRepository(cfg, log), log)
	chat.RegisterRoutes(api, chat.NewHandler(hub, llmClient, log, registry, safetyEngine, auditService))

	return &App{Config: cfg, Logger: log, Engine: engine}
}

// Run starts the HTTP server and blocks.
func (a *App) Run() error {
	a.Logger.Info("charli backend listening", zap.String("port", a.Config.Port))
	return a.Engine.Run(":" + a.Config.Port)
}

// newAuditRepository connects to Postgres for persisted audit logging, if
// configured. Returns nil (audit.Service degrades to log-only) when no
// DatabaseURL is set or the database is unreachable — a missing/broken
// audit database must never stop the backend from starting.
func newAuditRepository(cfg *config.Config, log *zap.Logger) auditdomain.Repository {
	if cfg.DatabaseURL == "" {
		return nil
	}
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Warn("audit: database unavailable, persisting to logs only", zap.Error(err))
		return nil
	}
	repo, err := auditinfra.NewGormRepository(db)
	if err != nil {
		log.Warn("audit: migration failed, persisting to logs only", zap.Error(err))
		return nil
	}
	return repo
}
