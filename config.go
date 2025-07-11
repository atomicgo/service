package service

import (
	"log/slog"
	"os"
	"time"

	"github.com/caarlos0/env/v11"
)

// Config holds all configuration for the service
type Config struct {
	// HTTP Server configuration
	Addr         string        `env:"ADDR" envDefault:":8080"`
	ReadTimeout  time.Duration `env:"READ_TIMEOUT" envDefault:"10s"`
	WriteTimeout time.Duration `env:"WRITE_TIMEOUT" envDefault:"10s"`
	IdleTimeout  time.Duration `env:"IDLE_TIMEOUT" envDefault:"120s"`

	// Metrics server configuration
	MetricsAddr string `env:"METRICS_ADDR" envDefault:":9090"`
	MetricsPath string `env:"METRICS_PATH" envDefault:"/metrics"`

	// Graceful shutdown configuration
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"30s"`

	// Logger configuration
	Logger *slog.Logger `env:"-"`

	// Custom shutdown hooks
	ShutdownHooks []func() error `env:"-"`
}

// DefaultConfig creates a new config with default values
func DefaultConfig() *Config {
	return &Config{
		Addr:            ":8080",
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    10 * time.Second,
		IdleTimeout:     120 * time.Second,
		MetricsAddr:     ":9090",
		MetricsPath:     "/metrics",
		ShutdownTimeout: 30 * time.Second,
		Logger:          slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})),
		ShutdownHooks:   make([]func() error, 0),
	}
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() (*Config, error) {
	config := DefaultConfig()

	if err := env.Parse(config); err != nil {
		return nil, err
	}

	return config, nil
}

// AddShutdownHook adds a function to be called during graceful shutdown
func (c *Config) AddShutdownHook(hook func() error) {
	c.ShutdownHooks = append(c.ShutdownHooks, hook)
}
