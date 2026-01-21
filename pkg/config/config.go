package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all configuration for our application
type Config struct {
	Environment   string
	Debug         bool
	Service       ServiceConfig
	Database      DatabaseConfig
	Observability ObservabilityConfig
	Security      SecurityConfig
	API           ApiConfig
}

type ApiConfig struct {
	ApiName  string
	BasePath string
}
type ServiceConfig struct {
	Name string
	Host string
	Port int
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
	//pool related
	MaxConns            int32
	MinConns            int32
	MaxConnLifetime     time.Duration
	MaxConnIdleTime     time.Duration
	HealthCheckInterval time.Duration
}

type ObservabilityConfig struct {
	LogLevel       string
	LogFormat      string // json o console
	MetricsPath    string
	HealthPath     string
	ReadyPath      string
	MetricsEnabled bool
	Tracing        TracingConfig
	Logging        LoggingConfig // 🆕 Configuración de logs OTEL
}

type TracingConfig struct {
	Enabled        bool
	ServiceName    string
	ServiceVersion string
	OTLPEndpoint   string
	Environment    string
}

// LoggingConfig configuración para enviar logs a OTEL Collector
type LoggingConfig struct {
	Enabled        bool   // Si está habilitado el envío a OTEL (false = solo console)
	ServiceName    string // Nombre del servicio (reutiliza del servicio principal)
	ServiceVersion string // Versión del servicio
	OTLPEndpoint   string // Endpoint del Collector (mismo que tracing)
	Environment    string // Entorno (development, staging, production)
}

type SecurityConfig struct {
	JWTSecret string // Legacy (deprecated)
	// Auth0 OAuth 2.0
	AuthEnabled  bool
	AuthJWKSUrl  string
	AuthAudience string
	AuthIssuer   string
}

// Load reads configuration from environment variables
func Load(serviceName string) (*Config, error) {
	config := &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Debug:       getEnvBool("DEBUG", true),
		Service: ServiceConfig{
			Name: serviceName,
			Host: getEnv("SERVICE_HOST", "0.0.0.0"),
			Port: getEnvInt("SERVICE_PORT", 8080),
		},
		Database: DatabaseConfig{
			Host:                getEnv("DB_HOST", "localhost"),
			Port:                getEnvInt("DB_PORT", 5432),
			User:                getEnv("DB_USER", "postgres"),
			Password:            getEnv("DB_PASSWORD", "postgres"),
			Name:                getEnv("DB_NAME", "observ-db"),
			SSLMode:             getEnv("DB_SSL_MODE", "disable"),
			MaxConns:            int32(getEnvInt("DB_MAX_CONNS", 25)),
			MinConns:            int32(getEnvInt("DB_MIN_CONNS", 5)),
			MaxConnLifetime:     getEnvDuration("DB_MAX_CONN_LIFETIME", 1*time.Hour),
			MaxConnIdleTime:     getEnvDuration("DB_MAX_CONN_IDLE_TIME", 30*time.Minute),
			HealthCheckInterval: getEnvDuration("DB_HEALTH_CHECK_INTERVAL", 1*time.Minute),
		},
		Observability: ObservabilityConfig{
			LogLevel:       getEnv("LOG_LEVEL", "info"),
			LogFormat:      getEnv("LOG_FORMAT", "json"),
			MetricsPath:    getEnv("METRICS_PATH", "/metrics"),
			HealthPath:     getEnv("HEALTH_PATH", "/health"),
			ReadyPath:      getEnv("READY_PATH", "/ready"),
			MetricsEnabled: getEnvBool("METRICS_ENABLED", true),

			Tracing: TracingConfig{
				Enabled:        getEnvBool("TRACING_ENABLED", true),
				ServiceName:    getEnv("TRACING_SERVICE_NAME", serviceName),
				ServiceVersion: getEnv("TRACING_SERVICE_VERSION", "1.0.0"),
				OTLPEndpoint:   getEnv("TRACING_OTLP_ENDPOINT", "localhost:4317"),
				Environment:    getEnv("TRACING_ENVIRONMENT", "development"),
			},

			// 🆕 Configuración de Logs OTEL
			// Por defecto reutiliza los mismos valores que Tracing (mismo Collector)
			Logging: LoggingConfig{
				Enabled:        getEnvBool("LOGGING_OTEL_ENABLED", true), // false = solo console
				ServiceName:    getEnv("LOGGING_SERVICE_NAME", serviceName),
				ServiceVersion: getEnv("LOGGING_SERVICE_VERSION", "1.0.0"),
				OTLPEndpoint:   getEnv("LOGGING_OTLP_ENDPOINT", "localhost:4317"), // Mismo Collector
				Environment:    getEnv("LOGGING_ENVIRONMENT", "development"),
			},
		},
		Security: SecurityConfig{
			JWTSecret:    getEnv("JWT_SECRET", "change-me-in-production"),
			AuthEnabled:  getEnvBool("AUTH_ENABLED", false),
			AuthJWKSUrl:  getEnv("AUTH_JWKS_URL", ""),
			AuthAudience: getEnv("AUTH_AUDIENCE", ""),
			AuthIssuer:   getEnv("AUTH_ISSUER", ""),
		},
		API: ApiConfig{
			BasePath: getEnv("API_BASE_PATH", "/api/v1"),
		},
	}

	// Service-specific port override
	servicePortEnv := strings.ToUpper(strings.ReplaceAll(serviceName, "-", "_")) + "_PORT"
	// Note: removed println to avoid non-JSON logs that break Loki parsing
	if port := os.Getenv(servicePortEnv); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Service.Port = p
		}
	}

	return config, config.Validate()
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func (c *Config) Validate() error {
	if c.Service.Name == "" {
		return fmt.Errorf("service name is required")
	}
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	return nil
}

func (c *Config) GetDatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.Name,
		c.Database.SSLMode,
	)
}

func (c *Config) GetServiceAddress() string {
	return fmt.Sprintf("%s:%d", c.Service.Host, c.Service.Port)
}

func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
