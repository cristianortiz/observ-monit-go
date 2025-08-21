# 🔍 E-Commerce Observability Platform

A comprehensive microservices project built with Go, focused on implementing enterprise-grade **Observability** and **Monitoring** patterns. This project prioritizes observability implementation over complex business logic, making it perfect for learning distributed systems monitoring.

## 🎯 Project Overview

This is a simple e-commerce system designed as a learning platform for mastering **Monitoring and Observability** in microservices architecture. The project implements the three pillars of observability while maintaining clean, production-ready code using Clean Architecture principles.

### Key Learning Objectives

- **Distributed Tracing**: End-to-end request tracking across services
- **Metrics Collection**: Business and system metrics with Prometheus
- **Centralized Logging**: Structured logging with correlation
- **Health Monitoring**: Service health checks and alerting
- **Performance Monitoring**: APM and profiling integration
- **API Gateway Integration**: Context propagation through KrakenD

## 📊 Observability Fundamentals

### Monitoring vs Observability

**Monitoring** is the process of collecting, aggregating, and analyzing metrics to understand system behavior. It tells you **WHAT** is happening.

**Observability** is the ability to understand the internal state of a system based on its external outputs. It helps you understand **WHY** something is happening.

### The Three Pillars of Observability

#### 1. 📈 Metrics

Numerical data aggregated over time that provide quantitative insights into system behavior.

- **Types**: CPU usage, memory consumption, request latency, throughput, error rates
- **Tools**: Prometheus, Grafana
- **Use Cases**: Alerting, capacity planning, performance trending

#### 2. 📝 Logs

Discrete event records that provide detailed information about system activities.

- **Types**: Structured (JSON) vs unstructured text logs
- **Levels**: DEBUG, INFO, WARN, ERROR, FATAL
- **Tools**: ELK Stack, Fluentd, Loki
- **Use Cases**: Debugging, audit trails, security analysis

#### 3. 🔗 Traces

Distributed request tracking that shows the journey of a request across multiple services.

- **Components**: Spans, trace context, baggage
- **Standards**: OpenTelemetry, OpenTracing
- **Tools**: Jaeger, Zipkin
- **Use Cases**: Performance optimization, dependency mapping, error root cause analysis

### Observability Methodologies

#### Golden Signals (Google SRE)

1. **Latency**: Response time of requests
2. **Traffic**: Demand on the system (requests per second)
3. **Errors**: Rate of failed requests
4. **Saturation**: How "full" the service is (resource utilization)

#### RED Method (Request-oriented services)

- **Rate**: Requests per second
- **Errors**: Error rate percentage
- **Duration**: Response time distribution

#### USE Method (Resource-oriented)

- **Utilization**: Percentage of time the resource is busy
- **Saturation**: Amount of work queued
- **Errors**: Error count for the resource

## 🏗️ Architecture Overview

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Client    │────│   KrakenD   │────│   Users     │
│             │    │   Gateway   │    │  Service    │
└─────────────┘    └─────────────┘    └─────────────┘
                           │
                           ├──────────────────────────┐
                           │                          │
                           ▼                          ▼
                   ┌─────────────┐            ┌─────────────┐
                   │  Products   │            │  Orders     │
                   │  Service    │◄───────────┤  Service    │
                   └─────────────┘            └─────────────┘
                           │                          │
                           ▼                          ▼
                   ┌─────────────────────────────────────────┐
                   │         Observability Stack             │
                   │  ┌─────────┐ ┌─────────┐ ┌─────────┐   │
                   │  │Prometheus│ │ Jaeger  │ │ Grafana │   │
                   │  └─────────┘ └─────────┘ └─────────┘   │
                   │  ┌─────────┐ ┌─────────┐               │
                   │  │  Loki   │ │AlertMgr │               │
                   │  └─────────┘ └─────────┘               │
                   └─────────────────────────────────────────┘
```

## 📁 Project Structure

This project follows **Clean Architecture** principles with a focus on maintainability and testability:

```
observ-monit-go/
├── cmd/                           # 🎯 Applications (main packages)
│   ├── users/                     # Users microservice
│   ├── products/                  # Products microservice
│   └── orders/                    # Orders microservice
│
├── internal/                      # 🔒 Private application code
│   ├── domain/                    # 🏛️ Domain Layer (Entities)
│   │   ├── user/                  # User domain logic
│   │   ├── product/               # Product domain logic
│   │   └── order/                 # Order domain logic
│   │
│   ├── usecases/                  # 🎯 Use Cases Layer (Business Rules)
│   │   ├── user/                  # User business logic
│   │   ├── product/               # Product business logic
│   │   └── order/                 # Order business logic
│   │
│   ├── adapters/                  # 🔌 Interface Adapters
│   │   ├── http/                  # HTTP handlers (Fiber v2)
│   │   └── persistence/           # Database implementations
│   │       └── postgres/          # PostgreSQL repositories
│   │
│   └── infrastructure/            # 🛠️ Frameworks & Drivers
│       ├── server/                # HTTP server (Fiber v2)
│       │   └── middleware/        # HTTP middlewares
│       ├── database/              # Database connections
│       └── config/                # Configuration management
│
├── pkg/                          # 📦 Shared/Reusable code
│   ├── observability/            # 🎯 Core observability components
│   │   ├── logger/               # Structured logging (Zap)
│   │   ├── metrics/              # Prometheus metrics
│   │   ├── tracing/              # OpenTelemetry tracing
│   │   └── health/               # Health check endpoints
│   │
│   ├── http-util/                # HTTP utilities
│   ├── validation/               # Request validation
│   └── database/                 # Database utilities
│
├── configs/                      # 📝 Configuration files
│   ├── prometheus/               # Prometheus configuration
│   ├── grafana/                  # Grafana dashboards
│   ├── jaeger/                   # Jaeger configuration
│   └── krakend/                  # API Gateway configuration
│
├── migrations/                   # 🗃️ Database migrations
├── scripts/                      # 🔧 Utility scripts
└── docs/                        # 📖 Documentation
    └── api/                     # API documentation
```

### Architecture Principles

#### Clean Architecture Layers

1. **Domain Layer** (`internal/domain/`): Core business entities and rules

   - Contains business logic and domain models
   - No dependencies on external frameworks
   - Defines repository interfaces

2. **Use Cases Layer** (`internal/usecases/`): Application business rules

   - Orchestrates the flow of data between entities
   - Contains application-specific business rules
   - Implements domain interfaces

3. **Interface Adapters** (`internal/adapters/`): Convert data between layers

   - HTTP handlers (controllers)
   - Repository implementations
   - External service adapters

4. **Frameworks & Drivers** (`internal/infrastructure/`): External concerns
   - Web frameworks (Fiber v2)
   - Database drivers
   - Configuration management

#### Benefits of This Structure

- **Dependency Inversion**: Core business logic doesn't depend on frameworks
- **Testability**: Easy to mock interfaces and test business logic
- **Observability Integration**: Instrumentation at every layer
- **Maintainability**: Clear separation of concerns
- **Flexibility**: Easy to swap implementations (database, HTTP framework, etc.)

## 🛠️ Technology Stack

### Backend Services

- **Language**: Go 1.21+
- **HTTP Framework**: Fiber v2 (high-performance web framework)
- **Database**: PostgreSQL with pgx driver
- **Migrations**: golang-migrate

### Observability Stack

- **Tracing**: OpenTelemetry + Jaeger
- **Metrics**: Prometheus + Grafana
- **Logging**: Zap (structured logging) + Loki
- **Health Checks**: Custom health endpoints
- **APM**: Go pprof integration

### Infrastructure

- **API Gateway**: KrakenD (with OpenTelemetry integration)
- **Containerization**: Docker + Docker Compose
- **Configuration**: Viper (12-factor app compliance)

### Key Go Libraries

```go
// HTTP & Web
github.com/gofiber/fiber/v2
github.com/gofiber/contrib/otelfiber

// Observability
go.opentelemetry.io/otel
go.opentelemetry.io/otel/exporters/jaeger
github.com/prometheus/client_golang
go.uber.org/zap

// Database
github.com/jackc/pgx/v5
github.com/golang-migrate/migrate/v4

// Configuration & Validation
github.com/spf13/viper
github.com/go-playground/validator/v10
```

## 🚀 Getting Started

> **Note**: This project is currently under development. Setup instructions will be added as we progress through each implementation phase.

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- PostgreSQL (or use Docker Compose)

### Current Status

- [x] Project structure setup
- [x] Clean Architecture foundation
- [ ] Basic Users service implementation
- [ ] Observability stack configuration
- [ ] Docker Compose setup

## 📈 Observability Features (Planned)

### Metrics Collection

- **Golden Signals**: Latency, Traffic, Errors, Saturation
- **Business Metrics**: Orders/minute, User registrations, Revenue
- **System Metrics**: CPU, Memory, Database connections
- **Custom Metrics**: Domain-specific measurements

### Distributed Tracing

- **Request Tracing**: End-to-end request journey
- **Context Propagation**: Trace context across service boundaries
- **Database Spans**: Database operation tracing
- **Error Tracking**: Error propagation and correlation

### Centralized Logging

- **Structured Logging**: JSON-formatted logs with context
- **Log Correlation**: Connect logs with traces using correlation IDs
- **Log Levels**: Configurable log levels per service
- **Log Aggregation**: Centralized log collection and searching

### Health Monitoring

- **Health Endpoints**: Service health status
- **Readiness Checks**: Service readiness for traffic
- **Liveness Checks**: Service health for container orchestration
- **Dependency Checks**: External dependency health monitoring

## 🎯 Learning Path

This project is designed to be built incrementally, with each phase building upon the previous one:

1. **Week 1**: Basic structure + Health checks + Metrics
2. **Week 2**: Distributed tracing with OpenTelemetry
3. **Week 3**: Inter-service communication with tracing
4. **Week 4**: API Gateway integration (KrakenD)
5. **Week 5**: Advanced metrics + Grafana dashboards
6. **Week 6**: Centralized logging + Alerting
7. **Week 7**: Performance monitoring + Profiling
8. **Week 8**: Chaos engineering + Production readiness

## 📚 Resources

- [OpenTelemetry Go Documentation](https://opentelemetry.io/docs/instrumentation/go/)
- [Prometheus Best Practices](https://prometheus.io/docs/practices/)
- [Google SRE Book](https://sre.google/books/)
- [Clean Architecture by Robert Martin](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Fiber v2 Documentation](https://docs.gofiber.io/)

## 🤝 Contributing

This is a learning project focused on observability patterns. Contributions are welcome, especially those that improve observability implementation or documentation.

## 📄 License

This project is open source and available under the [MIT License](LICENSE).

---

**Note**: This project prioritizes observability implementation over complex business logic. The e-commerce domain is kept simple intentionally to focus on monitoring and observability patterns.
