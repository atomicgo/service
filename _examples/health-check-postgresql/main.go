package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	"atomicgo.dev/service"
	"github.com/hellofresh/health-go/v5"
	healthPostgres "github.com/hellofresh/health-go/v5/checks/postgres"
	_ "github.com/lib/pq" // PostgreSQL driver
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

	// Open database connection for stats endpoint
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		svc.Logger.Error("Failed to open database connection", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Simple handler
	svc.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logger := service.GetLogger(r)
		logger.Info("PostgreSQL health check service")
		w.Write([]byte("PostgreSQL Health Check Service\n"))
		w.Write([]byte("Check health at: /health\n"))
	})

	// Database stats endpoint
	svc.HandleFunc("/db-stats", func(w http.ResponseWriter, r *http.Request) {
		logger := service.GetLogger(r)
		logger.Info("Database stats requested")

		stats := db.Stats()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{
			"open_connections": %d,
			"in_use": %d,
			"idle": %d,
			"wait_count": %d,
			"wait_duration": "%s",
			"max_idle_closed": %d,
			"max_lifetime_closed": %d
		}`, stats.OpenConnections, stats.InUse, stats.Idle, stats.WaitCount,
			stats.WaitDuration, stats.MaxIdleClosed, stats.MaxLifetimeClosed)))
	})

	svc.Logger.Info("Starting PostgreSQL health check service...")
	svc.Logger.Info("Service available at http://localhost:8080")
	svc.Logger.Info("Health check at http://localhost:9090/health")
	svc.Logger.Info("Database stats at http://localhost:8080/db-stats")
	svc.Logger.Info("Database URL: " + dbURL)

	if err := svc.Start(); err != nil {
		svc.Logger.Error("Failed to start service", "error", err)
		os.Exit(1)
	}
}
