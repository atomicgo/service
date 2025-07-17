package service

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"
)

func TestNewMetricsCollector(t *testing.T) {
	t.Parallel()

	metrics := NewMetricsCollector("test-service")

	if metrics == nil {
		t.Fatal("expected metrics collector to be created")
	}

	if metrics.serviceName != "test-service" {
		t.Errorf("expected service name 'test-service', got %s", metrics.serviceName)
	}

	if metrics.registry == nil {
		t.Fatal("expected registry to be created")
	}

	if metrics.counters == nil {
		t.Fatal("expected counters map to be initialized")
	}

	if metrics.gauges == nil {
		t.Fatal("expected gauges map to be initialized")
	}

	if metrics.histograms == nil {
		t.Fatal("expected histograms map to be initialized")
	}

	if metrics.summaries == nil {
		t.Fatal("expected summaries map to be initialized")
	}

	// Check that built-in HTTP metrics are created
	if metrics.httpRequestsTotal == nil {
		t.Fatal("expected HTTP requests total metric to be created")
	}

	if metrics.httpRequestDuration == nil {
		t.Fatal("expected HTTP request duration metric to be created")
	}

	if metrics.httpRequestsInFlight == nil {
		t.Fatal("expected HTTP requests in flight metric to be created")
	}
}

func TestMetricsCollector_RegisterCounter(t *testing.T) {
	t.Parallel()

	t.Run("registers counter successfully", func(t *testing.T) {
		t.Parallel()

		metrics := NewMetricsCollector("test-service")

		config := MetricConfig{
			Name:   "test_counter",
			Help:   "Test counter metric",
			Labels: []string{"label1", "label2"},
		}

		err := metrics.RegisterCounter(config)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if _, exists := metrics.counters["test-service_test_counter"]; !exists {
			t.Error("expected counter to be registered")
		}
	})

	t.Run("fails to register duplicate counter", func(t *testing.T) {
		t.Parallel()

		metrics := NewMetricsCollector("test-service")

		config := MetricConfig{
			Name:   "duplicate_counter",
			Help:   "Duplicate counter metric",
			Labels: []string{"label1"},
		}

		err := metrics.RegisterCounter(config)
		if err != nil {
			t.Fatalf("expected no error on first registration, got %v", err)
		}

		err = metrics.RegisterCounter(config)
		if err == nil {
			t.Error("expected error on duplicate registration")
		}

		if !strings.Contains(err.Error(), "already exists") {
			t.Errorf("expected 'already exists' error, got %v", err)
		}
	})
}

func TestMetricsCollector_RegisterGauge(t *testing.T) {
	t.Parallel()

	metrics := NewMetricsCollector("test-service")

	config := MetricConfig{
		Name:   "test_gauge",
		Help:   "Test gauge metric",
		Labels: []string{"label1"},
	}

	err := metrics.RegisterGauge(config)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if _, exists := metrics.gauges["test-service_test_gauge"]; !exists {
		t.Error("expected gauge to be registered")
	}
}

func TestMetricsCollector_RegisterHistogram(t *testing.T) {
	t.Parallel()

	t.Run("registers histogram with default buckets", func(t *testing.T) {
		t.Parallel()

		metrics := NewMetricsCollector("test-service")

		config := MetricConfig{
			Name:   "test_histogram",
			Help:   "Test histogram metric",
			Labels: []string{"label1"},
		}

		err := metrics.RegisterHistogram(config)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if _, exists := metrics.histograms["test-service_test_histogram"]; !exists {
			t.Error("expected histogram to be registered")
		}
	})

	t.Run("registers histogram with custom buckets", func(t *testing.T) {
		t.Parallel()

		metrics := NewMetricsCollector("test-service")

		config := MetricConfig{
			Name:    "test_histogram_custom",
			Help:    "Test histogram with custom buckets",
			Labels:  []string{"label1"},
			Buckets: []float64{0.1, 0.5, 1.0, 5.0, 10.0},
		}

		err := metrics.RegisterHistogram(config)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if _, exists := metrics.histograms["test-service_test_histogram_custom"]; !exists {
			t.Error("expected histogram to be registered")
		}
	})
}

func TestMetricsCollector_RegisterSummary(t *testing.T) {
	t.Parallel()

	t.Run("registers summary with default objectives", func(t *testing.T) {
		t.Parallel()

		metrics := NewMetricsCollector("test-service")

		config := MetricConfig{
			Name:   "test_summary",
			Help:   "Test summary metric",
			Labels: []string{"label1"},
		}

		err := metrics.RegisterSummary(config)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if _, exists := metrics.summaries["test-service_test_summary"]; !exists {
			t.Error("expected summary to be registered")
		}
	})

	t.Run("registers summary with custom objectives", func(t *testing.T) {
		t.Parallel()

		metrics := NewMetricsCollector("test-service")

		config := MetricConfig{
			Name:       "test_summary_custom",
			Help:       "Test summary with custom objectives",
			Labels:     []string{"label1"},
			Objectives: map[float64]float64{0.5: 0.05, 0.95: 0.01},
		}

		err := metrics.RegisterSummary(config)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if _, exists := metrics.summaries["test-service_test_summary_custom"]; !exists {
			t.Error("expected summary to be registered")
		}
	})
}

func TestMetricsCollector_CounterOperations(t *testing.T) {
	t.Parallel()

	t.Run("increments counter", func(t *testing.T) {
		t.Parallel()

		metrics := NewMetricsCollector("test-service")

		// Register a counter
		config := MetricConfig{
			Name:   "test_counter_ops",
			Help:   "Test counter operations",
			Labels: []string{"operation"},
		}

		err := metrics.RegisterCounter(config)
		if err != nil {
			t.Fatalf("failed to register counter: %v", err)
		}

		err = metrics.IncCounter("test_counter_ops", "inc")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify the counter was incremented
		counter := metrics.counters["test-service_test_counter_ops"]
		metric := &dto.Metric{}

		err = counter.WithLabelValues("inc").Write(metric)
		if err != nil {
			t.Fatalf("failed to write metric: %v", err)
		}

		if metric.GetCounter().GetValue() != 1 {
			t.Errorf("expected counter value 1, got %f", metric.GetCounter().GetValue())
		}
	})

	t.Run("adds value to counter", func(t *testing.T) {
		t.Parallel()

		metrics := NewMetricsCollector("test-service")

		// Register a counter
		config := MetricConfig{
			Name:   "test_counter_ops",
			Help:   "Test counter operations",
			Labels: []string{"operation"},
		}

		err := metrics.RegisterCounter(config)
		if err != nil {
			t.Fatalf("failed to register counter: %v", err)
		}

		err = metrics.AddCounter("test_counter_ops", 5.5, "add")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify the counter was incremented by 5.5
		counter := metrics.counters["test-service_test_counter_ops"]
		metric := &dto.Metric{}

		err = counter.WithLabelValues("add").Write(metric)
		if err != nil {
			t.Fatalf("failed to write metric: %v", err)
		}

		if metric.GetCounter().GetValue() != 5.5 {
			t.Errorf("expected counter value 5.5, got %f", metric.GetCounter().GetValue())
		}
	})

	t.Run("fails with non-existent counter", func(t *testing.T) {
		t.Parallel()

		metrics := NewMetricsCollector("test-service")

		err := metrics.IncCounter("non_existent_counter", "test")
		if err == nil {
			t.Error("expected error for non-existent counter")
		}

		if !strings.Contains(err.Error(), "not found") {
			t.Errorf("expected 'not found' error, got %v", err)
		}
	})
}

//nolint:gocognit
func TestMetricsCollector_GaugeOperations(t *testing.T) {
	t.Parallel()

	t.Run("sets gauge value", func(t *testing.T) {
		t.Parallel()

		metrics := NewMetricsCollector("test-service")

		// Register a gauge
		config := MetricConfig{
			Name:   "test_gauge_ops",
			Help:   "Test gauge operations",
			Labels: []string{"operation"},
		}

		err := metrics.RegisterGauge(config)
		if err != nil {
			t.Fatalf("failed to register gauge: %v", err)
		}

		err = metrics.SetGauge("test_gauge_ops", 42.5, "set")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify the gauge value
		gauge := metrics.gauges["test-service_test_gauge_ops"]
		metric := &dto.Metric{}

		err = gauge.WithLabelValues("set").Write(metric)
		if err != nil {
			t.Fatalf("failed to write metric: %v", err)
		}

		if metric.GetGauge().GetValue() != 42.5 {
			t.Errorf("expected gauge value 42.5, got %f", metric.GetGauge().GetValue())
		}
	})

	t.Run("increments gauge", func(t *testing.T) {
		t.Parallel()

		metrics := NewMetricsCollector("test-service")

		// Register a gauge
		config := MetricConfig{
			Name:   "test_gauge_ops",
			Help:   "Test gauge operations",
			Labels: []string{"operation"},
		}

		err := metrics.RegisterGauge(config)
		if err != nil {
			t.Fatalf("failed to register gauge: %v", err)
		}

		// First set a value
		err = metrics.SetGauge("test_gauge_ops", 10, "inc")
		if err != nil {
			t.Fatalf("failed to set initial gauge value: %v", err)
		}

		err = metrics.IncGauge("test_gauge_ops", "inc")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify the gauge was incremented
		gauge := metrics.gauges["test-service_test_gauge_ops"]
		metric := &dto.Metric{}

		err = gauge.WithLabelValues("inc").Write(metric)
		if err != nil {
			t.Fatalf("failed to write metric: %v", err)
		}

		if metric.GetGauge().GetValue() != 11 {
			t.Errorf("expected gauge value 11, got %f", metric.GetGauge().GetValue())
		}
	})

	t.Run("decrements gauge", func(t *testing.T) {
		t.Parallel()

		metrics := NewMetricsCollector("test-service")

		// Register a gauge
		config := MetricConfig{
			Name:   "test_gauge_ops",
			Help:   "Test gauge operations",
			Labels: []string{"operation"},
		}

		err := metrics.RegisterGauge(config)
		if err != nil {
			t.Fatalf("failed to register gauge: %v", err)
		}

		// First set a value
		err = metrics.SetGauge("test_gauge_ops", 10, "dec")
		if err != nil {
			t.Fatalf("failed to set initial gauge value: %v", err)
		}

		err = metrics.DecGauge("test_gauge_ops", "dec")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify the gauge was decremented
		gauge := metrics.gauges["test-service_test_gauge_ops"]
		metric := &dto.Metric{}

		err = gauge.WithLabelValues("dec").Write(metric)
		if err != nil {
			t.Fatalf("failed to write metric: %v", err)
		}

		if metric.GetGauge().GetValue() != 9 {
			t.Errorf("expected gauge value 9, got %f", metric.GetGauge().GetValue())
		}
	})

	t.Run("adds value to gauge", func(t *testing.T) {
		t.Parallel()

		metrics := NewMetricsCollector("test-service")

		// Register a gauge
		config := MetricConfig{
			Name:   "test_gauge_ops",
			Help:   "Test gauge operations",
			Labels: []string{"operation"},
		}

		err := metrics.RegisterGauge(config)
		if err != nil {
			t.Fatalf("failed to register gauge: %v", err)
		}

		// First set a value
		err = metrics.SetGauge("test_gauge_ops", 10, "add")
		if err != nil {
			t.Fatalf("failed to set initial gauge value: %v", err)
		}

		err = metrics.AddGauge("test_gauge_ops", 5.5, "add")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify the gauge value
		gauge := metrics.gauges["test-service_test_gauge_ops"]
		metric := &dto.Metric{}

		err = gauge.WithLabelValues("add").Write(metric)
		if err != nil {
			t.Fatalf("failed to write metric: %v", err)
		}

		if metric.GetGauge().GetValue() != 15.5 {
			t.Errorf("expected gauge value 15.5, got %f", metric.GetGauge().GetValue())
		}
	})
}

func TestMetricsCollector_HistogramOperations(t *testing.T) {
	t.Parallel()

	t.Run("observes histogram values", func(t *testing.T) {
		t.Parallel()

		metrics := NewMetricsCollector("test-service")

		// Register a histogram
		config := MetricConfig{
			Name:   "test_histogram_ops",
			Help:   "Test histogram operations",
			Labels: []string{"operation"},
		}

		err := metrics.RegisterHistogram(config)
		if err != nil {
			t.Fatalf("failed to register histogram: %v", err)
		}

		values := []float64{0.1, 0.5, 1.0, 2.0, 5.0}
		for _, value := range values {
			err := metrics.ObserveHistogram("test_histogram_ops", value, "observe")
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		}

		// Verify the histogram recorded the observations by checking the registry
		registry := metrics.GetRegistry()

		metricFamilies, err := registry.Gather()
		if err != nil {
			t.Fatalf("failed to gather metrics: %v", err)
		}

		found := false

		for _, mf := range metricFamilies {
			if mf.GetName() == "test-service_test_histogram_ops" {
				found = true

				for _, metric := range mf.GetMetric() {
					if metric.GetHistogram().GetSampleCount() == uint64(len(values)) {
						return // Test passed
					}
				}
			}
		}

		if !found {
			t.Error("expected to find histogram metric")
		}
	})
}

func TestMetricsCollector_SummaryOperations(t *testing.T) {
	t.Parallel()

	t.Run("observes summary values", func(t *testing.T) {
		t.Parallel()

		metrics := NewMetricsCollector("test-service")

		// Register a summary
		config := MetricConfig{
			Name:   "test_summary_ops",
			Help:   "Test summary operations",
			Labels: []string{"operation"},
		}

		err := metrics.RegisterSummary(config)
		if err != nil {
			t.Fatalf("failed to register summary: %v", err)
		}

		values := []float64{0.1, 0.5, 1.0, 2.0, 5.0}
		for _, value := range values {
			err := metrics.ObserveSummary("test_summary_ops", value, "observe")
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		}

		// Verify the summary recorded the observations by checking the registry
		registry := metrics.GetRegistry()

		metricFamilies, err := registry.Gather()
		if err != nil {
			t.Fatalf("failed to gather metrics: %v", err)
		}

		found := false

		for _, mf := range metricFamilies {
			if mf.GetName() == "test-service_test_summary_ops" {
				found = true

				for _, metric := range mf.GetMetric() {
					if metric.GetSummary().GetSampleCount() == uint64(len(values)) {
						return // Test passed
					}
				}
			}
		}

		if !found {
			t.Error("expected to find summary metric")
		}
	})
}

func TestHelperFunctions(t *testing.T) {
	t.Parallel()

	svc := New("test-service", nil)

	// Register custom metrics
	err := svc.RegisterCounter(MetricConfig{
		Name:   "test_helper_counter",
		Help:   "Test helper counter",
		Labels: []string{"action"},
	})
	if err != nil {
		t.Fatalf("failed to register counter: %v", err)
	}

	err = svc.RegisterGauge(MetricConfig{
		Name:   "test_helper_gauge",
		Help:   "Test helper gauge",
		Labels: []string{"action"},
	})
	if err != nil {
		t.Fatalf("failed to register gauge: %v", err)
	}

	// Test helper functions with middleware
	handler := applyMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Test counter helpers
		err := IncCounter(r, "test_helper_counter", "increment")
		if err != nil {
			t.Errorf("IncCounter failed: %v", err)
		}

		err = AddCounter(r, "test_helper_counter", 5.0, "add")
		if err != nil {
			t.Errorf("AddCounter failed: %v", err)
		}

		// Test gauge helpers
		err = SetGauge(r, "test_helper_gauge", 42.0, "set")
		if err != nil {
			t.Errorf("SetGauge failed: %v", err)
		}

		err = IncGauge(r, "test_helper_gauge", "increment")
		if err != nil {
			t.Errorf("IncGauge failed: %v", err)
		}

		w.WriteHeader(http.StatusOK)
	}), MetricsMiddleware(svc.Metrics))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", recorder.Code)
	}
}

func TestHelperFunctionsWithoutMetrics(t *testing.T) {
	t.Parallel()

	// Test helper functions without metrics in context
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	err := IncCounter(req, "test_counter", "test")
	if err == nil {
		t.Error("expected error when metrics not available")
	}

	if !strings.Contains(err.Error(), "not available") {
		t.Errorf("expected 'not available' error, got %v", err)
	}
}

func TestMetricsMiddleware_CustomMetrics(t *testing.T) {
	t.Parallel()

	svc := New("test-service", nil)

	handler := applyMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate some processing time
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	}), MetricsMiddleware(svc.Metrics))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", recorder.Code)
	}

	// Verify metrics were recorded
	registry := svc.Metrics.GetRegistry()

	metricFamilies, err := registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}

	// Check that we have the expected metrics
	foundRequestsTotal := false
	foundRequestDuration := false
	foundRequestsInFlight := false

	for _, mf := range metricFamilies {
		switch mf.GetName() {
		case "test-service_http_requests_total":
			foundRequestsTotal = true
		case "test-service_http_request_duration_seconds":
			foundRequestDuration = true
		case "test-service_http_requests_in_flight":
			foundRequestsInFlight = true
		}
	}

	if !foundRequestsTotal {
		t.Error("expected to find http_requests_total metric")
	}

	if !foundRequestDuration {
		t.Error("expected to find http_request_duration_seconds metric")
	}

	if !foundRequestsInFlight {
		t.Error("expected to find http_requests_in_flight metric")
	}
}

func TestMetricsRegistry(t *testing.T) {
	t.Parallel()

	metrics := NewMetricsCollector("test-service")

	// Test that we can get the registry
	registry := metrics.GetRegistry()
	if registry == nil {
		t.Fatal("expected registry to be returned")
	}

	// Test that we can create a custom handler
	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	if handler == nil {
		t.Fatal("expected handler to be created")
	}

	// Test that the handler works
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", recorder.Code)
	}

	// Check that the response contains metrics
	body := recorder.Body.String()
	if !strings.Contains(body, "test_service_http_requests_in_flight") {
		t.Errorf("expected metrics output to contain http_requests_in_flight, got: %s", body)
	}
}

func TestService_RegisterMetrics(t *testing.T) {
	t.Parallel()

	t.Run("registers counter via service", func(t *testing.T) {
		t.Parallel()

		svc := New("test-service", nil)

		err := svc.RegisterCounter(MetricConfig{
			Name:   "service_test_counter",
			Help:   "Test counter via service",
			Labels: []string{"label1"},
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if _, exists := svc.Metrics.counters["test-service_service_test_counter"]; !exists {
			t.Error("expected counter to be registered")
		}
	})

	t.Run("registers gauge via service", func(t *testing.T) {
		t.Parallel()

		svc := New("test-service", nil)

		err := svc.RegisterGauge(MetricConfig{
			Name:   "service_test_gauge",
			Help:   "Test gauge via service",
			Labels: []string{"label1"},
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if _, exists := svc.Metrics.gauges["test-service_service_test_gauge"]; !exists {
			t.Error("expected gauge to be registered")
		}
	})

	t.Run("registers histogram via service", func(t *testing.T) {
		t.Parallel()

		svc := New("test-service", nil)

		err := svc.RegisterHistogram(MetricConfig{
			Name:   "service_test_histogram",
			Help:   "Test histogram via service",
			Labels: []string{"label1"},
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if _, exists := svc.Metrics.histograms["test-service_service_test_histogram"]; !exists {
			t.Error("expected histogram to be registered")
		}
	})

	t.Run("registers summary via service", func(t *testing.T) {
		t.Parallel()

		svc := New("test-service", nil)

		err := svc.RegisterSummary(MetricConfig{
			Name:   "service_test_summary",
			Help:   "Test summary via service",
			Labels: []string{"label1"},
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if _, exists := svc.Metrics.summaries["test-service_service_test_summary"]; !exists {
			t.Error("expected summary to be registered")
		}
	})
}
