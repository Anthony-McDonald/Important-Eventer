package main

import (
	"fmt"

	"github.com/caarlos0/env/v10"
)

// Config holds environment-driven configuration for the server and calendar.
type Config struct {
	Server struct {
		BaseURL string `env:"SERVER_URL" envDefault:"http://127.0.0.1"`
		Port    int    `env:"SERVER_PORT" envDefault:"8080"`
	}
	Calendar struct {
		CalendarURL    string `env:"CALENDAR_URL"`
		EventsToReturn int    `env:"CALENDAR_EVENTS_TO_RETURN" envDefault:"4"`
		RefreshMinutes int    `env:"CALENDAR_REFRESH_MINUTES" envDefault:"1"`
	}
	EmbeddingVectorURL string `env:"EMBEDDING_VECTOR_URL"`
}

// config is the global runtime configuration instance.
var config Config

// loadConfig parses environment variables into Config and validates required fields.
func loadConfig() (Config, error) {
	cfg := Config{}

	if err := env.Parse(&cfg); err != nil {
		return cfg, fmt.Errorf("env parse error: %w", err)
	}

	if cfg.Calendar.CalendarURL == "" {
		return cfg, fmt.Errorf("CALENDAR_URL must be set")
	}

	return cfg, nil
}
