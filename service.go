package service

import (
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// Service represents the main service instance
type Service struct {
	Name    string
	Config  *Config
	Logger  *slog.Logger
	Metrics *MetricsCollector

	server        *http.Server
	metricsServer *http.Server
	mux           *http.ServeMux
	middlewares   []Middleware
}

// New creates a new service instance
func New(name string, config *Config) *Service {
	if config == nil {
		config = DefaultConfig()
	}

	// Create metrics collector
	metrics := NewMetricsCollector(name)

	svc := &Service{
		Name:    name,
		Config:  config,
		Logger:  config.Logger,
		Metrics: metrics,
		mux:     http.NewServeMux(),
	}

	// Add default middleware (order matters: metrics should be first to capture all requests)
	svc.middlewares = []Middleware{
		MetricsMiddleware(metrics),
		LoggerMiddleware(config.Logger),
		RecoveryMiddleware(config.Logger),
		RequestLoggingMiddleware(config.Logger),
	}

	return svc
}

// HandleFunc registers a handler function for the given pattern
func (s *Service) HandleFunc(pattern string, handler http.HandlerFunc) {
	// Apply middleware to the handler
	wrappedHandler := applyMiddleware(handler, s.middlewares...)
	s.mux.Handle(pattern, wrappedHandler)
}

// Handle registers a handler for the given pattern
func (s *Service) Handle(pattern string, handler http.Handler) {
	// Apply middleware to the handler
	wrappedHandler := applyMiddleware(handler, s.middlewares...)
	s.mux.Handle(pattern, wrappedHandler)
}

// Use adds middleware to the service
func (s *Service) Use(middleware Middleware) {
	s.middlewares = append(s.middlewares, middleware)
}

// Start starts the service with graceful shutdown handling
func (s *Service) Start() error {
	// Create a channel to receive OS signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start the servers in goroutines
	serverErrors := make(chan error, 2)

	// Start metrics server
	go func() {
		if err := s.startMetricsServer(); err != nil && err != http.ErrServerClosed {
			s.Logger.Error("metrics server error", "error", err)
			serverErrors <- err
		}
	}()

	// Start main HTTP server
	go func() {
		s.server = &http.Server{
			Addr:         s.Config.Addr,
			Handler:      s.mux,
			ReadTimeout:  s.Config.ReadTimeout,
			WriteTimeout: s.Config.WriteTimeout,
			IdleTimeout:  s.Config.IdleTimeout,
		}

		s.Logger.Info("starting service", "name", s.Name, "addr", s.Config.Addr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.Logger.Error("server error", "error", err)
			serverErrors <- err
		}
	}()

	// Wait for either a signal or a server error
	select {
	case <-quit:
		s.Logger.Info("received shutdown signal")
	case err := <-serverErrors:
		s.Logger.Error("server error, shutting down", "error", err)
		return err
	}

	// Perform graceful shutdown
	return s.gracefulShutdown()
}
