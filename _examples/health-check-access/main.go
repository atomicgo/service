package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"atomicgo.dev/service"
	"github.com/hellofresh/health-go/v5"
	healthHttp "github.com/hellofresh/health-go/v5/checks/http"
	_ "github.com/lib/pq"
)

func main() {
	// Create service
	svc := service.New("accessing-health-checker-from-handlers", nil)

	// Register external API health check using built-in checker
	svc.RegisterHealthCheck(health.Config{
		Name:      "external-api-should-success",
		Timeout:   time.Second * 5,
		SkipOnErr: false,
		Check: healthHttp.New(healthHttp.Config{
			URL: "https://httb.dev/status/200",
		}),
	})

	svc.RegisterHealthCheck(health.Config{
		Name:      "external-api-should-fail",
		Timeout:   time.Second * 5,
		SkipOnErr: false,
		Check: healthHttp.New(healthHttp.Config{
			URL: "https://httb.dev/status/503",
		}),
	})

	// Simple handler that accesses the health checker
	svc.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		healthChecker := service.GetHealthChecker(r)
		if healthChecker != nil {
			check := healthChecker.Measure(r.Context())
			// Pretty print the check
			json, err := json.MarshalIndent(check, "", "  ")
			if err != nil {
				w.Write([]byte(fmt.Sprintf("Error marshalling check: %v", err)))
			}
			w.Write(json)
		}
	})

	svc.Logger.Info("Service available at http://localhost:8080")
	svc.Logger.Info("Health check at http://localhost:9090/health")

	if err := svc.Start(); err != nil {
		svc.Logger.Error("Failed to start service", "error", err)
		os.Exit(1)
	}
}
