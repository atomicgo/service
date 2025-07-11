package service

import (
	"context"
	"log/slog"
	"net/http"
)

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

const (
	// LoggerKey is the context key for the logger
	LoggerKey ContextKey = "logger"
	// MetricsKey is the context key for metrics
	MetricsKey ContextKey = "metrics"
)

// Middleware represents a middleware function
type Middleware func(http.Handler) http.Handler

// LoggerMiddleware injects the logger into the request context
func LoggerMiddleware(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a new context with the logger
			ctx := context.WithValue(r.Context(), LoggerKey, logger)

			// Create a new request with the updated context
			r = r.WithContext(ctx)

			// Call the next handler
			next.ServeHTTP(w, r)
		})
	}
}

// GetLogger retrieves the logger from the request context
func GetLogger(r *http.Request) *slog.Logger {
	logger, ok := r.Context().Value(LoggerKey).(*slog.Logger)
	if !ok {
		// Return a default logger if none is found
		return slog.Default()
	}
	return logger
}

// RecoveryMiddleware recovers from panics and logs them
func RecoveryMiddleware(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("panic recovered", "error", err, "path", r.URL.Path, "method", r.Method)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// RequestLoggingMiddleware logs incoming requests
func RequestLoggingMiddleware(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info("incoming request",
				"method", r.Method,
				"path", r.URL.Path,
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent())

			next.ServeHTTP(w, r)
		})
	}
}

// applyMiddleware applies multiple middleware functions to a handler
func applyMiddleware(h http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}
