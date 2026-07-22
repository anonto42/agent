// Package config loads runtime configuration from the environment.
package config

import (
	"strings"

	"github.com/spf13/viper"
)

// Config holds all runtime configuration for the backend.
type Config struct {
	Env      string
	Port     string
	LogLevel string

	DatabaseURL string
	RedisURL    string

	AnthropicAPIKey string
}

// Load reads configuration from environment variables (and any exported .env).
func Load() (*Config, error) {
	v := viper.New()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.SetDefault("ENV", "development")
	v.SetDefault("PORT", "8080")
	v.SetDefault("LOG_LEVEL", "info")

	return &Config{
		Env:             v.GetString("ENV"),
		Port:            v.GetString("PORT"),
		LogLevel:        v.GetString("LOG_LEVEL"),
		DatabaseURL:     v.GetString("DATABASE_URL"),
		RedisURL:        v.GetString("REDIS_URL"),
		AnthropicAPIKey: v.GetString("ANTHROPIC_API_KEY"),
	}, nil
}
