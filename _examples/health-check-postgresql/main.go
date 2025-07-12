package main

import (
	"net/http"
	"os"
	"time"

	"atomicgo.dev/service"
	"github.com/hellofresh/health-go/v5"
	healthPostgres "github.com/hellofresh/health-go/v5/checks/postgres"
	_ "github.com/lib/pq"
)

func main() {
	// Database connection string - in production, use environment variables
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://user:password@localhost/dbname?sslmode=disable"
	}

	// Create service
	svc := service.New("postgresql-health-service", nil)

	// Register PostgreSQL health check using built-in checker
	svc.RegisterHealthCheck(health.Config{
		Name:      "postgresql",
		Timeout:   time.Second * 5,
		SkipOnErr: false, // Critical check - service is unhealthy if DB is down
		Check: healthPostgres.New(healthPostgres.Config{
			DSN: dbURL,
		}),
	})

	// Simple handler
	svc.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	svc.Logger.Info("Service available at http://localhost:8080")
	svc.Logger.Info("Health check at http://localhost:9090/health")
	svc.Logger.Info("Database URL: " + dbURL)

	if err := svc.Start(); err != nil {
		svc.Logger.Error("Failed to start service", "error", err)
		os.Exit(1)
	}
}
