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
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	logg, err := logger.New(cfg.Env)
	if err != nil {
		log.Fatalf("init logger: %v", err)
	}
	defer func() { _ = logg.Sync() }()

	if err := app.New(cfg, logg).Run(); err != nil {
		logg.Fatal("server exited", zap.Error(err))
	}
}
