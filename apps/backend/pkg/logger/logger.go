// Package logger builds the application's structured logger (zap).
package logger

import "go.uber.org/zap"

// New returns a zap logger tuned for the given environment.
func New(env string) (*zap.Logger, error) {
	if env == "production" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}
