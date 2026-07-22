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

	// LLM — OpenAI-compatible endpoint (defaults to Google Gemini's free tier).
	// Swap provider by changing these three env vars only.
	LLMBaseURL string
	LLMAPIKey  string
	LLMModel   string
}

// Load reads configuration from environment variables (and any exported .env).
func Load() (*Config, error) {
	v := viper.New()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.SetDefault("ENV", "development")
	v.SetDefault("PORT", "8080")
	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("LLM_BASE_URL", "https://generativelanguage.googleapis.com/v1beta/openai")
	v.SetDefault("LLM_MODEL", "gemini-2.0-flash")

	return &Config{
		Env:         v.GetString("ENV"),
		Port:        v.GetString("PORT"),
		LogLevel:    v.GetString("LOG_LEVEL"),
		DatabaseURL: v.GetString("DATABASE_URL"),
		RedisURL:    v.GetString("REDIS_URL"),
		LLMBaseURL:  v.GetString("LLM_BASE_URL"),
		LLMAPIKey:   v.GetString("LLM_API_KEY"),
		LLMModel:    v.GetString("LLM_MODEL"),
	}, nil
}
