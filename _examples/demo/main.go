package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"atomicgo.dev/service"
	"github.com/hellofresh/health-go/v5"
)

func main() {
	// Load configuration from environment variables
	config, err := service.LoadFromEnv()
	if err != nil {
		// We can't use svc.Logger here yet, so we'll use slog for this error
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

	// Register health checks
	svc.RegisterHealthCheck(health.Config{
		Name:      "database",
		Timeout:   time.Second * 5,
		SkipOnErr: false, // This check is critical
		Check: func(ctx context.Context) error {
			// Simulate database health check
			// In a real application, you would check your database connection
			slog.Info("checking database health")
			time.Sleep(100 * time.Millisecond) // Simulate some work
			return nil                         // Return nil for healthy, error for unhealthy
		},
	})

	svc.RegisterHealthCheck(health.Config{
		Name:      "cache",
		Timeout:   time.Second * 3,
		SkipOnErr: true, // This check is optional
		Check: func(ctx context.Context) error {
			// Simulate cache health check
			// In a real application, you would check your Redis/Memcached connection
			slog.Info("checking cache health")
			time.Sleep(50 * time.Millisecond) // Simulate some work
			return nil                        // Return nil for healthy, error for unhealthy
		},
	})

	svc.RegisterHealthCheck(health.Config{
		Name:      "external-api",
		Timeout:   time.Second * 10,
		SkipOnErr: true, // External dependencies are often optional
		Check: func(ctx context.Context) error {
			// Simulate external API health check
			slog.Info("checking external API health")

			// Create a simple HTTP request to check external service
			client := &http.Client{Timeout: 5 * time.Second}
			req, err := http.NewRequestWithContext(ctx, "GET", "https://httb.dev/status/200", nil)
			if err != nil {
				return err
			}

			resp, err := client.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("external API returned status %d", resp.StatusCode)
			}

			return nil
		},
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
	svc.HandleFunc("/health-demo", handleHealthDemo)
	svc.HandleFunc("/metrics-demo", handleMetricsDemo)

	// Start service with graceful shutdown
	svc.Logger.Info("starting service with graceful shutdown support")
	svc.Logger.Info("health endpoints available at:")
	svc.Logger.Info("  - http://localhost:9090/health (comprehensive health check)")
	svc.Logger.Info("  - http://localhost:9090/ready (readiness probe)")
	svc.Logger.Info("  - http://localhost:9090/live (liveness probe)")
	svc.Logger.Info("  - http://localhost:9090/metrics (prometheus metrics)")

	if err := svc.Start(); err != nil {
		svc.Logger.Error("failed to start service", "error", err)
		os.Exit(1)
	}

	svc.Logger.Info("service stopped")
}

func handleHelloWorld(w http.ResponseWriter, r *http.Request) {
	logger := service.GetLogger(r)
	logger.Info("Hello, World! endpoint called")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello, World!"))
}

func handleHealthDemo(w http.ResponseWriter, r *http.Request) {
	logger := service.GetLogger(r)
	logger.Info("health demo endpoint called")

	// Access the health checker from the request context
	healthChecker := service.GetHealthChecker(r)
	if healthChecker == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Health checker not available"))
		return
	}

	// Get current health status
	check := healthChecker.Measure(r.Context())

	// Return health status information
	w.Header().Set("Content-Type", "application/json")
	if check.Status == "OK" {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	// In a real application, you might want to use json.Marshal
	response := fmt.Sprintf(`{
		"status": "%s",
		"timestamp": "%s",
		"component": {
			"name": "%s",
			"version": "%s"
		}
	}`, check.Status, check.Timestamp.Format(time.RFC3339), check.Component.Name, check.Component.Version)

	w.Write([]byte(response))
}

func handleMetricsDemo(w http.ResponseWriter, r *http.Request) {
	logger := service.GetLogger(r)
	logger.Info("metrics demo endpoint called")

	// Simulate some work that might be measured
	time.Sleep(100 * time.Millisecond)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Metrics demo - check /metrics endpoint for Prometheus metrics"))
}
