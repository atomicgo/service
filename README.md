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

A lightweight, production-ready HTTP service framework for Go applications designed to be Kubernetes-ready and follow best practices for highly available microservices.

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

| Variable | Default | Description |
|----------|---------|-------------|
| `ADDR` | `:8080` | HTTP server address |
| `METRICS_ADDR` | `:9090` | Metrics server address |
| `METRICS_PATH` | `/metrics` | Metrics endpoint path |
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

### Running the Example

```bash
cd _example
go run main.go
```

Test the endpoints:
```bash
curl http://localhost:8080/
curl http://localhost:8080/health
curl http://localhost:8080/metrics-demo
curl http://localhost:9090/metrics
```

## Testing

The framework includes comprehensive tests and benchmarks:

```bash
# Run all tests
go test -v ./...

# Run benchmarks
go test -bench=. ./...

# Run with coverage
go test -cover ./...
```

## Best Practices

1. **Always use graceful shutdown** for production deployments
2. **Configure appropriate timeouts** based on your application needs
3. **Add custom shutdown hooks** for resource cleanup
4. **Use structured logging** for better observability
5. **Monitor metrics** in production environments
6. **Set up health checks** for Kubernetes deployments

## Dependencies

- `github.com/caarlos0/env/v11`: Environment variable parsing
- `github.com/prometheus/client_golang`: Prometheus metrics
- `log/slog`: Structured logging (Go 1.21+)

## Contributing

We welcome contributions! Please see our [Contributing Guide](https://github.com/atomicgo/atomicgo/blob/main/CONTRIBUTING.md) for details.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

> [AtomicGo.dev](https://atomicgo.dev) &nbsp;&middot;&nbsp;
> with ❤️ by [@MarvinJWendt](https://github.com/MarvinJWendt) |
> [MarvinJWendt.com](https://marvinjwendt.com)
