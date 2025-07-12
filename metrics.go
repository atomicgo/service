package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsCollector holds all the metrics for the service with a flexible registry
type MetricsCollector struct {
	serviceName string
	registry    *prometheus.Registry
	mu          sync.RWMutex

	// Built-in HTTP metrics (always available)
	httpRequestsTotal    *prometheus.CounterVec
	httpRequestDuration  *prometheus.HistogramVec
	httpRequestsInFlight prometheus.Gauge

	// Custom metrics registry
	counters   map[string]*prometheus.CounterVec
	gauges     map[string]*prometheus.GaugeVec
	histograms map[string]*prometheus.HistogramVec
	summaries  map[string]*prometheus.SummaryVec
}

// MetricConfig holds configuration for creating custom metrics
type MetricConfig struct {
	Name       string
	Help       string
	Labels     []string
	Buckets    []float64           // For histograms
	Objectives map[float64]float64 // For summaries
}

// NewMetricsCollector creates a new metrics collector with a flexible registry
func NewMetricsCollector(serviceName string) *MetricsCollector {
	registry := prometheus.NewRegistry()

	metricsCollector := &MetricsCollector{
		serviceName: serviceName,
		registry:    registry,
		counters:    make(map[string]*prometheus.CounterVec),
		gauges:      make(map[string]*prometheus.GaugeVec),
		histograms:  make(map[string]*prometheus.HistogramVec),
		summaries:   make(map[string]*prometheus.SummaryVec),
	}

	// Create built-in HTTP metrics
	metricsCollector.httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: serviceName + "_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status_code"},
	)

	metricsCollector.httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    serviceName + "_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint", "status_code"},
	)

	metricsCollector.httpRequestsInFlight = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: serviceName + "_http_requests_in_flight",
			Help: "Number of HTTP requests currently being processed",
		},
	)

	// Register built-in metrics
	registry.MustRegister(metricsCollector.httpRequestsTotal)
	registry.MustRegister(metricsCollector.httpRequestDuration)
	registry.MustRegister(metricsCollector.httpRequestsInFlight)

	return metricsCollector
}

// RegisterCounter registers a new counter metric
func (mc *MetricsCollector) RegisterCounter(config MetricConfig) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Ensure metric name has service prefix
	prefixedName := mc.ensureMetricNamePrefix(config.Name)

	if _, exists := mc.counters[prefixedName]; exists {
		return fmt.Errorf("counter %s already exists", prefixedName) //nolint:err113
	}

	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: prefixedName,
			Help: config.Help,
		},
		config.Labels,
	)

	if err := mc.registry.Register(counter); err != nil {
		return fmt.Errorf("failed to register counter %s: %w", prefixedName, err)
	}

	mc.counters[prefixedName] = counter

	return nil
}

// RegisterGauge registers a new gauge metric
func (mc *MetricsCollector) RegisterGauge(config MetricConfig) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Ensure metric name has service prefix
	prefixedName := mc.ensureMetricNamePrefix(config.Name)

	if _, exists := mc.gauges[prefixedName]; exists {
		return fmt.Errorf("gauge %s already exists", prefixedName) //nolint:err113
	}

	gauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: prefixedName,
			Help: config.Help,
		},
		config.Labels,
	)

	if err := mc.registry.Register(gauge); err != nil {
		return fmt.Errorf("failed to register gauge %s: %w", prefixedName, err)
	}

	mc.gauges[prefixedName] = gauge

	return nil
}

// RegisterHistogram registers a new histogram metric
func (mc *MetricsCollector) RegisterHistogram(config MetricConfig) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Ensure metric name has service prefix
	prefixedName := mc.ensureMetricNamePrefix(config.Name)

	if _, exists := mc.histograms[prefixedName]; exists {
		return fmt.Errorf("histogram %s already exists", prefixedName) //nolint:err113
	}

	buckets := config.Buckets
	if len(buckets) == 0 {
		buckets = prometheus.DefBuckets
	}

	histogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    prefixedName,
			Help:    config.Help,
			Buckets: buckets,
		},
		config.Labels,
	)

	if err := mc.registry.Register(histogram); err != nil {
		return fmt.Errorf("failed to register histogram %s: %w", prefixedName, err)
	}

	mc.histograms[prefixedName] = histogram

	return nil
}

// RegisterSummary registers a new summary metric
func (mc *MetricsCollector) RegisterSummary(config MetricConfig) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Ensure metric name has service prefix
	prefixedName := mc.ensureMetricNamePrefix(config.Name)

	if _, exists := mc.summaries[prefixedName]; exists {
		return fmt.Errorf("summary %s already exists", prefixedName) //nolint:err113
	}

	objectives := config.Objectives
	if len(objectives) == 0 {
		objectives = map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001}
	}

	summary := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       prefixedName,
			Help:       config.Help,
			Objectives: objectives,
		},
		config.Labels,
	)

	if err := mc.registry.Register(summary); err != nil {
		return fmt.Errorf("failed to register summary %s: %w", prefixedName, err)
	}

	mc.summaries[prefixedName] = summary

	return nil
}

// IncCounter increments a counter metric
func (mc *MetricsCollector) IncCounter(name string, labels ...string) error {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	// Ensure metric name has service prefix
	prefixedName := mc.ensureMetricNamePrefix(name)

	counter, exists := mc.counters[prefixedName]
	if !exists {
		return fmt.Errorf("counter %s not found", prefixedName) //nolint:err113
	}

	counter.WithLabelValues(labels...).Inc()

	return nil
}

// AddCounter adds a value to a counter metric
func (mc *MetricsCollector) AddCounter(name string, value float64, labels ...string) error {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	// Ensure metric name has service prefix
	prefixedName := mc.ensureMetricNamePrefix(name)

	counter, exists := mc.counters[prefixedName]
	if !exists {
		return fmt.Errorf("counter %s not found", prefixedName) //nolint:err113
	}

	counter.WithLabelValues(labels...).Add(value)

	return nil
}

// SetGauge sets a gauge metric value
func (mc *MetricsCollector) SetGauge(name string, value float64, labels ...string) error {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	// Ensure metric name has service prefix
	prefixedName := mc.ensureMetricNamePrefix(name)

	gauge, exists := mc.gauges[prefixedName]
	if !exists {
		return fmt.Errorf("gauge %s not found", prefixedName) //nolint:err113
	}

	gauge.WithLabelValues(labels...).Set(value)

	return nil
}

// IncGauge increments a gauge metric
func (mc *MetricsCollector) IncGauge(name string, labels ...string) error {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	// Ensure metric name has service prefix
	prefixedName := mc.ensureMetricNamePrefix(name)

	gauge, exists := mc.gauges[prefixedName]
	if !exists {
		return fmt.Errorf("gauge %s not found", prefixedName) //nolint:err113
	}

	gauge.WithLabelValues(labels...).Inc()

	return nil
}

// DecGauge decrements a gauge metric
func (mc *MetricsCollector) DecGauge(name string, labels ...string) error {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	// Ensure metric name has service prefix
	prefixedName := mc.ensureMetricNamePrefix(name)

	gauge, exists := mc.gauges[prefixedName]
	if !exists {
		return fmt.Errorf("gauge %s not found", prefixedName) //nolint:err113
	}

	gauge.WithLabelValues(labels...).Dec()

	return nil
}

// AddGauge adds a value to a gauge metric
func (mc *MetricsCollector) AddGauge(name string, value float64, labels ...string) error {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	// Ensure metric name has service prefix
	prefixedName := mc.ensureMetricNamePrefix(name)

	gauge, exists := mc.gauges[prefixedName]
	if !exists {
		return fmt.Errorf("gauge %s not found", prefixedName) //nolint:err113
	}

	gauge.WithLabelValues(labels...).Add(value)

	return nil
}

// ObserveHistogram observes a value in a histogram metric
func (mc *MetricsCollector) ObserveHistogram(name string, value float64, labels ...string) error {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	// Ensure metric name has service prefix
	prefixedName := mc.ensureMetricNamePrefix(name)

	histogram, exists := mc.histograms[prefixedName]
	if !exists {
		return fmt.Errorf("histogram %s not found", prefixedName) //nolint:err113
	}

	histogram.WithLabelValues(labels...).Observe(value)

	return nil
}

// ObserveSummary observes a value in a summary metric
func (mc *MetricsCollector) ObserveSummary(name string, value float64, labels ...string) error {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	// Ensure metric name has service prefix
	prefixedName := mc.ensureMetricNamePrefix(name)

	summary, exists := mc.summaries[prefixedName]
	if !exists {
		return fmt.Errorf("summary %s not found", prefixedName) //nolint:err113
	}

	summary.WithLabelValues(labels...).Observe(value)

	return nil
}

// GetRegistry returns the Prometheus registry for custom integrations
func (mc *MetricsCollector) GetRegistry() *prometheus.Registry {
	return mc.registry
}

// ensureMetricNamePrefix ensures the metric name has the service name prefix
func (mc *MetricsCollector) ensureMetricNamePrefix(name string) string {
	if !strings.HasPrefix(name, mc.serviceName+"_") {
		return mc.serviceName + "_" + name
	}

	return name
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

// Helper functions for easy metric manipulation from handlers

// IncCounter increments a counter metric from a request context
func IncCounter(r *http.Request, name string, labels ...string) error {
	metrics := GetMetrics(r)
	if metrics == nil {
		return errors.New("metrics not available in request context") //nolint:err113
	}

	return metrics.IncCounter(name, labels...)
}

// AddCounter adds a value to a counter metric from a request context
func AddCounter(r *http.Request, name string, value float64, labels ...string) error {
	metrics := GetMetrics(r)
	if metrics == nil {
		return errors.New("metrics not available in request context") //nolint:err113
	}

	return metrics.AddCounter(name, value, labels...)
}

// SetGauge sets a gauge metric value from a request context
func SetGauge(r *http.Request, name string, value float64, labels ...string) error {
	metrics := GetMetrics(r)
	if metrics == nil {
		return errors.New("metrics not available in request context") //nolint:err113
	}

	return metrics.SetGauge(name, value, labels...)
}

// IncGauge increments a gauge metric from a request context
func IncGauge(r *http.Request, name string, labels ...string) error {
	metrics := GetMetrics(r)
	if metrics == nil {
		return errors.New("metrics not available in request context") //nolint:err113
	}

	return metrics.IncGauge(name, labels...)
}

// DecGauge decrements a gauge metric from a request context
func DecGauge(r *http.Request, name string, labels ...string) error {
	metrics := GetMetrics(r)
	if metrics == nil {
		return errors.New("metrics not available in request context") //nolint:err113
	}

	return metrics.DecGauge(name, labels...)
}

// AddGauge adds a value to a gauge metric from a request context
func AddGauge(r *http.Request, name string, value float64, labels ...string) error {
	metrics := GetMetrics(r)
	if metrics == nil {
		return errors.New("metrics not available in request context") //nolint:err113
	}

	return metrics.AddGauge(name, value, labels...)
}

// ObserveHistogram observes a value in a histogram metric from a request context
func ObserveHistogram(r *http.Request, name string, value float64, labels ...string) error {
	metrics := GetMetrics(r)
	if metrics == nil {
		return errors.New("metrics not available in request context") //nolint:err113
	}

	return metrics.ObserveHistogram(name, value, labels...)
}

// ObserveSummary observes a value in a summary metric from a request context
func ObserveSummary(r *http.Request, name string, value float64, labels ...string) error {
	metrics := GetMetrics(r)
	if metrics == nil {
		return errors.New("metrics not available in request context") //nolint:err113
	}

	return metrics.ObserveSummary(name, value, labels...)
}

// startMetricsServer starts the Prometheus metrics server
func (s *Service) startMetricsServer() error {
	mux := http.NewServeMux()

	// Use the custom registry from metrics collector
	handler := promhttp.HandlerFor(s.Metrics.GetRegistry(), promhttp.HandlerOpts{})
	mux.Handle(s.Config.MetricsPath, handler)

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
		mux.HandleFunc(s.Config.HealthPath, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		})
		mux.HandleFunc(s.Config.ReadinessPath, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Ready"))
		})
		mux.HandleFunc(s.Config.LivenessPath, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Alive"))
		})
	}

	s.metricsServer = &http.Server{
		Addr:         s.Config.MetricsAddr,
		Handler:      mux,
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 5 * time.Minute,
		IdleTimeout:  5 * time.Minute,
	}

	s.Logger.Info("starting metrics server", "addr", s.Config.MetricsAddr, "path", s.Config.MetricsPath)

	return s.metricsServer.ListenAndServe() //nolint:wrapcheck
}
