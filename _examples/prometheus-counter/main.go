package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"atomicgo.dev/service"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Custom counters
	requestCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "myapp_requests_total",
			Help: "Total number of requests processed",
		},
		[]string{"method", "endpoint", "status"},
	)

	businessMetricCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "myapp_business_events_total",
			Help: "Total number of business events processed",
		},
		[]string{"event_type", "result"},
	)

	// Custom gauges
	activeUsersGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "myapp_active_users",
			Help: "Number of currently active users",
		},
	)

	queueSizeGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "myapp_queue_size",
			Help: "Current size of processing queues",
		},
		[]string{"queue_name"},
	)

	// Custom histograms
	processingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "myapp_processing_duration_seconds",
			Help:    "Time spent processing requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	// Custom summary
	requestSize = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "myapp_request_size_bytes",
			Help:       "Size of requests in bytes",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"endpoint"},
	)
)

func main() {
	svc := service.New("prometheus-counter-service", nil)

	// Simulate background metrics collection
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Simulate changing active users
				activeUsersGauge.Set(float64(rand.Intn(100) + 50))

				// Simulate queue sizes
				queueSizeGauge.WithLabelValues("email").Set(float64(rand.Intn(20)))
				queueSizeGauge.WithLabelValues("notifications").Set(float64(rand.Intn(50)))
				queueSizeGauge.WithLabelValues("analytics").Set(float64(rand.Intn(30)))

				// Simulate some business events
				eventTypes := []string{"user_signup", "purchase", "login", "logout"}
				results := []string{"success", "failure"}

				for i := 0; i < rand.Intn(5); i++ {
					eventType := eventTypes[rand.Intn(len(eventTypes))]
					result := results[rand.Intn(len(results))]
					businessMetricCounter.WithLabelValues(eventType, result).Inc()
				}
			}
		}
	}()

	// Custom middleware to track request metrics
	svc.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Track request size
			if r.ContentLength > 0 {
				requestSize.WithLabelValues(r.URL.Path).Observe(float64(r.ContentLength))
			}

			// Wrap response writer to capture status code
			wrapper := &responseWriter{ResponseWriter: w, statusCode: 200}

			// Process request
			next.ServeHTTP(wrapper, r)

			// Record metrics
			duration := time.Since(start)
			requestCounter.WithLabelValues(r.Method, r.URL.Path, strconv.Itoa(wrapper.statusCode)).Inc()
			processingDuration.WithLabelValues("http_request").Observe(duration.Seconds())
		})
	})

	// Main handler
	svc.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logger := service.GetLogger(r)
		logger.Info("Home page requested")

		w.Write([]byte("Prometheus Counter Demo Service\n"))
		w.Write([]byte("Available endpoints:\n"))
		w.Write([]byte("  - /api/users (GET, POST)\n"))
		w.Write([]byte("  - /api/orders (GET, POST)\n"))
		w.Write([]byte("  - /api/process (POST)\n"))
		w.Write([]byte("  - /metrics (Prometheus metrics)\n"))
	})

	// Users API endpoint
	svc.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		logger := service.GetLogger(r)

		switch r.Method {
		case "GET":
			logger.Info("Fetching users")
			// Simulate processing time
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)

			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"users": ["alice", "bob", "charlie"]}`))

		case "POST":
			logger.Info("Creating user")
			// Simulate processing time
			time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)

			// Simulate success/failure
			if rand.Float32() < 0.9 {
				businessMetricCounter.WithLabelValues("user_creation", "success").Inc()
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte(`{"status": "created"}`))
			} else {
				businessMetricCounter.WithLabelValues("user_creation", "failure").Inc()
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error": "validation failed"}`))
			}

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Orders API endpoint
	svc.HandleFunc("/api/orders", func(w http.ResponseWriter, r *http.Request) {
		logger := service.GetLogger(r)

		switch r.Method {
		case "GET":
			logger.Info("Fetching orders")
			time.Sleep(time.Duration(rand.Intn(150)) * time.Millisecond)

			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"orders": [{"id": 1, "amount": 100}, {"id": 2, "amount": 200}]}`))

		case "POST":
			logger.Info("Creating order")
			time.Sleep(time.Duration(rand.Intn(300)) * time.Millisecond)

			// Simulate success/failure
			if rand.Float32() < 0.8 {
				businessMetricCounter.WithLabelValues("order_creation", "success").Inc()
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte(`{"status": "created", "order_id": 123}`))
			} else {
				businessMetricCounter.WithLabelValues("order_creation", "failure").Inc()
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error": "insufficient funds"}`))
			}

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Heavy processing endpoint
	svc.HandleFunc("/api/process", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		logger := service.GetLogger(r)
		logger.Info("Processing heavy operation")

		// Simulate heavy processing
		start := time.Now()
		time.Sleep(time.Duration(rand.Intn(1000)+500) * time.Millisecond)
		duration := time.Since(start)

		// Record processing duration
		processingDuration.WithLabelValues("heavy_processing").Observe(duration.Seconds())

		// Simulate success/failure
		if rand.Float32() < 0.7 {
			businessMetricCounter.WithLabelValues("heavy_processing", "success").Inc()
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf(`{"status": "completed", "duration": "%v"}`, duration)))
		} else {
			businessMetricCounter.WithLabelValues("heavy_processing", "failure").Inc()
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "processing failed"}`))
		}
	})

	// Metrics summary endpoint
	svc.HandleFunc("/metrics-summary", func(w http.ResponseWriter, r *http.Request) {
		logger := service.GetLogger(r)
		logger.Info("Metrics summary requested")

		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Custom Metrics Summary\n"))
		w.Write([]byte("======================\n\n"))
		w.Write([]byte("Available custom metrics:\n"))
		w.Write([]byte("- myapp_requests_total: Total HTTP requests by method, endpoint, and status\n"))
		w.Write([]byte("- myapp_business_events_total: Business events by type and result\n"))
		w.Write([]byte("- myapp_active_users: Current number of active users\n"))
		w.Write([]byte("- myapp_queue_size: Size of processing queues\n"))
		w.Write([]byte("- myapp_processing_duration_seconds: Request processing time\n"))
		w.Write([]byte("- myapp_request_size_bytes: Size of incoming requests\n\n"))
		w.Write([]byte("View all metrics at: http://localhost:9090/metrics\n"))
	})

	svc.Logger.Info("Starting Prometheus counter demo service...")
	svc.Logger.Info("Service available at http://localhost:8080")
	svc.Logger.Info("Prometheus metrics at http://localhost:9090/metrics")
	svc.Logger.Info("Metrics summary at http://localhost:8080/metrics-summary")
	svc.Logger.Info("")
	svc.Logger.Info("Try making requests to generate metrics:")
	svc.Logger.Info("  curl http://localhost:8080/api/users")
	svc.Logger.Info("  curl -X POST http://localhost:8080/api/users")
	svc.Logger.Info("  curl -X POST http://localhost:8080/api/process")
	svc.Logger.Info("")
	svc.Logger.Info("Then check metrics at: http://localhost:9090/metrics")

	if err := svc.Start(); err != nil {
		svc.Logger.Error("Failed to start service", "error", err)
		os.Exit(1)
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
