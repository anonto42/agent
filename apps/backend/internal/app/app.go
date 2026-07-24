// Package app wires Charli's dependencies and HTTP router together.
// All construction happens here; leaf packages stay wiring-free.
package app

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	googleoauth "golang.org/x/oauth2/google"
	"gorm.io/gorm"

	auditapp "github.com/levelaxis/charli/backend/internal/modules/audit/application"
	auditdomain "github.com/levelaxis/charli/backend/internal/modules/audit/domain"
	auditinfra "github.com/levelaxis/charli/backend/internal/modules/audit/infrastructure"
	chat "github.com/levelaxis/charli/backend/internal/modules/chat/interfaces"
	googleapp "github.com/levelaxis/charli/backend/internal/modules/google/application"
	googledomain "github.com/levelaxis/charli/backend/internal/modules/google/domain"
	googleinfra "github.com/levelaxis/charli/backend/internal/modules/google/infrastructure"
	google "github.com/levelaxis/charli/backend/internal/modules/google/interfaces"
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

	db := connectDatabase(cfg, log)

	// L4: connect to Google, act on a Sheet on the user's behalf.
	googleService := googleapp.NewService(newGoogleOAuthConfig(cfg), newGoogleRepository(db, log))
	google.RegisterRoutes(api.Group("/integrations/google"), google.NewHandler(googleService))

	// Realtime chat over SSE (stream down) + POST (send up), answered by the LLM.
	hub := stream.NewHub()
	llmClient := llm.New(cfg.LLMBaseURL, cfg.LLMAPIKey, cfg.LLMModel)
	registry := tools.Default()
	safetyEngine := safety.NewEngine(registry)
	auditService := auditapp.NewService(newAuditRepository(db, log), log)
	chat.RegisterRoutes(api, chat.NewHandler(hub, llmClient, log, registry, safetyEngine, auditService))

	return &App{Config: cfg, Logger: log, Engine: engine}
}

// Run starts the HTTP server and blocks.
func (a *App) Run() error {
	a.Logger.Info("charli backend listening", zap.String("port", a.Config.Port))
	return a.Engine.Run(":" + a.Config.Port)
}

// connectDatabase opens the one shared Postgres connection every DB-backed
// module's repository is built from. Returns nil when no DatabaseURL is set
// or the database is unreachable — a missing/broken database must never
// stop the backend from starting; callers degrade to their non-persisted
// behavior instead.
func connectDatabase(cfg *config.Config, log *zap.Logger) *gorm.DB {
	if cfg.DatabaseURL == "" {
		return nil
	}
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Warn("database unavailable, persisted features are disabled", zap.Error(err))
		return nil
	}
	return db
}

// newAuditRepository builds the audit log's repository, if a database is
// available. Returns nil (audit.Service degrades to log-only) otherwise.
func newAuditRepository(db *gorm.DB, log *zap.Logger) auditdomain.Repository {
	if db == nil {
		return nil
	}
	repo, err := auditinfra.NewGormRepository(db)
	if err != nil {
		log.Warn("audit: migration failed, persisting to logs only", zap.Error(err))
		return nil
	}
	return repo
}

// newGoogleRepository builds the Google integration's repository, if a
// database is available. Returns nil (google.Service reports itself
// unavailable) otherwise — there's nowhere to durably store OAuth tokens
// without one.
func newGoogleRepository(db *gorm.DB, log *zap.Logger) googledomain.Repository {
	if db == nil {
		return nil
	}
	repo, err := googleinfra.NewGormRepository(db)
	if err != nil {
		log.Warn("google: migration failed, integration unavailable", zap.Error(err))
		return nil
	}
	return repo
}

// newGoogleOAuthConfig builds the OAuth config for the Google integration,
// if credentials are set. Returns nil (google.Service reports itself
// unavailable) otherwise.
func newGoogleOAuthConfig(cfg *config.Config) *oauth2.Config {
	if cfg.GoogleClientID == "" || cfg.GoogleClientSecret == "" {
		return nil
	}
	return &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  fmt.Sprintf("http://localhost:%s/api/v1/integrations/google/callback", cfg.Port),
		Scopes:       []string{"https://www.googleapis.com/auth/spreadsheets"},
		Endpoint:     googleoauth.Endpoint,
	}
}
