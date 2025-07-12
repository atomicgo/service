package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"atomicgo.dev/service"
)

var startTime = time.Now()

// Simulate various resources that need cleanup
type DatabaseConnection struct {
	connected bool
	mu        sync.Mutex
}

func (db *DatabaseConnection) Connect() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Note: We can't use svc.Logger here as it's not available in this context
	// In a real application, you'd pass the logger to this function
	time.Sleep(100 * time.Millisecond) // Simulate connection time
	db.connected = true
	return nil
}

func (db *DatabaseConnection) Close() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if !db.connected {
		return nil
	}

	// Note: We can't use svc.Logger here as it's not available in this context
	// In a real application, you'd pass the logger to this function
	time.Sleep(200 * time.Millisecond) // Simulate cleanup time
	db.connected = false
	return nil
}

func (db *DatabaseConnection) IsConnected() bool {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.connected
}

type CacheService struct {
	active bool
	mu     sync.Mutex
}

func (c *CacheService) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Note: We can't use svc.Logger here as it's not available in this context
	// In a real application, you'd pass the logger to this function
	time.Sleep(50 * time.Millisecond)
	c.active = true
	return nil
}

func (c *CacheService) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.active {
		return nil
	}

	// Note: We can't use svc.Logger here as it's not available in this context
	// In a real application, you'd pass the logger to this function
	time.Sleep(150 * time.Millisecond)
	c.active = false
	return nil
}

func main() {
	// Initialize resources
	db := &DatabaseConnection{}
	cache := &CacheService{}

	// Connect to resources
	if err := db.Connect(); err != nil {
		// We can't use svc.Logger here yet, so we'll use slog for this error
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	if err := cache.Start(); err != nil {
		// We can't use svc.Logger here yet, so we'll use slog for this error
		slog.Error("Failed to start cache service", "error", err)
		os.Exit(1)
	}

	// Create service
	svc := service.New("shutdown-hook-service", nil)

	// Register shutdown hooks in reverse order of initialization
	// The last registered hook runs first during shutdown

	// Hook 1: Cache cleanup (runs first during shutdown)
	svc.AddShutdownHook(func() error {
		slog.Info("Shutdown hook: Cleaning up cache service...")
		return cache.Stop()
	})

	// Hook 2: Database cleanup (runs second during shutdown)
	svc.AddShutdownHook(func() error {
		slog.Info("Shutdown hook: Cleaning up database connection...")
		return db.Close()
	})

	// Hook 3: Final cleanup (runs last during shutdown)
	svc.AddShutdownHook(func() error {
		slog.Info("Shutdown hook: Performing final cleanup...")

		// Simulate final cleanup operations
		slog.Info("Saving application state...")
		time.Sleep(100 * time.Millisecond)

		slog.Info("Flushing logs...")
		time.Sleep(50 * time.Millisecond)

		slog.Info("Final cleanup completed")
		return nil
	})

	// Hook 4: Demonstrate error handling in shutdown hooks
	svc.AddShutdownHook(func() error {
		slog.Info("Shutdown hook: Demonstrating error handling...")

		// Simulate a non-critical error during shutdown
		if time.Now().UnixNano()%2 == 0 {
			slog.Warn("Non-critical error during shutdown (this is expected)")
			return fmt.Errorf("simulated non-critical shutdown error")
		}

		slog.Info("Shutdown hook completed without errors")
		return nil
	})

	// Register handlers
	svc.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logger := service.GetLogger(r)
		logger.Info("Request received")

		w.Write([]byte("Shutdown Hook Demo Service\n"))
		w.Write([]byte(fmt.Sprintf("Database connected: %v\n", db.IsConnected())))
		w.Write([]byte("Send SIGTERM or SIGINT to trigger graceful shutdown\n"))
		w.Write([]byte("Press Ctrl+C to test shutdown hooks\n"))
	})

	// Status endpoint
	svc.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		logger := service.GetLogger(r)
		logger.Info("Status check requested")

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{
			"database_connected": %v,
			"cache_active": %v,
			"uptime": "%v"
		}`, db.IsConnected(), cache.active, time.Since(startTime))))
	})

	// Simulate some background work
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if db.IsConnected() {
					svc.Logger.Info("Background task: Database is healthy")
				}
			}
		}
	}()

	svc.Logger.Info("Starting shutdown hook demo service...")
	svc.Logger.Info("Service available at http://localhost:8080")
	svc.Logger.Info("Status at http://localhost:8080/status")
	svc.Logger.Info("Health check at http://localhost:9090/health")
	svc.Logger.Info("")
	svc.Logger.Info("To test shutdown hooks:")
	svc.Logger.Info("  - Press Ctrl+C")
	svc.Logger.Info("  - Send SIGTERM: kill -TERM <pid>")
	svc.Logger.Info("  - Send SIGINT: kill -INT <pid>")
	svc.Logger.Info("")
	svc.Logger.Info("Watch the logs to see shutdown hooks execute in order")

	if err := svc.Start(); err != nil {
		svc.Logger.Error("Failed to start service", "error", err)
		os.Exit(1)
	}

	svc.Logger.Info("Service shutdown complete")
}
