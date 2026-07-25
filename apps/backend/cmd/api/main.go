// Command api is the entrypoint for the Charli agent backend.
package main

import (
	"log"

	"go.uber.org/zap"

	"github.com/levelaxis/charli/backend/internal/app"
	"github.com/levelaxis/charli/backend/internal/shared/config"
	"github.com/levelaxis/charli/backend/pkg/logger"
)

func main() {
	// Load runtime configuration from environment variables and .env file.
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// Initialize the structured logger (zap) — production or development mode based on ENV.
	logg, err := logger.New(cfg.Env)
	if err != nil {
		log.Fatalf("init logger: %v", err)
	}
	// Flush any buffered log entries before the process exits.
	defer func() { _ = logg.Sync() }()

	// Wire all dependencies together and start the HTTP server. Blocks until the server exits.
	if err := app.New(cfg, logg).Run(); err != nil {
		logg.Fatal("server exited", zap.Error(err))
	}
}
