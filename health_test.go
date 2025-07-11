package service

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hellofresh/health-go/v5"
)

func TestNewHealthChecker(t *testing.T) {
	t.Parallel()

	t.Run("creates health checker successfully", func(t *testing.T) {
		t.Parallel()

		healthChecker, err := NewHealthChecker("test-service", "v1.0.0")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if healthChecker == nil {
			t.Fatal("expected health checker to be created")
		}

		if healthChecker.checker == nil {
			t.Fatal("expected internal health checker to be initialized")
		}
	})
}

func TestHealthChecker_Register(t *testing.T) {
	t.Parallel()

	healthChecker, err := NewHealthChecker("test-service", "v1.0.0")
	if err != nil {
		t.Fatalf("failed to create health checker: %v", err)
	}

	// Register a simple health check
	healthChecker.Register(health.Config{
		Name: "test-check",
		Check: func(ctx context.Context) error {
			return nil
		},
	})

	// Verify health check was registered by measuring health
	check := healthChecker.Measure(context.Background())

	if check.Status != health.StatusOK {
		t.Errorf("expected status OK, got %s", check.Status)
	}
}

func TestHealthChecker_IsHealthy(t *testing.T) {
	t.Parallel()

	healthChecker, err := NewHealthChecker("test-service", "v1.0.0")
	if err != nil {
		t.Fatalf("failed to create health checker: %v", err)
	}

	t.Run("returns true when all checks pass", func(t *testing.T) {
		t.Parallel()

		healthChecker.Register(health.Config{
			Name: "passing-check",
			Check: func(_ context.Context) error {
				return nil
			},
		})

		if !healthChecker.IsHealthy(context.Background()) {
			t.Error("expected IsHealthy to return true")
		}
	})

	t.Run("returns false when check fails", func(t *testing.T) {
		t.Parallel()

		healthChecker.Register(health.Config{
			Name: "failing-check",
			Check: func(ctx context.Context) error {
				return errors.New("check failed") //nolint:err113
			},
		})

		if healthChecker.IsHealthy(context.Background()) {
			t.Error("expected IsHealthy to return false")
		}
	})
}

func TestHealthChecker_IsReady(t *testing.T) {
	t.Parallel()

	healthChecker, err := NewHealthChecker("test-service", "v1.0.0")
	if err != nil {
		t.Fatalf("failed to create health checker: %v", err)
	}

	t.Run("returns true when healthy", func(t *testing.T) {
		t.Parallel()

		healthChecker.Register(health.Config{
			Name: "ready-check",
			Check: func(ctx context.Context) error {
				return nil
			},
		})

		if !healthChecker.IsReady(context.Background()) {
			t.Error("expected IsReady to return true")
		}
	})
}

func TestHealthChecker_IsAlive(t *testing.T) {
	t.Parallel()

	healthChecker, err := NewHealthChecker("test-service", "v1.0.0")
	if err != nil {
		t.Fatalf("failed to create health checker: %v", err)
	}

	// IsAlive should always return true for a running service
	if !healthChecker.IsAlive(context.Background()) {
		t.Error("expected IsAlive to return true")
	}
}

func TestHealthChecker_Handlers(t *testing.T) {
	t.Parallel()

	healthChecker, err := NewHealthChecker("test-service", "v1.0.0")
	if err != nil {
		t.Fatalf("failed to create health checker: %v", err)
	}

	// Register a health check
	healthChecker.Register(health.Config{
		Name: "test-check",
		Check: func(ctx context.Context) error {
			return nil
		},
	})

	t.Run("Handler returns 200 for healthy service", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		recorder := httptest.NewRecorder()

		healthChecker.Handler().ServeHTTP(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", recorder.Code)
		}
	})

	t.Run("HandlerFunc returns 200 for healthy service", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		recorder := httptest.NewRecorder()

		healthChecker.HandlerFunc(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", recorder.Code)
		}
	})

	t.Run("ReadinessHandler returns 200 for ready service", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/ready", nil)
		recorder := httptest.NewRecorder()

		healthChecker.ReadinessHandler()(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", recorder.Code)
		}

		if recorder.Body.String() != "Ready" {
			t.Errorf("expected body 'Ready', got %s", recorder.Body.String())
		}
	})

	t.Run("LivenessHandler returns 200 for alive service", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/live", nil)
		recorder := httptest.NewRecorder()

		healthChecker.LivenessHandler()(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", recorder.Code)
		}

		if recorder.Body.String() != "Alive" {
			t.Errorf("expected body 'Alive', got %s", recorder.Body.String())
		}
	})
}

func TestHealthChecker_HandlersWithFailures(t *testing.T) {
	t.Parallel()

	healthChecker, err := NewHealthChecker("test-service", "v1.0.0")
	if err != nil {
		t.Fatalf("failed to create health checker: %v", err)
	}

	// Register a failing health check
	healthChecker.Register(health.Config{
		Name: "failing-check",
		Check: func(ctx context.Context) error {
			return errors.New("check failed") //nolint:err113
		},
	})

	t.Run("ReadinessHandler returns 503 for not ready service", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/ready", nil)
		recorder := httptest.NewRecorder()

		healthChecker.ReadinessHandler()(recorder, req)

		if recorder.Code != http.StatusServiceUnavailable {
			t.Errorf("expected status 503, got %d", recorder.Code)
		}

		if recorder.Body.String() != "Not Ready" {
			t.Errorf("expected body 'Not Ready', got %s", recorder.Body.String())
		}
	})
}

func TestGetHealthChecker(t *testing.T) {
	t.Parallel()

	healthChecker, err := NewHealthChecker("test-service", "v1.0.0")
	if err != nil {
		t.Fatalf("failed to create health checker: %v", err)
	}

	t.Run("returns health checker from context", func(t *testing.T) {
		t.Parallel()

		handler := HealthCheckerMiddleware(healthChecker)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			retrievedHC := GetHealthChecker(r)
			if retrievedHC == nil {
				t.Error("expected health checker to be retrieved from context")
			}

			if retrievedHC != healthChecker {
				t.Error("expected retrieved health checker to match original")
			}

			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		recorder := httptest.NewRecorder()

		handler.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", recorder.Code)
		}
	})

	t.Run("returns nil when not in context", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/test", nil)

		retrievedHC := GetHealthChecker(req)
		if retrievedHC != nil {
			t.Error("expected health checker to be nil when not in context")
		}
	})
}

func TestHealthCheckerMiddleware(t *testing.T) {
	t.Parallel()

	healthChecker, err := NewHealthChecker("test-service", "v1.0.0")
	if err != nil {
		t.Fatalf("failed to create health checker: %v", err)
	}

	middleware := HealthCheckerMiddleware(healthChecker)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify health checker is available in context
		retrievedHC := GetHealthChecker(r)
		if retrievedHC == nil {
			t.Error("health checker should be available in context")
		}

		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", recorder.Code)
	}
}

func TestHealthChecker_Timeout(t *testing.T) {
	t.Parallel()

	healthChecker, err := NewHealthChecker("test-service", "v1.0.0")
	if err != nil {
		t.Fatalf("failed to create health checker: %v", err)
	}

	// Register a health check with a timeout
	healthChecker.Register(health.Config{
		Name:    "slow-check",
		Timeout: 100 * time.Millisecond,
		Check: func(ctx context.Context) error {
			// Simulate a slow operation
			time.Sleep(200 * time.Millisecond)
			return nil
		},
	})

	// This should timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	if healthChecker.IsHealthy(ctx) {
		t.Error("expected health check to fail due to timeout")
	}
}

func TestService_RegisterHealthCheck(t *testing.T) {
	t.Parallel()

	svc := New("test-service", nil)

	t.Run("registers health check successfully", func(t *testing.T) {
		t.Parallel()

		checkCalled := false

		svc.RegisterHealthCheck(health.Config{
			Name: "test-check",
			Check: func(ctx context.Context) error {
				checkCalled = true
				return nil
			},
		})

		// Verify the check was registered by measuring health
		if svc.HealthChecker != nil {
			check := svc.HealthChecker.Measure(context.Background())

			if check.Status != health.StatusOK {
				t.Errorf("expected status OK, got %s", check.Status)
			}

			if !checkCalled {
				t.Error("expected health check to be called")
			}
		}
	})

	t.Run("handles nil health checker gracefully", func(t *testing.T) {
		t.Parallel()

		svcWithoutHealth := &Service{
			Name:          "test",
			Logger:        svc.Logger,
			HealthChecker: nil,
		}

		// This should not panic
		svcWithoutHealth.RegisterHealthCheck(health.Config{
			Name: "test-check",
			Check: func(ctx context.Context) error {
				return nil
			},
		})
	})
}

func TestService_GetHealthChecker(t *testing.T) {
	t.Parallel()

	svc := New("test-service", nil)

	healthChecker := svc.GetHealthChecker()
	if healthChecker == nil {
		t.Error("expected health checker to be available")
	}

	if healthChecker != svc.HealthChecker {
		t.Error("expected returned health checker to match service health checker")
	}
}
