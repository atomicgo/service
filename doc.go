/*
Package service provides a minimal boilerplate wrapper for building production-ready Go HTTP services.

This library reduces the boilerplate of writing production/enterprise-grade Go services to a minimum.
It does NOT provide an HTTP framework, business logic, or impose restrictions on web frameworks.
Instead, it's a lightweight wrapper around http.Server that provides essential production features
out of the box for high availability services.

Key benefits:
- Minimal boilerplate for Kubernetes and containerized production deployments
- Built-in Prometheus metrics collection and health checks
- Graceful shutdown with signal handling
- Structured logging with slog integration
- Environment-based configuration with sensible defaults
- Extensible middleware system
- No restrictions on HTTP frameworks

The framework is designed to be a thin layer that handles the operational concerns of production
services while letting you write HTTP handlers exactly as you prefer, using any patterns or
frameworks you choose.
*/
package service
