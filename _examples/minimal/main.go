package main

import (
	"net/http"
	"os"

	"atomicgo.dev/service"
)

func main() {
	// Create service with default configuration
	svc := service.New("minimal-service", nil)

	// Simple handler
	svc.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logger := service.GetLogger(r)
		logger.Info("Hello from minimal service!")
		w.Write([]byte("Hello from minimal service!"))
	})

	// Start service - includes graceful shutdown, metrics, and health checks
	svc.Logger.Info("Starting minimal service...")
	svc.Logger.Info("Service available at http://localhost:8080")
	svc.Logger.Info("Health check at http://localhost:9090/health")
	svc.Logger.Info("Metrics at http://localhost:9090/metrics")

	if err := svc.Start(); err != nil {
		svc.Logger.Error("Failed to start service", "error", err)
		os.Exit(1)
	}
}
