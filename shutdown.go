package service

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// StartWithGracefulShutdown starts the service with graceful shutdown handling
func (s *Service) StartWithGracefulShutdown() error {
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

// gracefulShutdown performs graceful shutdown of the service
func (s *Service) gracefulShutdown() error {
	s.Logger.Info("starting graceful shutdown")

	// Create a context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), s.Config.ShutdownTimeout)
	defer cancel()

	// Execute shutdown hooks
	for i, hook := range s.Config.ShutdownHooks {
		s.Logger.Info("executing shutdown hook", "index", i)
		if err := hook(); err != nil {
			s.Logger.Error("shutdown hook failed", "index", i, "error", err)
			// Continue with other hooks even if one fails
		}
	}

	// Shutdown servers
	var shutdownErrors []error

	// Shutdown main HTTP server
	if s.server != nil {
		s.Logger.Info("shutting down HTTP server")
		if err := s.server.Shutdown(ctx); err != nil {
			s.Logger.Error("HTTP server shutdown error", "error", err)
			shutdownErrors = append(shutdownErrors, err)
		}
	}

	// Shutdown metrics server
	if s.metricsServer != nil {
		s.Logger.Info("shutting down metrics server")
		if err := s.metricsServer.Shutdown(ctx); err != nil {
			s.Logger.Error("metrics server shutdown error", "error", err)
			shutdownErrors = append(shutdownErrors, err)
		}
	}

	if len(shutdownErrors) > 0 {
		s.Logger.Error("shutdown completed with errors", "error_count", len(shutdownErrors))
		return shutdownErrors[0] // Return first error
	}

	s.Logger.Info("graceful shutdown completed")
	return nil
}

// AddShutdownHook adds a function to be called during graceful shutdown
func (s *Service) AddShutdownHook(hook func() error) {
	s.Config.ShutdownHooks = append(s.Config.ShutdownHooks, hook)
}

// Stop stops the service gracefully
func (s *Service) Stop() error {
	return s.gracefulShutdown()
}
