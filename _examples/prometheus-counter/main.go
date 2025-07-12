package main

import (
	"net/http"
	"os"

	"atomicgo.dev/service"
)

func main() {
	svc := service.New("prometheus-counter-service", nil)

	// Register a simple custom counter to demonstrate the metrics system
	err := svc.RegisterCounter(service.MetricConfig{
		Name:   "demo_counter",
		Help:   "Total number of demo events processed",
		Labels: []string{"event_type", "result"},
	})
	if err != nil {
		svc.Logger.Error("Failed to register demo counter", "error", err)
		os.Exit(1)
	}

	// Main handler
	svc.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Hello, World!"))
	})

	svc.HandleFunc("/demo", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Demo counter incremented"))
	})

	svc.Logger.Info("Service available at http://localhost:8080")
	svc.Logger.Info("Demo counter at http://localhost:8080/demo")
	svc.Logger.Info("Prometheus metrics at http://localhost:9090/metrics")

	if err := svc.Start(); err != nil {
		svc.Logger.Error("Failed to start service", "error", err)
		os.Exit(1)
	}
}
