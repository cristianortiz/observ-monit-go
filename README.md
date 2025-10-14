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

##### What are Metrics?

Metrics are **time-series data points** that measure specific aspects of a system. Each metric consists of:
- **Name**: Identifier for the metric (e.g., `http_requests_total`)
- **Value**: Numerical measurement
- **Timestamp**: When the measurement was taken
- **Labels**: Key-value pairs for filtering and grouping (e.g., `method="GET"`, `status="200"`)

##### Prometheus Metric Types

Prometheus supports four fundamental metric types:

###### 1. **Counter** 📈 (Only Goes Up)

A cumulative metric that **only increases** or resets to zero on restart.

**Characteristics:**
- Starts at 0 when the service starts
- Only increments, never decrements
- Resets to 0 on service restart
- Used to track totals over time

**Use Cases:**
- Total number of HTTP requests
- Total number of errors
- Total bytes processed
- Total items sold

**Example:**
```go
httpRequestsTotal.Inc() // Increment by 1

// With labels
httpRequestsTotal.WithLabelValues("users-service", "GET", "/api/users", "200").Inc()
```

**Prometheus Query Examples:**
```promql
# Total requests in the last hour
http_requests_total[1h]

# Rate of requests per second
rate(http_requests_total[5m])

# Total errors in the last hour
rate(http_requests_total{status=~"5.."}[1h])
```

**Real-world Example:**
```
Time     Value    Meaning
10:00    0        Service started
10:01    150      150 requests processed
10:02    325      175 new requests (325 - 150)
10:03    500      175 new requests (500 - 325)
```

###### 2. **Gauge** 🌡️ (Goes Up and Down)

A metric that represents a **single value that can go up or down**.

**Characteristics:**
- Can increase or decrease
- Represents a snapshot of a current state
- No automatic reset on restart
- Used for values that fluctuate

**Use Cases:**
- Current memory usage
- Active connections
- Queue size
- Current temperature
- Number of items in cache
- Goroutines count

**Example:**
```go
activeConnections.Inc()    // Increment by 1
activeConnections.Dec()    // Decrement by 1
activeConnections.Set(42)  // Set to specific value
activeConnections.Add(5)   // Add 5
activeConnections.Sub(3)   // Subtract 3
```

**Prometheus Query Examples:**
```promql
# Current active connections
http_active_connections

# Average connections over 5 minutes
avg_over_time(http_active_connections[5m])

# Maximum connections in the last hour
max_over_time(http_active_connections[1h])
```

**Real-world Example:**
```
Time     Value    Meaning
10:00    5        5 active connections
10:01    10       5 new connections opened (went up)
10:02    3        7 connections closed (went down)
10:03    15       12 new connections opened (went up)
```

**Important Note:** Gauges don't have "zones" (like red/yellow/green) built-in. You define alert thresholds in Alertmanager or visualize them in Grafana:

```yaml
# Alertmanager rule
- alert: HighConnectionCount
  expr: http_active_connections > 100
  for: 5m
  annotations:
    summary: "High connection count detected"
```

###### 3. **Histogram** 📊 (Distribution with Buckets)

A metric that **samples observations** and counts them in configurable buckets, allowing calculation of quantiles.

**Characteristics:**
- Groups values into predefined buckets (ranges)
- Automatically calculates sum and count
- Enables percentile calculation (p50, p95, p99)
- Server-side (Prometheus) calculates quantiles
- More flexible than Summary for aggregation

**Use Cases:**
- Request latency/duration
- Response size distribution
- Processing time
- Query execution time

**What are Buckets?**

Buckets are **predefined ranges** where Prometheus groups measurements. Think of them as "bins" that count how many observations fall into each range.

**Default Buckets (`prometheus.DefBuckets`):**
```go
[]float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}
```

**Translated to milliseconds:**
- `0.005s` = 5ms
- `0.01s` = 10ms
- `0.025s` = 25ms
- `0.05s` = 50ms
- `0.1s` = 100ms
- `0.25s` = 250ms
- `0.5s` = 500ms
- `1s` = 1 second
- `2.5s` = 2.5 seconds
- `5s` = 5 seconds
- `10s` = 10 seconds

**How Buckets Work - Practical Example:**

Imagine you receive these request latencies:
```
Request 1: 0.003s (3ms)
Request 2: 0.020s (20ms)
Request 3: 0.150s (150ms)
Request 4: 0.600s (600ms)
Request 5: 1.200s (1.2s)
```

**Prometheus groups them into cumulative buckets:**
```
≤ 0.005s (≤5ms):    1 request   [Request 1: 3ms]
≤ 0.01s (≤10ms):    1 request   [same as above]
≤ 0.025s (≤25ms):   2 requests  [Request 1, 2]
≤ 0.05s (≤50ms):    2 requests  [same as above]
≤ 0.1s (≤100ms):    2 requests  [same as above]
≤ 0.25s (≤250ms):   3 requests  [Request 1, 2, 3]
≤ 0.5s (≤500ms):    3 requests  [same as above]
≤ 1s (≤1s):         3 requests  [same as above]
≤ 2.5s (≤2.5s):     5 requests  [Request 1, 2, 3, 4, 5]
≤ 5s (≤5s):         5 requests  [same as above]
≤ 10s (≤10s):       5 requests  [same as above]
≤ +Inf:             5 requests  [all requests]
```

**What Prometheus Actually Records:**
```
http_request_duration_seconds_bucket{le="0.005"} 1
http_request_duration_seconds_bucket{le="0.01"} 1
http_request_duration_seconds_bucket{le="0.025"} 2
http_request_duration_seconds_bucket{le="0.25"} 3
http_request_duration_seconds_bucket{le="2.5"} 5
http_request_duration_seconds_bucket{le="+Inf"} 5
http_request_duration_seconds_sum 2.973      ← Sum of all durations
http_request_duration_seconds_count 5        ← Total observations
```

**Example:**
```go
// Record a request duration
requestDuration.Observe(0.150) // 150ms

// With labels
httpRequestDuration.WithLabelValues("users-service", "GET", "/api/users", "200").Observe(0.150)
```

**Prometheus Query Examples:**
```promql
# Calculate 95th percentile (p95) - 95% of requests are faster than this
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# Calculate 50th percentile (median) - half of requests are faster
histogram_quantile(0.50, rate(http_request_duration_seconds_bucket[5m]))

# Calculate 99th percentile (p99) - 99% of requests are faster
histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))

# Average request duration
rate(http_request_duration_seconds_sum[5m]) / rate(http_request_duration_seconds_count[5m])
```

**Custom Buckets for Different Use Cases:**

```go
// For very fast APIs (microseconds to milliseconds)
Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1}

// For slow background processes (seconds to minutes)
Buckets: []float64{1, 5, 10, 30, 60, 120, 300, 600}

// For file uploads (exponential growth)
Buckets: prometheus.ExponentialBuckets(1, 2, 10) // 1, 2, 4, 8, 16, 32, 64, 128, 256, 512 MB

// For specific SLA requirements (e.g., p95 < 200ms)
Buckets: []float64{0.01, 0.05, 0.1, 0.2, 0.3, 0.5, 1, 2, 5}
```

###### 4. **Summary** 📋 (Pre-calculated Quantiles)

Similar to Histogram but calculates **quantiles on the client side** (within the application).

**Characteristics:**
- Calculates quantiles in the application
- Cannot aggregate across multiple instances
- Lower server load (pre-calculated)
- Less flexible than Histogram
- Good for single-instance services

**Use Cases:**
- Request/Response size
- Processing time (when you don't need aggregation)
- Single-instance metrics

**Histogram vs Summary:**

| Feature | Histogram | Summary |
|---------|-----------|---------|
| Quantile Calculation | Server (Prometheus) | Client (Application) |
| Aggregation | ✅ Yes (across instances) | ❌ No |
| Flexibility | ✅ High | ⚠️ Limited |
| Server Load | Higher | Lower |
| Bucket Configuration | Required | Not needed |
| Use When | Multiple instances | Single instance |

**Example:**
```go
// Record response size
httpResponseSize.Observe(5000) // 5KB response

// With labels
httpResponseSize.WithLabelValues("users-service", "GET", "/api/users").Observe(5000)
```

**What Summary Records:**
```
http_response_size_bytes_sum 500000      ← Total bytes sent
http_response_size_bytes_count 100       ← Number of responses
```

**Calculate Average:**
```
Average = 500000 / 100 = 5000 bytes per response
```

**Prometheus Query Examples:**
```promql
# Average response size
rate(http_response_size_bytes_sum[5m]) / rate(http_response_size_bytes_count[5m])

# Total bytes sent per second
rate(http_response_size_bytes_sum[5m])
```

##### Metrics Comparison Table

| Type | Direction | Resets on Restart | Aggregatable | Best For |
|------|-----------|-------------------|--------------|----------|
| **Counter** | ↑ Only Up | ✅ Yes (to 0) | ✅ Yes | Totals, rates |
| **Gauge** | ↑↓ Up/Down | ❌ No | ⚠️ Limited | Current values |
| **Histogram** | ↑ Only Up | ✅ Yes (to 0) | ✅ Yes | Latencies, sizes |
| **Summary** | ↑ Only Up | ✅ Yes (to 0) | ❌ No | Single instance |

##### Golden Signals Implementation

Based on Google's SRE principles, four critical metrics should be tracked:

**1. Latency** (How long requests take)
```go
// Histogram for request duration
httpRequestDuration := prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name: "http_request_duration_seconds",
        Help: "HTTP request duration in seconds",
        Buckets: prometheus.DefBuckets,
    },
    []string{"service", "method", "path", "status"},
)
```

**2. Traffic** (How many requests)
```go
// Counter for total requests
httpRequestsTotal := prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "http_requests_total",
        Help: "Total number of HTTP requests",
    },
    []string{"service", "method", "path", "status"},
)
```

**3. Errors** (How many requests fail)
```go
// Counter filtered by error status codes
// Query: rate(http_requests_total{status=~"5.."}[5m])
```

**4. Saturation** (How "full" the system is)
```go
// Gauge for active connections
activeConnections := prometheus.NewGauge(
    prometheus.GaugeOpts{
        Name: "http_active_connections",
        Help: "Number of active HTTP connections",
    },
)
```

##### Practical Metrics Usage

**Complete Request Flow:**

```
┌─────────────────────────────────────────────────────────────────┐
│ 1. Normal HTTP Request (e.g., GET /api/users/123)              │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│ 2. Metrics Middleware Captures Data                            │
│                                                                  │
│    BEFORE c.Next():                                            │
│    - activeConnections.Inc()        → Gauge: 5 → 6            │
│    - start := time.Now()                                       │
│    - path := c.Route().Path         → "/api/users/:id"        │
│                                                                  │
│    c.Next() → Your handler runs                                │
│                                                                  │
│    AFTER c.Next():                                             │
│    - duration := time.Since(start)  → 0.125s                  │
│    - httpRequestsTotal.Inc()        → Counter: 1547 → 1548    │
│    - httpRequestDuration.Observe()  → Histogram: +1 sample    │
│    - httpRequestSize.Observe()      → Summary: 150 bytes      │
│    - httpResponseSize.Observe()     → Summary: 5000 bytes     │
│    - activeConnections.Dec()        → Gauge: 6 → 5            │
└─────────────────────────────────────────────────────────────────┘
```

**How Prometheus Collects These Metrics:**

```
┌──────────────────────────────────────────────────────────────┐
│ Your Go Application                                           │
│                                                               │
│  User Requests  →  Middleware  →  Metrics Registry          │
│  (Business)         (Captures)     (Stores in memory)        │
│                                                               │
│                                    ↓                          │
│                           /metrics Endpoint                   │
│                           (Exposes as text)                   │
└────────────────────────────────┬─────────────────────────────┘
                                 │
                                 │ GET /metrics
                                 │ every 15 seconds
                                 │
                 ┌───────────────▼──────────────┐
                 │                              │
                 │    Prometheus Server         │
                 │                              │
                 │  • Scrapes (pulls) metrics   │
                 │  • Stores time-series data   │
                 │  • Enables PromQL queries    │
                 │                              │
                 └───────────────┬──────────────┘
                                 │
                                 │ Data source
                                 │
                 ┌───────────────▼──────────────┐
                 │                              │
                 │         Grafana              │
                 │                              │
                 │  • Visualizes dashboards     │
                 │  • Creates alerts            │
                 │                              │
                 └──────────────────────────────┘
```

**What `/metrics` Endpoint Returns:**

When Prometheus scrapes `GET http://localhost:3000/metrics`, it receives plain text:

```
# HELP http_requests_total Total number of HTTP requests
# TYPE http_requests_total counter
http_requests_total{method="GET",path="/api/users/:id",service="users-service",status="200"} 1547
http_requests_total{method="POST",path="/api/users",service="users-service",status="201"} 256

# HELP http_request_duration_seconds HTTP request duration in seconds
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{method="GET",path="/api/users/:id",status="200",le="0.005"} 234
http_request_duration_seconds_bucket{method="GET",path="/api/users/:id",status="200",le="0.01"} 987
http_request_duration_seconds_bucket{method="GET",path="/api/users/:id",status="200",le="0.025"} 2150
http_request_duration_seconds_bucket{method="GET",path="/api/users/:id",status="200",le="+Inf"} 3421
http_request_duration_seconds_sum{method="GET",path="/api/users/:id",status="200"} 245.67
http_request_duration_seconds_count{method="GET",path="/api/users/:id",status="200"} 3421

# HELP http_active_connections Number of active HTTP connections
# TYPE http_active_connections gauge
http_active_connections 5
```

**Understanding the Format:**

```
metric_name{label1="value1",label2="value2"} numerical_value

Example:
http_requests_total{method="GET",path="/api/users/:id",status="200"} 1547
│                   │                                            │
│                   └── Labels (for filtering)                  └── The actual value
└── Metric name
```

**Why c.Route().Path Instead of c.Path()?**

```go
// ❌ BAD: c.Path() creates too many unique metrics (high cardinality)
// Requests: /api/users/1, /api/users/2, /api/users/3...
http_requests_total{path="/api/users/1"} 1
http_requests_total{path="/api/users/2"} 1
http_requests_total{path="/api/users/3"} 1
// Result: Thousands of time series = Memory explosion 💥

// ✅ GOOD: c.Route().Path groups by route template (low cardinality)
// All requests to /api/users/:id
http_requests_total{path="/api/users/:id"} 1000
// Result: Single time series = Efficient ✨
```

**Basic PromQL Queries:**

```promql
# Request rate (requests per second)
rate(http_requests_total[5m])

# Error rate percentage
rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m]) * 100

# 95th percentile latency (95% of requests are faster than this)
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# Current active connections
http_active_connections
```

<!-- ...existing code continues... -->

#### 2. 📝 Logs

Discrete event records that provide detailed information about system activities.

- **Types**: Structured (JSON) vs unstructured text logs
- **Levels**: DEBUG, INFO, WARN, ERROR, FATAL
- **Tools**: Zap (structured logging), Loki (aggregation), Grafana (visualization)
- **Use Cases**: Debugging, audit trails, security analysis, error investigation

#### 3. 🔗 Traces

Distributed request tracking that shows the journey of a request across multiple services.

- **Components**: Spans (units of work), trace context (correlation IDs), baggage (metadata)
- **Standards**: OpenTelemetry (unified standard), OpenTracing (legacy)
- **Tools**: Jaeger (tracing backend), Zipkin (alternative)
- **Use Cases**: Performance optimization, dependency mapping, error root cause analysis, latency debugging

### Observability Methodologies

#### Golden Signals (Google SRE)

The four metrics that should be measured for every user-facing system:

1. **Latency**: Response time of requests (measured with Histogram)
2. **Traffic**: Demand on the system - requests per second (measured with Counter + rate())
3. **Errors**: Rate of failed requests (measured with Counter filtered by status)
4. **Saturation**: How "full" the service is - resource utilization (measured with Gauge)

#### RED Method (Request-oriented services)

Focused on user-facing services:

- **Rate**: Requests per second
- **Errors**: Error rate percentage  
- **Duration**: Response time distribution (p50, p95, p99)

#### USE Method (Resource-oriented)

Focused on infrastructure and resource monitoring:

- **Utilization**: Percentage of time the resource is busy
- **Saturation**: Amount of work queued (waiting)
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

- Go 1.24+
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
