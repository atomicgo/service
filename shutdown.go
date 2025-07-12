package service

import (
	"context"
)

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
