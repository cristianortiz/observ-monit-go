package main

import (
	"log"

	"github.com/cristianortiz/observ-monit-go/pkg/config"
	"github.com/cristianortiz/observ-monit-go/pkg/observability/logger"
	"go.uber.org/zap"
)

func main() {
	// Cargar configuración
	cfg, err := config.Load("test-service")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Crear logger
	appLogger, err := logger.New(cfg.Observability.LogLevel, cfg.IsDevelopment())
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer appLogger.Sync()

	// 🚀 Ejemplos de uso del logger
	appLogger.Info("Application starting",
		zap.String("service", cfg.Service.Name),
		zap.String("environment", cfg.Environment),
		zap.String("log_level", cfg.Observability.LogLevel),
		zap.String("log_format", cfg.Observability.LogFormat),
		zap.String("address", cfg.GetServiceAddress()),
	)

	// Logger con componente específico
	userLogger := appLogger.WithComponent("user-service")
	userLogger.Info("User service component initialized",
		zap.String("database_url", cfg.GetDatabaseURL()),
	)

	// Logger con request ID (simulando una request HTTP)
	requestLogger := appLogger.WithRequestID("req-12345")
	requestLogger.Info("Processing user request",
		zap.String("action", "create_user"),
		zap.String("user_id", "user-456"),
		zap.String("ip", "192.168.1.100"),
	)

	// Diferentes niveles de log
	appLogger.Debug("Debug message - solo en development",
		zap.String("debug_info", "detailed debugging information"),
	)

	appLogger.Info("Info message - operación normal",
		zap.String("status", "running"),
		zap.Int("active_connections", 5),
	)

	appLogger.Warn("Warning message - algo podría estar mal",
		zap.String("warning", "database connection slow"),
		zap.Duration("latency", 500),
	)

	appLogger.Error("Error message - algo salió mal",
		zap.String("error", "failed to connect to external service"),
		zap.String("service", "payment-gateway"),
	)

	// Logger con múltiples campos estructurados
	orderLogger := appLogger.WithComponent("order-service").WithRequestID("req-67890")
	orderLogger.Info("Order processing",
		zap.String("order_id", "order-123"),
		zap.String("user_id", "user-456"),
		zap.Float64("amount", 99.99),
		zap.String("currency", "USD"),
		zap.Bool("payment_success", true),
	)

	appLogger.Info("Application test completed successfully")
}
