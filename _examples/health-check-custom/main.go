package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"atomicgo.dev/service"
	"github.com/hellofresh/health-go/v5"
)

func main() {
	svc := service.New("custom-health-service", nil)

	// go-health provides an http health check, but for this example we'll build our own
	// This health check will pass, as the external API is reachable
	svc.RegisterHealthCheck(health.Config{
		Name:    "external-api",
		Timeout: time.Second * 5,
		Check:   checkExternalAPI("https://httb.dev/status/200"),
	})

	// This will fail, as the external API is not reachable
	svc.RegisterHealthCheck(health.Config{
		Name:    "external-api-failing",
		Timeout: time.Second * 5,
		Check:   checkExternalAPI("https://httb.dev/status/404"),
	})

	// Main handler
	svc.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	svc.Logger.Info("Service available at http://localhost:8080")
	svc.Logger.Info("Health check at http://localhost:9090/health")

	if err := svc.Start(); err != nil {
		svc.Logger.Error("Failed to start service", "error", err)
		os.Exit(1)
	}
}

// Custom health check function that checks if the external API is reachable
func checkExternalAPI(url string) health.CheckFunc {
	return func(ctx context.Context) error {
		client := &http.Client{Timeout: 3 * time.Second}

		// Replace with a non-existent URL to simulate a failed health check
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
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
	}
}
