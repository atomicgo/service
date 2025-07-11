/*
Package service provides a lightweight, production-ready HTTP service framework for Go applications.

The framework is designed to be Kubernetes-ready and follows best practices for highly available
microservices. It includes built-in support for graceful shutdown, Prometheus metrics, structured
logging with slog, extensible middleware, and environment-based configuration.

Key features:
- Configurable HTTP server with graceful shutdown
- Built-in Prometheus metrics collection
- Structured logging with context-aware loggers
- Extensible middleware system
- Environment-based configuration with sensible defaults
- Health check endpoints
- Kubernetes-ready design

The framework is designed to be lightweight while providing all essential features for
production-ready microservices.
*/
package service
