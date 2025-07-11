package service

import (
	"context"
	"net/http"
	"time"

	"github.com/hellofresh/health-go/v5"
)

// HealthChecker wraps the health-go library health checker
type HealthChecker struct {
	checker *health.Health
}

// NewHealthChecker creates a new health checker with the service component information
func NewHealthChecker(serviceName, version string) (*HealthChecker, error) {
	checker, err := health.New(
		health.WithComponent(health.Component{
			Name:    serviceName,
			Version: version,
		}),
	)
	if err != nil {
		return nil, err
	}

	return &HealthChecker{
		checker: checker,
	}, nil
}

// Register adds a health check to the health checker
func (hc *HealthChecker) Register(config health.Config) {
	hc.checker.Register(config)
}

// Handler returns the HTTP handler for health checks
func (hc *HealthChecker) Handler() http.Handler {
	return hc.checker.Handler()
}

// HandlerFunc returns the HTTP handler function for health checks
func (hc *HealthChecker) HandlerFunc(w http.ResponseWriter, r *http.Request) {
	hc.checker.HandlerFunc(w, r)
}

// Measure returns the current health status
func (hc *HealthChecker) Measure(ctx context.Context) health.Check {
	return hc.checker.Measure(ctx)
}

// IsHealthy returns true if all health checks are passing
func (hc *HealthChecker) IsHealthy(ctx context.Context) bool {
	check := hc.Measure(ctx)
	return check.Status == health.StatusOK
}

// IsReady returns true if the service is ready to serve requests
// This is typically used for Kubernetes readiness probes
func (hc *HealthChecker) IsReady(ctx context.Context) bool {
	// For readiness, we want to check if critical services are available
	// This is the same as health check for now, but can be customized
	return hc.IsHealthy(ctx)
}

// IsAlive returns true if the service is alive
// This is typically used for Kubernetes liveness probes
func (hc *HealthChecker) IsAlive(ctx context.Context) bool {
	// For liveness, we want to check if the service is still running
	// This should be more lenient than health checks
	// For now, we'll just return true as the service is running if this is called
	return true
}

// ReadinessHandler returns an HTTP handler for readiness checks
func (hc *HealthChecker) ReadinessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		if hc.IsReady(ctx) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Ready"))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Not Ready"))
		}
	}
}

// LivenessHandler returns an HTTP handler for liveness checks
func (hc *HealthChecker) LivenessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		if hc.IsAlive(ctx) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Alive"))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Not Alive"))
		}
	}
}

// GetHealthChecker retrieves the health checker from the request context
func GetHealthChecker(r *http.Request) *HealthChecker {
	hc, ok := r.Context().Value(HealthCheckerKey).(*HealthChecker)
	if !ok {
		return nil
	}
	return hc
}
