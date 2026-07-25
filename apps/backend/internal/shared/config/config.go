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

	// LLM — defaults to Google Gemini's free tier over its OpenAI-compatible
	// endpoint. Swap provider by changing these four env vars: LLM_PROVIDER
	// picks the implementation ("openai", "google", or "deepseek"; see
	// internal/shared/infrastructure/llm), the rest configure it.
	LLMProvider string
	LLMBaseURL  string
	LLMAPIKey   string
	LLMModel    string

	// Google OAuth (L4 — Sheets integration). Empty means the integration is
	// unavailable; it never blocks startup.
	GoogleClientID     string
	GoogleClientSecret string
}

// Load reads configuration from environment variables and .env file.
func Load() (*Config, error) {
	// Create a new Viper instance and bind it to environment variables.
	v := viper.New()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Attempt to read the .env file; silently skip if it doesn't exist.
	v.SetConfigFile(".env")
	v.SetConfigType("env")
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	// Set defaults for fields that have sensible fallback values.
	v.SetDefault("ENV", "development")
	v.SetDefault("PORT", "8080")
	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("LLM_PROVIDER", "openai")
	v.SetDefault("LLM_BASE_URL", "https://generativelanguage.googleapis.com/v1beta/openai")
	v.SetDefault("LLM_MODEL", "gemini-2.0-flash")

	// Extract every config value from Viper (env > .env > default) and return.
	return &Config{
		Env:         v.GetString("ENV"),
		Port:        v.GetString("PORT"),
		LogLevel:    v.GetString("LOG_LEVEL"),
		DatabaseURL: v.GetString("DATABASE_URL"),
		RedisURL:    v.GetString("REDIS_URL"),
		LLMProvider: v.GetString("LLM_PROVIDER"),
		LLMBaseURL:  v.GetString("LLM_BASE_URL"),
		LLMAPIKey:   v.GetString("LLM_API_KEY"),
		LLMModel:    v.GetString("LLM_MODEL"),

		GoogleClientID:     v.GetString("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: v.GetString("GOOGLE_CLIENT_SECRET"),
	}, nil
}
