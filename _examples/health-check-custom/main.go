package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"atomicgo.dev/service"
	"github.com/hellofresh/health-go/v5"
)

var (
	// Simulate application state
	isReady      int64 = 0
	startTime          = time.Now()
	requestCount int64
)

func main() {
	svc := service.New("custom-health-service", nil)

	// Custom readiness check - simulates application warm-up
	svc.RegisterHealthCheck(health.Config{
		Name:      "readiness",
		Timeout:   time.Second * 3,
		SkipOnErr: false,
		Check: func(ctx context.Context) error {
			if atomic.LoadInt64(&isReady) == 0 {
				return fmt.Errorf("application is still warming up")
			}
			return nil
		},
	})

	// Custom uptime check
	svc.RegisterHealthCheck(health.Config{
		Name:      "uptime",
		Timeout:   time.Second * 1,
		SkipOnErr: true, // Non-critical check
		Check: func(ctx context.Context) error {
			uptime := time.Since(startTime)
			if uptime < 10*time.Second {
				return fmt.Errorf("service recently started (uptime: %v)", uptime)
			}
			return nil
		},
	})

	// Custom external dependency check
	svc.RegisterHealthCheck(health.Config{
		Name:      "external-api",
		Timeout:   time.Second * 5,
		SkipOnErr: true, // External dependencies are often optional
		Check: func(ctx context.Context) error {
			// Check external service availability
			client := &http.Client{Timeout: 3 * time.Second}
			req, err := http.NewRequestWithContext(ctx, "GET", "https://httpbin.org/status/200", nil)
			if err != nil {
				return fmt.Errorf("failed to create request: %w", err)
			}

			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("external API unreachable: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("external API returned status %d", resp.StatusCode)
			}

			return nil
		},
	})

	// Main handler
	svc.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logger := service.GetLogger(r)
		logger.Info("Request received")

		// Increment request counter
		atomic.AddInt64(&requestCount, 1)

		w.Write([]byte("Custom Health Check Service\n"))
		w.Write([]byte(fmt.Sprintf("Uptime: %v\n", time.Since(startTime))))
		w.Write([]byte(fmt.Sprintf("Request count: %d\n", atomic.LoadInt64(&requestCount))))
		w.Write([]byte("Health check at: /health\n"))
	})

	// Simulate application warm-up
	go func() {
		svc.Logger.Info("Starting application warm-up...")
		time.Sleep(5 * time.Second)
		atomic.StoreInt64(&isReady, 1)
		svc.Logger.Info("Application warm-up complete, now ready")
	}()

	svc.Logger.Info("Starting custom health check service...")
	svc.Logger.Info("Service available at http://localhost:8080")
	svc.Logger.Info("Health check at http://localhost:9090/health")
	svc.Logger.Info("Toggle readiness with: POST http://localhost:8080/ready")
	svc.Logger.Info("View stats at http://localhost:8080/stats")

	if err := svc.Start(); err != nil {
		svc.Logger.Error("Failed to start service", "error", err)
		os.Exit(1)
	}
}
