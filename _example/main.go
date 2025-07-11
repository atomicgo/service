package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"atomicgo.dev/service"
)

func main() {
	// Load configuration from environment variables
	config, err := service.LoadFromEnv()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Create service with loaded configuration
	svc := service.New("example", config)

	// Add custom middleware
	svc.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Service", "example")
			next.ServeHTTP(w, r)
		})
	})

	// Add shutdown hook to demonstrate graceful shutdown
	svc.AddShutdownHook(func() error {
		slog.Info("cleaning up resources...")
		time.Sleep(1 * time.Second) // Simulate cleanup
		slog.Info("cleanup complete")
		return nil
	})

	// Register handlers
	svc.HandleFunc("/", handleHelloWorld)
	svc.HandleFunc("/health", handleHealth)
	svc.HandleFunc("/metrics-demo", handleMetricsDemo)

	// Start service with graceful shutdown
	slog.Info("starting service with graceful shutdown support")
	if err := svc.Start(); err != nil {
		svc.Logger.Error("failed to start service", "error", err)
		os.Exit(1)
	}

	slog.Info("service stopped")
}

func handleHelloWorld(w http.ResponseWriter, r *http.Request) {
	logger := service.GetLogger(r)
	logger.Info("Hello, World! endpoint called")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello, World!"))
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	logger := service.GetLogger(r)
	logger.Info("health check called")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func handleMetricsDemo(w http.ResponseWriter, r *http.Request) {
	logger := service.GetLogger(r)
	logger.Info("metrics demo endpoint called")

	// Simulate some work
	time.Sleep(100 * time.Millisecond)

	// Example of using metrics in a handler
	// The metrics middleware automatically tracks requests, but you can also
	// interact with metrics manually if needed
	metrics := service.GetMetrics(r)
	if metrics != nil {
		logger.Info("metrics collector is available in context")
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Metrics demo completed. Check :9090/metrics for Prometheus metrics.")))
}
