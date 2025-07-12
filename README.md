<h1 align="center">AtomicGo | service</h1>

<p align="center">
<img src="https://img.shields.io/endpoint?url=https%3A%2F%2Fatomicgo.dev%2Fapi%2Fshields%2Fservice&style=flat-square" alt="Downloads">

<a href="https://github.com/atomicgo/service/releases">
<img src="https://img.shields.io/github/v/release/atomicgo/service?style=flat-square" alt="Latest Release">
</a>

<a href="https://codecov.io/gh/atomicgo/service" target="_blank">
<img src="https://img.shields.io/github/actions/workflow/status/atomicgo/service/go.yml?style=flat-square" alt="Tests">
</a>

<a href="https://codecov.io/gh/atomicgo/service" target="_blank">
<img src="https://img.shields.io/codecov/c/gh/atomicgo/service?color=magenta&logo=codecov&style=flat-square" alt="Coverage">
</a>

<a href="https://codecov.io/gh/atomicgo/service">
<!-- unittestcount:start --><img src="https://img.shields.io/badge/Unit_Tests-14-magenta?style=flat-square" alt="Unit test count"><!-- unittestcount:end -->
</a>

<a href="https://opensource.org/licenses/MIT" target="_blank">
<img src="https://img.shields.io/badge/License-MIT-yellow.svg?style=flat-square" alt="License: MIT">
</a>
  
<a href="https://goreportcard.com/report/github.com/atomicgo/service" target="_blank">
<img src="https://goreportcard.com/badge/github.com/atomicgo/service?style=flat-square" alt="Go report">
</a>   

</p>

---

<p align="center">
<strong><a href="https://pkg.go.dev/atomicgo.dev/service#section-documentation" target="_blank">Documentation</a></strong>
|
<strong><a href="https://github.com/atomicgo/atomicgo/blob/main/CONTRIBUTING.md" target="_blank">Contributing</a></strong>
|
<strong><a href="https://github.com/atomicgo/atomicgo/blob/main/CODE_OF_CONDUCT.md" target="_blank">Code of Conduct</a></strong>
</p>

---

<p align="center">
  <img src="https://raw.githubusercontent.com/atomicgo/atomicgo/main/assets/header.png" alt="AtomicGo">
</p>

<p align="center">
<table>
<tbody>
</tbody>
</table>
</p>
<h3  align="center"><pre>go get atomicgo.dev/service</pre></h3>
<p align="center">
<table>
<tbody>
</tbody>
</table>
</p>

---

A minimal boilerplate wrapper for building production-ready Go HTTP services. This library reduces the boilerplate of writing production/enterprise-grade Go services to a minimum.

**What this library provides:**
- Essential production features out of the box (metrics, health checks, graceful shutdown)
- Kubernetes and containerization boilerplate
- Lightweight wrapper around http.Server for high availability services

**What this library does NOT provide:**
- HTTP framework or routing
- Business logic or application patterns
- Restrictions on how you write your HTTP handlers
- Opinionated application architecture

Write HTTP handlers exactly as you prefer, using any patterns or frameworks you choose. This library handles the operational concerns while staying out of your application logic.

## Features

- **Minimal Boilerplate**: Reduces production service setup to a few lines of code
- **HTTP Server Wrapper**: Lightweight wrapper around http.Server with production defaults
- **Metrics**: Built-in Prometheus metrics collection with automatic request tracking
- **Logging**: Structured logging with slog integration and context-aware loggers
- **Middleware**: Extensible middleware system with built-in recovery and logging
- **Configuration**: Environment-based configuration with sensible defaults
- **Graceful Shutdown**: Signal handling with configurable shutdown hooks
- **Health Checks**: Built-in health check endpoints for Kubernetes
- **Framework Agnostic**: Works with any HTTP patterns or frameworks you prefer (as long as the framework supports the standard `http` package)

## Quick Start

Minimal boilerplate to get a production-ready service with metrics, health checks, and graceful shutdown:

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

    // Write HTTP handlers exactly as you prefer
    svc.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        logger := service.GetLogger(r) // Easy access to the logger
        logger.Info("Hello, World!")
        w.Write([]byte("Hello, World!"))
    })

    // Start service (includes graceful shutdown, metrics, health checks)
    if err := svc.Start(); err != nil {
        os.Exit(1)
    }
}
```

That's it! Your service now has:
- Prometheus metrics at `:9090/metrics`
- Comprehensive health checks at `:9090/health`
- Kubernetes readiness probe at `:9090/ready`
- Kubernetes liveness probe at `:9090/live`
- Graceful shutdown handling
- Structured logging
- Kubernetes-ready configuration

## Configuration

The framework supports configuration via environment variables with sensible defaults:

| Variable | Default | Description |
|----------|---------|-------------|
| `ADDR` | `:8080` | HTTP server address |
| `METRICS_ADDR` | `:9090` | Metrics server address |
| `METRICS_PATH` | `/metrics` | Metrics endpoint path |
| `HEALTH_PATH` | `/health` | Health check endpoint path |
| `READINESS_PATH` | `/ready` | Readiness probe endpoint path |
| `LIVENESS_PATH` | `/live` | Liveness probe endpoint path |
| `SERVICE_VERSION` | `v1.0.0` | Service version for health checks |
| `READ_TIMEOUT` | `10s` | HTTP read timeout |
| `WRITE_TIMEOUT` | `10s` | HTTP write timeout |
| `IDLE_TIMEOUT` | `120s` | HTTP idle timeout |
| `SHUTDOWN_TIMEOUT` | `30s` | Graceful shutdown timeout |

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

The framework provides a flexible metrics system with built-in HTTP metrics and support for custom metrics.

### Built-in HTTP Metrics

The framework automatically collects these Prometheus metrics for every service:

- `{service_name}_http_requests_total`: Total HTTP requests by method, endpoint, and status
- `{service_name}_http_request_duration_seconds`: Request duration histogram by method, endpoint, and status
- `{service_name}_http_requests_in_flight`: Current number of in-flight requests

These metrics are provided automatically without any configuration required.

### Custom Metrics

You can easily register and use custom metrics in your service:

```go
func main() {
    svc := service.New("my-service", nil)
    
    // Register custom metrics
    svc.RegisterCounter(service.MetricConfig{
        Name:   "user_registrations_total",
        Help:   "Total number of user registrations",
        Labels: []string{"source", "status"},
    })
    
    svc.RegisterGauge(service.MetricConfig{
        Name:   "active_users",
        Help:   "Number of currently active users",
        Labels: []string{"user_type"},
    })
    
    svc.RegisterHistogram(service.MetricConfig{
        Name:   "request_processing_duration_seconds",
        Help:   "Time spent processing requests",
        Labels: []string{"operation"},
        Buckets: []float64{0.001, 0.01, 0.1, 1.0, 10.0},
    })
    
    svc.RegisterSummary(service.MetricConfig{
        Name:   "response_size_bytes",
        Help:   "Size of responses in bytes",
        Labels: []string{"endpoint"},
        Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
    })
    
    // Use metrics in handlers
    svc.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
        // Increment counter
        service.IncCounter(r, "user_registrations_total", "web", "success")
        
        // Set gauge value
        service.SetGauge(r, "active_users", 42.0, "premium")
        
        // Observe histogram
        service.ObserveHistogram(r, "request_processing_duration_seconds", 0.25, "registration")
        
        // Observe summary
        service.ObserveSummary(r, "response_size_bytes", 1024.0, "/register")
        
        w.Write([]byte("User registered"))
    })
    
    svc.Start()
}
```

### Metric Types

The framework supports all standard Prometheus metric types:

- **Counter**: Monotonically increasing values (e.g., total requests, errors)
- **Gauge**: Values that can go up and down (e.g., active connections, memory usage)
- **Histogram**: Observations in configurable buckets (e.g., request duration, response size)
- **Summary**: Observations with configurable quantiles (e.g., request latency percentiles)

### Direct Access

For advanced use cases, you can access the metrics collector directly:

```go
func myHandler(w http.ResponseWriter, r *http.Request) {
    metrics := service.GetMetrics(r)
    if metrics != nil {
        // Direct access to metrics collector
        metrics.IncCounter("my_counter", "label_value")
        metrics.SetGauge("my_gauge", 42.0, "label_value")
        
        // Access the underlying Prometheus registry
        registry := metrics.GetRegistry()
        // Use registry for custom integrations
    }
}
```

All metrics are available at `:9090/metrics` by default.

## Graceful Shutdown

The framework includes graceful shutdown by default with signal handling and custom hooks:

```go
// Add shutdown hooks
svc.AddShutdownHook(func() error {
    // Cleanup resources
    return nil
})

// Start service (includes graceful shutdown)
svc.Start()
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

The framework integrates with [HelloFresh's health-go library](https://github.com/hellofresh/health-go) to provide comprehensive health checking capabilities.

### Built-in Health Endpoints

Health check endpoints are automatically available on the metrics server:

- `:9090/health`: Comprehensive health check with detailed status information
- `:9090/ready`: Kubernetes readiness probe endpoint
- `:9090/live`: Kubernetes liveness probe endpoint
- `:9090/metrics`: Prometheus metrics

### Adding Custom Health Checks

You can register custom health checks using the `RegisterHealthCheck` method:

```go
func main() {
    svc := service.New("my-service", nil)

    // Register a database health check
    svc.RegisterHealthCheck(health.Config{
        Name:      "database",
        Timeout:   time.Second * 5,
        SkipOnErr: false, // This check is critical
        Check: func(ctx context.Context) error {
            // Your database health check logic here
            return db.PingContext(ctx)
        },
    })

    // Register a Redis health check
    svc.RegisterHealthCheck(health.Config{
        Name:      "redis",
        Timeout:   time.Second * 3,
        SkipOnErr: true, // This check is optional
        Check: func(ctx context.Context) error {
            // Your Redis health check logic here
            return redisClient.Ping(ctx).Err()
        },
    })

    svc.Start()
}
```

### Using Built-in Health Checkers

The health-go library provides several built-in health checkers for common services:

```go
import (
    "github.com/hellofresh/health-go/v5"
    healthPostgres "github.com/hellofresh/health-go/v5/checks/postgres"
    healthRedis "github.com/hellofresh/health-go/v5/checks/redis"
)

func main() {
    svc := service.New("my-service", nil)

    // MySQL health check
	svc.RegisterHealthCheck(health.Config{
		Name:      "postgresql",
		Timeout:   time.Second * 5,
		SkipOnErr: false, // Critical check - service is unhealthy if DB is down
		Check: healthPostgres.New(healthPostgres.Config{
			DSN: dbURL,
		}),
	})

    // Redis health check
    svc.RegisterHealthCheck(health.Config{
        Name:    "redis",
        Timeout: time.Second * 3,
        Check: healthRedis.New(healthRedis.Config{
            Addr: "localhost:6379",
        }),
    })

    svc.Start()
}
```

### Health Check Configuration

You can configure health check endpoints using environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `HEALTH_PATH` | `/health` | Main health check endpoint path |
| `READINESS_PATH` | `/ready` | Kubernetes readiness probe path |
| `LIVENESS_PATH` | `/live` | Kubernetes liveness probe path |

### Accessing Health Checker in Handlers

You can access the health checker in your HTTP handlers:

```go
func myHandler(w http.ResponseWriter, r *http.Request) {
    healthChecker := service.GetHealthChecker(r)
    if healthChecker != nil {
        check := healthChecker.Measure(r.Context())
        // Use status information
        w.Write([]byte(fmt.Sprintf("Health checks: %+v", check)))
    }
}
```

## Kubernetes Deployment

The library provides all the boilerplate needed for Kubernetes deployments:

- Graceful shutdown handling SIGTERM
- Health check endpoints for liveness/readiness probes
- Prometheus metrics for monitoring
- Configurable resource limits via environment variables
- No additional Kubernetes-specific code required

## Examples

See the `_examples/` directory for complete working examples demonstrating:

- **minimal/**: Basic service setup with default configuration
- **custom-metrics/**: Comprehensive custom metrics registration and usage
- **prometheus-counter/**: Simple custom counter example (HTTP metrics are automatic)
- **health-check-***: Various health check integrations
- **shutdown-hook/**: Graceful shutdown with custom cleanup

## Best Practices

1. **Minimal setup** - Start with default configuration and customize only what you need
2. **Write HTTP handlers naturally** - Use any patterns or frameworks you prefer
3. **Add custom shutdown hooks** for resource cleanup when needed
4. **Use structured logging** for better observability
5. **Monitor metrics** in production environments

## Contributing

We welcome contributions! Please see our [Contributing Guide](https://github.com/atomicgo/atomicgo/blob/main/CONTRIBUTING.md) for details.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

> [AtomicGo.dev](https://atomicgo.dev) &nbsp;&middot;&nbsp;
> with ❤️ by [@MarvinJWendt](https://github.com/MarvinJWendt) |
> [MarvinJWendt.com](https://marvinjwendt.com)
