package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"atomicgo.dev/service"
)

func main() {
	svc := service.New("custom-metrics-service", nil)

	// Register custom business metrics
	err := svc.RegisterCounter(service.MetricConfig{
		Name:   "user_registrations_total",
		Help:   "Total number of user registrations",
		Labels: []string{"source", "status"},
	})
	if err != nil {
		svc.Logger.Error("Failed to register user registrations counter", "error", err)
		os.Exit(1)
	}

	err = svc.RegisterCounter(service.MetricConfig{
		Name:   "orders_total",
		Help:   "Total number of orders placed",
		Labels: []string{"product_category", "payment_method"},
	})
	if err != nil {
		svc.Logger.Error("Failed to register orders counter", "error", err)
		os.Exit(1)
	}

	err = svc.RegisterGauge(service.MetricConfig{
		Name:   "active_users",
		Help:   "Number of currently active users",
		Labels: []string{"user_type"},
	})
	if err != nil {
		svc.Logger.Error("Failed to register active users gauge", "error", err)
		os.Exit(1)
	}

	err = svc.RegisterGauge(service.MetricConfig{
		Name:   "queue_size",
		Help:   "Current size of processing queues",
		Labels: []string{"queue_name"},
	})
	if err != nil {
		svc.Logger.Error("Failed to register queue size gauge", "error", err)
		os.Exit(1)
	}

	err = svc.RegisterHistogram(service.MetricConfig{
		Name:    "request_processing_duration_seconds",
		Help:    "Time spent processing business requests",
		Labels:  []string{"operation", "result"},
		Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0, 2.0, 5.0},
	})
	if err != nil {
		svc.Logger.Error("Failed to register request processing duration histogram", "error", err)
		os.Exit(1)
	}

	err = svc.RegisterSummary(service.MetricConfig{
		Name:   "response_size_bytes",
		Help:   "Size of API responses in bytes",
		Labels: []string{"endpoint", "content_type"},
		Objectives: map[float64]float64{
			0.5:  0.05,
			0.9:  0.01,
			0.95: 0.005,
			0.99: 0.001,
		},
	})
	if err != nil {
		svc.Logger.Error("Failed to register response size summary", "error", err)
		os.Exit(1)
	}

	// Simulate background metric updates
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Simulate changing active users
				premiumUsers := rand.Intn(50) + 20
				freeUsers := rand.Intn(200) + 100

				svc.Metrics.SetGauge("active_users", float64(premiumUsers), "premium")
				svc.Metrics.SetGauge("active_users", float64(freeUsers), "free")

				// Simulate queue sizes
				svc.Metrics.SetGauge("queue_size", float64(rand.Intn(20)), "email")
				svc.Metrics.SetGauge("queue_size", float64(rand.Intn(50)), "notifications")
				svc.Metrics.SetGauge("queue_size", float64(rand.Intn(30)), "analytics")

				svc.Logger.Info("Updated background metrics",
					"premium_users", premiumUsers,
					"free_users", freeUsers)
			}
		}
	}()

	// User registration endpoint
	svc.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		logger := service.GetLogger(r)
		start := time.Now()

		// Simulate registration logic
		source := r.URL.Query().Get("source")
		if source == "" {
			source = "web"
		}

		// Simulate random success/failure
		success := rand.Float32() > 0.1 // 90% success rate
		status := "success"
		if !success {
			status = "failure"
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Registration failed"))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Registration successful"))
		}

		// Record metrics
		err := service.IncCounter(r, "user_registrations_total", source, status)
		if err != nil {
			logger.Error("Failed to increment user registrations counter", "error", err)
		}

		// Record processing time
		duration := time.Since(start).Seconds()
		err = service.ObserveHistogram(r, "request_processing_duration_seconds", duration, "registration", status)
		if err != nil {
			logger.Error("Failed to observe processing duration", "error", err)
		}

		// Record response size
		responseSize := float64(len("Registration successful"))
		err = service.ObserveSummary(r, "response_size_bytes", responseSize, "/register", "text/plain")
		if err != nil {
			logger.Error("Failed to observe response size", "error", err)
		}

		logger.Info("User registration processed",
			"source", source,
			"status", status,
			"duration", duration)
	})

	// Order placement endpoint
	svc.HandleFunc("/order", func(w http.ResponseWriter, r *http.Request) {
		logger := service.GetLogger(r)
		start := time.Now()

		// Get parameters
		category := r.URL.Query().Get("category")
		if category == "" {
			category = "electronics"
		}

		paymentMethod := r.URL.Query().Get("payment")
		if paymentMethod == "" {
			paymentMethod = "credit_card"
		}

		// Simulate order processing
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)

		// Simulate random success/failure
		success := rand.Float32() > 0.05 // 95% success rate
		status := "success"
		if !success {
			status = "failure"
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Order failed"))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf("Order placed for %s via %s", category, paymentMethod)))
		}

		// Record metrics
		err := service.IncCounter(r, "orders_total", category, paymentMethod)
		if err != nil {
			logger.Error("Failed to increment orders counter", "error", err)
		}

		// Record processing time
		duration := time.Since(start).Seconds()
		err = service.ObserveHistogram(r, "request_processing_duration_seconds", duration, "order", status)
		if err != nil {
			logger.Error("Failed to observe processing duration", "error", err)
		}

		// Record response size
		responseContent := fmt.Sprintf("Order placed for %s via %s", category, paymentMethod)
		responseSize := float64(len(responseContent))
		err = service.ObserveSummary(r, "response_size_bytes", responseSize, "/order", "text/plain")
		if err != nil {
			logger.Error("Failed to observe response size", "error", err)
		}

		logger.Info("Order processed",
			"category", category,
			"payment_method", paymentMethod,
			"status", status,
			"duration", duration)
	})

	// Admin endpoint to manually update metrics
	svc.HandleFunc("/admin/metrics", func(w http.ResponseWriter, r *http.Request) {
		logger := service.GetLogger(r)

		switch r.Method {
		case http.MethodPost:
			// Simulate admin actions that affect metrics
			action := r.URL.Query().Get("action")

			switch action {
			case "add_users":
				// Simulate adding users
				count := rand.Intn(10) + 1
				err := service.AddGauge(r, "active_users", float64(count), "premium")
				if err != nil {
					logger.Error("Failed to add to active users gauge", "error", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.Write([]byte(fmt.Sprintf("Added %d premium users", count)))

			case "clear_queue":
				// Simulate clearing a queue
				queueName := r.URL.Query().Get("queue")
				if queueName == "" {
					queueName = "email"
				}
				err := service.SetGauge(r, "queue_size", 0, queueName)
				if err != nil {
					logger.Error("Failed to clear queue gauge", "error", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.Write([]byte(fmt.Sprintf("Cleared %s queue", queueName)))

			default:
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Unknown action"))
			}

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("Method not allowed"))
		}
	})

	// Status endpoint showing current metrics
	svc.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		logger := service.GetLogger(r)
		logger.Info("Status endpoint requested")

		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Custom Metrics Demo Service Status\n"))
		w.Write([]byte("==================================\n\n"))
		w.Write([]byte("Available endpoints:\n"))
		w.Write([]byte("- GET  /register?source=web|mobile|api\n"))
		w.Write([]byte("- GET  /order?category=electronics|books|clothing&payment=credit_card|paypal|crypto\n"))
		w.Write([]byte("- POST /admin/metrics?action=add_users|clear_queue&queue=email|notifications|analytics\n"))
		w.Write([]byte("- GET  /status (this endpoint)\n\n"))
		w.Write([]byte("Metrics available at: http://localhost:9090/metrics\n\n"))
		w.Write([]byte("Custom metrics registered:\n"))
		w.Write([]byte("- user_registrations_total: Counter tracking user registrations by source and status\n"))
		w.Write([]byte("- orders_total: Counter tracking orders by category and payment method\n"))
		w.Write([]byte("- active_users: Gauge showing current active users by type\n"))
		w.Write([]byte("- queue_size: Gauge showing current queue sizes\n"))
		w.Write([]byte("- request_processing_duration_seconds: Histogram of request processing times\n"))
		w.Write([]byte("- response_size_bytes: Summary of response sizes\n"))
		w.Write([]byte("\nBuilt-in HTTP metrics are also available (requests_total, request_duration, etc.)\n"))
	})

	// Root endpoint
	svc.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logger := service.GetLogger(r)
		logger.Info("Root endpoint accessed")

		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
    <title>Custom Metrics Demo</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        h1 { color: #333; }
        .endpoint { margin: 20px 0; padding: 15px; background: #f5f5f5; border-radius: 5px; }
        .method { font-weight: bold; color: #0066cc; }
        .url { font-family: monospace; background: #e8e8e8; padding: 2px 4px; }
        .description { margin-top: 10px; color: #666; }
        .metrics-link { display: inline-block; margin: 20px 0; padding: 10px 20px; background: #0066cc; color: white; text-decoration: none; border-radius: 5px; }
    </style>
</head>
<body>
    <h1>Custom Metrics Demo Service</h1>
    <p>This service demonstrates the new flexible metrics system with custom metrics registration and manipulation.</p>
    
    <div class="endpoint">
        <span class="method">GET</span> <span class="url">/register?source=web|mobile|api</span>
        <div class="description">Simulates user registration and tracks metrics by source and status.</div>
    </div>
    
    <div class="endpoint">
        <span class="method">GET</span> <span class="url">/order?category=electronics|books|clothing&payment=credit_card|paypal|crypto</span>
        <div class="description">Simulates order placement and tracks metrics by category and payment method.</div>
    </div>
    
    <div class="endpoint">
        <span class="method">POST</span> <span class="url">/admin/metrics?action=add_users|clear_queue&queue=email|notifications|analytics</span>
        <div class="description">Admin endpoint to manually manipulate metrics.</div>
    </div>
    
    <div class="endpoint">
        <span class="method">GET</span> <span class="url">/status</span>
        <div class="description">Shows service status and available endpoints.</div>
    </div>
    
    <a href="http://localhost:9090/metrics" class="metrics-link" target="_blank">View Prometheus Metrics</a>
    
    <h2>Try these commands:</h2>
    <pre>
# Generate some user registrations
curl "http://localhost:8080/register?source=web"
curl "http://localhost:8080/register?source=mobile"
curl "http://localhost:8080/register?source=api"

# Generate some orders
curl "http://localhost:8080/order?category=electronics&payment=credit_card"
curl "http://localhost:8080/order?category=books&payment=paypal"
curl "http://localhost:8080/order?category=clothing&payment=crypto"

# Admin actions
curl -X POST "http://localhost:8080/admin/metrics?action=add_users"
curl -X POST "http://localhost:8080/admin/metrics?action=clear_queue&queue=email"

# Check metrics
curl http://localhost:9090/metrics | grep -E "(user_registrations|orders_total|active_users|queue_size)"
    </pre>
</body>
</html>
		`))
	})

	svc.Logger.Info("Starting Custom Metrics Demo Service...")
	svc.Logger.Info("Service available at http://localhost:8080")
	svc.Logger.Info("Prometheus metrics at http://localhost:9090/metrics")
	svc.Logger.Info("")
	svc.Logger.Info("Custom metrics registered:")
	svc.Logger.Info("- user_registrations_total: Counter for user registrations")
	svc.Logger.Info("- orders_total: Counter for orders placed")
	svc.Logger.Info("- active_users: Gauge for active users")
	svc.Logger.Info("- queue_size: Gauge for queue sizes")
	svc.Logger.Info("- request_processing_duration_seconds: Histogram for processing times")
	svc.Logger.Info("- response_size_bytes: Summary for response sizes")
	svc.Logger.Info("")
	svc.Logger.Info("Try the endpoints:")
	svc.Logger.Info("  curl 'http://localhost:8080/register?source=web'")
	svc.Logger.Info("  curl 'http://localhost:8080/order?category=electronics&payment=credit_card'")
	svc.Logger.Info("  curl -X POST 'http://localhost:8080/admin/metrics?action=add_users'")
	svc.Logger.Info("  curl http://localhost:8080/status")

	if err := svc.Start(); err != nil {
		svc.Logger.Error("Failed to start service", "error", err)
		os.Exit(1)
	}
}
