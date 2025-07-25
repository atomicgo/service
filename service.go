package service

import (
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"syscall"

	"github.com/hellofresh/health-go/v5"
)

// Service represents the main service instance
type Service struct {
	Name          string
	Config        *Config
	Logger        *slog.Logger
	Metrics       *MetricsCollector
	HealthChecker *HealthChecker

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

	// Create health checker
	healthChecker, err := NewHealthChecker(name, config.Version)
	if err != nil {
		config.Logger.Error("failed to create health checker", "error", err)
		// Continue without health checker - it's not critical for basic operation
		healthChecker = nil
	}

	svc := &Service{
		Name:          name,
		Config:        config,
		Logger:        config.Logger,
		Metrics:       metrics,
		HealthChecker: healthChecker,
		mux:           http.NewServeMux(),
	}

	// Add default middleware (order matters: metrics should be first to capture all requests)
	svc.middlewares = []Middleware{
		MetricsMiddleware(metrics),
		LoggerMiddleware(config.Logger),
		RecoveryMiddleware(config.Logger),
		RequestLoggingMiddleware(config.Logger),
	}

	// Add health checker middleware if available
	if healthChecker != nil {
		svc.middlewares = append(svc.middlewares, HealthCheckerMiddleware(healthChecker))
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

// TestServer returns a httptest.Server with the service's mux
func (s *Service) TestServer() *httptest.Server {
	return httptest.NewServer(s.mux)
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
		if err := s.startMetricsServer(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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

		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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

// RegisterHealthCheck adds a health check to the service
func (s *Service) RegisterHealthCheck(config health.Config) error {
	if s.HealthChecker != nil {
		return s.HealthChecker.Register(config)
	}

	s.Logger.Warn("health checker not available, skipping health check registration", "name", config.Name)

	return nil
}

// RegisterCounter registers a new counter metric
func (s *Service) RegisterCounter(config MetricConfig) error {
	return s.Metrics.RegisterCounter(config)
}

// RegisterGauge registers a new gauge metric
func (s *Service) RegisterGauge(config MetricConfig) error {
	return s.Metrics.RegisterGauge(config)
}

// RegisterHistogram registers a new histogram metric
func (s *Service) RegisterHistogram(config MetricConfig) error {
	return s.Metrics.RegisterHistogram(config)
}

// RegisterSummary registers a new summary metric
func (s *Service) RegisterSummary(config MetricConfig) error {
	return s.Metrics.RegisterSummary(config)
}

// GetHealthChecker returns the health checker instance
func (s *Service) GetHealthChecker() *HealthChecker {
	return s.HealthChecker
}
