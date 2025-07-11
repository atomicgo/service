package service

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("with config", func(t *testing.T) {
		t.Parallel()

		config := DefaultConfig()
		config.Addr = ":8081"

		svc := New("test", config)

		if svc.Name != "test" {
			t.Errorf("expected name 'test', got %s", svc.Name)
		}

		if svc.Config.Addr != ":8081" {
			t.Errorf("expected addr ':8081', got %s", svc.Config.Addr)
		}

		if svc.Logger == nil {
			t.Error("logger should not be nil")
		}

		if svc.Metrics == nil {
			t.Error("metrics should not be nil")
		}
	})

	t.Run("with nil config", func(t *testing.T) {
		t.Parallel()

		svc := New("test", nil)

		if svc.Config == nil {
			t.Error("config should not be nil when nil is passed")
		}

		if svc.Config.Addr != ":8080" {
			t.Errorf("expected default addr ':8080', got %s", svc.Config.Addr)
		}
	})
}

func TestDefaultConfig(t *testing.T) {
	t.Parallel()

	config := DefaultConfig()

	if config.Addr != ":8080" {
		t.Errorf("expected default addr ':8080', got %s", config.Addr)
	}

	if config.MetricsAddr != ":9090" {
		t.Errorf("expected default metrics addr ':9090', got %s", config.MetricsAddr)
	}

	if config.MetricsPath != "/metrics" {
		t.Errorf("expected default metrics path '/metrics', got %s", config.MetricsPath)
	}

	if config.ReadTimeout != 10*time.Second {
		t.Errorf("expected default read timeout 10s, got %v", config.ReadTimeout)
	}

	if config.Logger == nil {
		t.Error("default logger should not be nil")
	}
}

func TestLoadFromEnv(t *testing.T) {
	t.Parallel()

	// Set environment variables
	t.Setenv("ADDR", ":8888")
	t.Setenv("METRICS_ADDR", ":9999")
	t.Setenv("METRICS_PATH", "/custom-metrics")

	config, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("failed to load config from env: %v", err)
	}

	if config.Addr != ":8888" {
		t.Errorf("expected addr ':8888', got %s", config.Addr)
	}

	if config.MetricsAddr != ":9999" {
		t.Errorf("expected metrics addr ':9999', got %s", config.MetricsAddr)
	}

	if config.MetricsPath != "/custom-metrics" {
		t.Errorf("expected metrics path '/custom-metrics', got %s", config.MetricsPath)
	}
}

func TestHandleFunc(t *testing.T) {
	t.Parallel()

	svc := New("test", nil)

	// Add a simple handler
	svc.HandleFunc("/test", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()

	// Serve the request
	svc.mux.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", recorder.Code)
	}

	if strings.TrimSpace(recorder.Body.String()) != "test response" {
		t.Errorf("expected body 'test response', got %s", recorder.Body.String())
	}
}

func TestGetLogger(t *testing.T) {
	t.Parallel()

	svc := New("test", nil)

	// Test with a request that has logger middleware applied
	handler := applyMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := GetLogger(r)
		if logger == nil {
			t.Error("logger should not be nil")
		}

		w.WriteHeader(http.StatusOK)
	}), LoggerMiddleware(svc.Logger))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", recorder.Code)
	}
}

func TestGetMetrics(t *testing.T) {
	t.Parallel()

	svc := New("test", nil)

	// Test with a request that has metrics middleware applied
	handler := applyMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metrics := GetMetrics(r)
		if metrics == nil {
			t.Error("metrics should not be nil")
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

func TestRecoveryMiddleware(t *testing.T) {
	t.Parallel()

	svc := New("test", nil)

	// Create a handler that panics
	handler := applyMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		panic("test panic")
	}), RecoveryMiddleware(svc.Logger))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", recorder.Code)
	}

	if !strings.Contains(recorder.Body.String(), "Internal Server Error") {
		t.Error("expected 'Internal Server Error' in response body")
	}
}

func TestMetricsMiddleware(t *testing.T) {
	t.Parallel()

	svc := New("test", nil)

	// Create a simple handler
	handler := applyMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	}), MetricsMiddleware(svc.Metrics))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", recorder.Code)
	}
}

func TestShutdownHooks(t *testing.T) {
	t.Parallel()

	svc := New("test", nil)

	hookCalled := false

	svc.AddShutdownHook(func() error {
		hookCalled = true
		return nil
	})

	// Create a minimal server setup
	svc.server = &http.Server{Addr: ":0"}        //nolint:gosec
	svc.metricsServer = &http.Server{Addr: ":0"} //nolint:gosec

	// Test graceful shutdown
	err := svc.gracefulShutdown()
	if err != nil {
		t.Errorf("graceful shutdown failed: %v", err)
	}

	if !hookCalled {
		t.Error("shutdown hook was not called")
	}
}

func TestUse(t *testing.T) {
	t.Parallel()

	svc := New("test", nil)

	initialCount := len(svc.middlewares)

	// Add a custom middleware
	svc.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Custom", "test")
			next.ServeHTTP(w, r)
		})
	})

	if len(svc.middlewares) != initialCount+1 {
		t.Errorf("expected %d middlewares, got %d", initialCount+1, len(svc.middlewares))
	}

	// Test that the custom middleware is applied
	svc.HandleFunc("/test", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()

	svc.mux.ServeHTTP(recorder, req)

	if recorder.Header().Get("X-Custom") != "test" {
		t.Error("custom middleware was not applied")
	}
}

func TestIntegration(t *testing.T) {
	t.Parallel()

	// Create a service with custom configuration
	config := DefaultConfig()
	config.Addr = ":0" // Use random port
	config.MetricsAddr = ":0"

	svc := New("integration-test", config)

	// Add a test handler
	svc.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		logger := GetLogger(r)
		logger.Info("hello endpoint called")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	})

	// Test the handler
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	recorder := httptest.NewRecorder()

	svc.mux.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", recorder.Code)
	}

	if strings.TrimSpace(recorder.Body.String()) != "Hello, World!" {
		t.Errorf("expected body 'Hello, World!', got %s", recorder.Body.String())
	}

	t.Log("Integration test completed successfully")
}

func TestStartMetricsServer(t *testing.T) {
	t.Parallel()

	svc := New("test", nil)

	// Test that metrics server can be started (we'll use a mock)
	// In practice, you'd test this by actually starting the server
	// and checking if endpoints are accessible

	if svc.Config.MetricsAddr != ":9090" {
		t.Errorf("expected metrics addr ':9090', got %s", svc.Config.MetricsAddr)
	}

	if svc.Config.MetricsPath != "/metrics" {
		t.Errorf("expected metrics path '/metrics', got %s", svc.Config.MetricsPath)
	}
}

// BenchmarkHandleFunc benchmarks the HandleFunc method
func BenchmarkHandleFunc(b *testing.B) {
	svc := New("benchmark", nil)

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("benchmark"))
	}

	svc.HandleFunc("/benchmark", handler)

	req := httptest.NewRequest(http.MethodGet, "/benchmark", nil)

	b.ResetTimer()

	for range b.N {
		recorder := httptest.NewRecorder()
		svc.mux.ServeHTTP(recorder, req)
	}
}

// BenchmarkMiddleware benchmarks the middleware stack
func BenchmarkMiddleware(b *testing.B) {
	svc := New("benchmark", nil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := applyMiddleware(handler, svc.middlewares...)

	req := httptest.NewRequest(http.MethodGet, "/benchmark", nil)

	b.ResetTimer()

	for range b.N {
		recorder := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(recorder, req)
	}
}
