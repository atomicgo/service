/*
Package service provides a lightweight, production-ready HTTP service framework for Go applications.

The service framework is designed to be Kubernetes-ready and follows best practices for
highly available microservices. It includes built-in support for graceful shutdown,
Prometheus metrics, structured logging, middleware, and environment-based configuration.

## Features

- **HTTP Server**: Configurable HTTP server with timeouts and graceful shutdown
- **Metrics**: Built-in Prometheus metrics collection with automatic request tracking
- **Logging**: Structured logging with slog integration and context-aware loggers
- **Middleware**: Extensible middleware system with built-in recovery and logging
- **Configuration**: Environment-based configuration with sensible defaults
- **Graceful Shutdown**: Signal handling with configurable shutdown hooks
- **Health Checks**: Built-in health check endpoints
- **Kubernetes Ready**: Designed for containerized deployments

## Quick Start

```go
package main

import (

	"log/slog"
	"net/http"
	"os"

	"atomicgo.dev/service"

)

	func main() {
	    // Create service with default configuration
	    svc := service.New("my-service", nil)

	    // Register handlers
	    svc.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	        logger := service.GetLogger(r)
	        logger.Info("Hello, World!")
	        w.Write([]byte("Hello, World!"))
	    })

	    // Start with graceful shutdown
	    if err := svc.StartWithGracefulShutdown(); err != nil {
	        os.Exit(1)
	    }
	}

```

## Configuration

The framework supports configuration via environment variables with sensible defaults:

- `ADDR`: HTTP server address (default: ":8080")
- `METRICS_ADDR`: Metrics server address (default: ":9090")
- `METRICS_PATH`: Metrics endpoint path (default: "/metrics")
- `READ_TIMEOUT`: HTTP read timeout (default: "10s")
- `WRITE_TIMEOUT`: HTTP write timeout (default: "10s")
- `IDLE_TIMEOUT`: HTTP idle timeout (default: "120s")
- `SHUTDOWN_TIMEOUT`: Graceful shutdown timeout (default: "30s")

```go
// Load configuration from environment
config, err := service.LoadFromEnv()

	if err != nil {
	    log.Fatal(err)
	}

// Create service with custom configuration
svc := service.New("my-service", config)
```

## Middleware

The framework includes several built-in middleware:

- **LoggerMiddleware**: Injects logger into request context
- **RecoveryMiddleware**: Recovers from panics and logs errors
- **RequestLoggingMiddleware**: Logs incoming requests
- **MetricsMiddleware**: Tracks HTTP metrics for Prometheus

```go
// Add custom middleware

	svc.Use(func(next http.Handler) http.Handler {
	    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	        w.Header().Set("X-Custom", "value")
	        next.ServeHTTP(w, r)
	    })
	})

```

## Metrics

The framework automatically collects Prometheus metrics:

- `{service_name}_http_requests_total`: Total HTTP requests
- `{service_name}_http_request_duration_seconds`: Request duration
- `{service_name}_http_requests_in_flight`: In-flight requests

Metrics are available at `:9090/metrics` by default.

```go
// Access metrics in handlers

	func myHandler(w http.ResponseWriter, r *http.Request) {
	    metrics := service.GetMetrics(r)
	    if metrics != nil {
	        // Custom metric operations can be added here
	    }
	}

```

## Graceful Shutdown

The framework supports graceful shutdown with signal handling and custom hooks:

```go
// Add shutdown hooks

	svc.AddShutdownHook(func() error {
	    // Cleanup resources
	    return nil
	})

// Start with graceful shutdown
svc.StartWithGracefulShutdown()
```

## Logging

The framework uses structured logging with slog and provides context-aware loggers:

```go

	func myHandler(w http.ResponseWriter, r *http.Request) {
	    logger := service.GetLogger(r)
	    logger.Info("request processed", "path", r.URL.Path)
	}

```

## Health Checks

Health check endpoints are automatically available:

- `:9090/health`: Basic health check
- `:9090/metrics`: Prometheus metrics

## Kubernetes Deployment

The framework is designed for Kubernetes deployments with:

- Graceful shutdown handling SIGTERM
- Health check endpoints for liveness/readiness probes
- Prometheus metrics for monitoring
- Configurable resource limits via environment variables

## Examples

See the `_example/` directory for complete working examples demonstrating:

- Basic service setup
- Custom middleware
- Environment configuration
- Graceful shutdown
- Metrics integration

The framework is designed to be lightweight while providing all essential features
for production-ready microservices.
*/
package service
