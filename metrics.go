package service

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsCollector holds all the metrics for the service
type MetricsCollector struct {
	httpRequestsTotal    *prometheus.CounterVec
	httpRequestDuration  *prometheus.HistogramVec
	httpRequestsInFlight prometheus.Gauge
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(serviceName string) *MetricsCollector {
	mc := &MetricsCollector{
		httpRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: serviceName + "_http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status_code"},
		),
		httpRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    serviceName + "_http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint", "status_code"},
		),
		httpRequestsInFlight: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: serviceName + "_http_requests_in_flight",
				Help: "Number of HTTP requests currently being processed",
			},
		),
	}

	// Register metrics with Prometheus (ignore if already registered)
	prometheus.DefaultRegisterer.Register(mc.httpRequestsTotal)
	prometheus.DefaultRegisterer.Register(mc.httpRequestDuration)
	prometheus.DefaultRegisterer.Register(mc.httpRequestsInFlight)

	return mc
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// MetricsMiddleware creates middleware that records HTTP metrics
func MetricsMiddleware(metrics *MetricsCollector) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Add metrics collector to context
			ctx := context.WithValue(r.Context(), MetricsKey, metrics)
			r = r.WithContext(ctx)

			// Track in-flight requests
			metrics.httpRequestsInFlight.Inc()
			defer metrics.httpRequestsInFlight.Dec()

			// Create wrapped response writer to capture status code
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     200, // Default status code
			}

			// Record request start time
			start := time.Now()

			// Call the next handler
			next.ServeHTTP(wrapped, r)

			// Record metrics
			duration := time.Since(start).Seconds()
			statusCode := strconv.Itoa(wrapped.statusCode)

			metrics.httpRequestsTotal.WithLabelValues(
				r.Method, r.URL.Path, statusCode,
			).Inc()

			metrics.httpRequestDuration.WithLabelValues(
				r.Method, r.URL.Path, statusCode,
			).Observe(duration)
		})
	}
}

// GetMetrics retrieves the metrics collector from the request context
func GetMetrics(r *http.Request) *MetricsCollector {
	metrics, ok := r.Context().Value(MetricsKey).(*MetricsCollector)
	if !ok {
		return nil
	}
	return metrics
}

// IncCounter increments a counter metric
func IncCounter(r *http.Request, name string, labels ...string) {
	metrics := GetMetrics(r)
	if metrics == nil {
		return
	}

	// This is a simplified version - in a real implementation,
	// you'd want to have a more flexible metric registration system
	switch name {
	case "http_requests_total":
		if len(labels) >= 3 {
			metrics.httpRequestsTotal.WithLabelValues(labels[0], labels[1], labels[2]).Inc()
		}
	}
}

// ObserveHistogram observes a histogram metric
func ObserveHistogram(r *http.Request, name string, value float64, labels ...string) {
	metrics := GetMetrics(r)
	if metrics == nil {
		return
	}

	switch name {
	case "http_request_duration_seconds":
		if len(labels) >= 3 {
			metrics.httpRequestDuration.WithLabelValues(labels[0], labels[1], labels[2]).Observe(value)
		}
	}
}

// startMetricsServer starts the Prometheus metrics server
func (s *Service) startMetricsServer() error {
	mux := http.NewServeMux()
	mux.Handle(s.Config.MetricsPath, promhttp.Handler())

	// Add health check endpoints
	if s.HealthChecker != nil {
		// Main health check endpoint (comprehensive health status)
		mux.Handle(s.Config.HealthPath, s.HealthChecker.Handler())

		// Kubernetes readiness probe endpoint
		mux.HandleFunc(s.Config.ReadinessPath, s.HealthChecker.ReadinessHandler())

		// Kubernetes liveness probe endpoint
		mux.HandleFunc(s.Config.LivenessPath, s.HealthChecker.LivenessHandler())
	} else {
		// Fallback basic health endpoints if health checker is not available
		mux.HandleFunc(s.Config.HealthPath, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})
		mux.HandleFunc(s.Config.ReadinessPath, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Ready"))
		})
		mux.HandleFunc(s.Config.LivenessPath, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Alive"))
		})
	}

	s.metricsServer = &http.Server{
		Addr:    s.Config.MetricsAddr,
		Handler: mux,
	}

	s.Logger.Info("starting metrics server", "addr", s.Config.MetricsAddr, "path", s.Config.MetricsPath)
	return s.metricsServer.ListenAndServe()
}
