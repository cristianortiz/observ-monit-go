package logger

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// OTELLoggerConfig configuración para el logger con OTEL Bridge
type OTELLoggerConfig struct {
	ServiceName    string // Nombre del servicio (e.g., "factorit")
	ServiceVersion string // Versión del servicio (e.g., "1.0.0")
	Environment    string // Entorno (development, staging, production)
	OTLPEndpoint   string // Endpoint del OTEL Collector (e.g., "otel-collector:4317")
	LogLevel       string // Nivel de logs (debug, info, warn, error)
	Enabled        bool   // Si está habilitado el envío a OTEL
}

// NewOTELLogger crea un logger Zap con OTEL Bridge
//
// Este logger funciona en modo "Tee" (Y):
// - Siempre escribe a console (stdout) → para docker logs
// - Si Enabled=true, también envía a OTEL Collector → Loki
//
// BENEFICIO: Correlación automática con traces
// - Si el log ocurre dentro de un span, automáticamente incluye trace_id y span_id
// - Permite hacer "jump to trace" desde Loki a Jaeger
//
// Retorna:
// - *Logger: El logger configurado
// - func(context.Context) error: Función de shutdown (IMPORTANTE llamar en defer)
// - error: Error si falla la inicialización
func NewOTELLogger(ctx context.Context, config OTELLoggerConfig) (*Logger, func(context.Context) error, error) {
	// ========================================
	// 1. CREAR ENCODER DE ZAP (formato JSON)
	// ========================================
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder   // Formato: 2025-12-10T10:30:00.123Z
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder // INFO, ERROR, DEBUG (mayúsculas)

	// Parsear nivel de log (info, debug, error, etc.)
	level, err := zapcore.ParseLevel(config.LogLevel)
	if err != nil {
		level = zapcore.InfoLevel // Default a INFO si el nivel es inválido
	}

	// ========================================
	// 2. CORE PARA CONSOLE (stdout)
	// ========================================
	// Este core SIEMPRE está activo (para docker logs y debugging local)
	consoleCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig), // Formato JSON
		zapcore.AddSync(os.Stdout),            // Escribir a stdout
		level,                                 // Nivel mínimo de logs
	)

	// Iniciar con el core de console
	var cores []zapcore.Core
	cores = append(cores, consoleCore)

	var shutdownFunc func(context.Context) error

	// ========================================
	// 3. SI OTEL ESTÁ HABILITADO, AGREGAR BRIDGE
	// ========================================
	if config.Enabled {
		// 3.1. Crear Resource (metadata del servicio)
		// Esta información se agrega a TODOS los logs enviados a OTEL
		res, err := resource.New(ctx,
			resource.WithAttributes(
				semconv.ServiceName(config.ServiceName),           // service.name = "factorit"
				semconv.ServiceVersion(config.ServiceVersion),     // service.version = "1.0.0"
				semconv.DeploymentEnvironment(config.Environment), // deployment.environment = "development"
			),
		)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create OTEL resource: %w", err)
		}

		// 3.2. Crear OTLP Log Exporter (envía logs al Collector vía gRPC)
		logExporter, err := otlploggrpc.New(ctx,
			otlploggrpc.WithEndpoint(config.OTLPEndpoint),
			otlploggrpc.WithInsecure(),
			// Para producción, usar:
			// otlploggrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(certPool, ""))
		)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create OTLP log exporter: %w", err)
		}

		// 3.3. Crear LoggerProvider (gestor de logs OTEL)
		// Similar a TracerProvider (traces) y MeterProvider (metrics)
		loggerProvider := log.NewLoggerProvider(
			log.WithResource(res), // Agregar metadata del servicio
			log.WithProcessor(
				// BatchProcessor: Agrupa logs antes de enviar (eficiencia)
				log.NewBatchProcessor(
					logExporter,
					log.WithExportInterval(10*time.Second), // Enviar batch cada 10 segundos
					log.WithExportMaxBatchSize(512),        // O cuando haya 512 logs acumulados
					// Esto evita hacer 1 request HTTP por cada log (overhead)
				),
			),
		)

		// 3.4. Crear OTEL Bridge Core (interceptor de logs)
		// Este es el componente mágico que:
		// - Intercepta todos los logs de Zap
		// - Los convierte al formato de OTEL
		// - Agrega automáticamente trace_id y span_id si están disponibles
		// - Los envía al LoggerProvider → Exporter → Collector
		//
		// NOTA: El filtrado por nivel (level) se hace en el Tee de Zap, no aquí
		// El otelCore recibe todos los logs que pasen el filtro del consoleCore
		otelCore := otelzap.NewCore(
			config.ServiceName,
			otelzap.WithLoggerProvider(loggerProvider),
			otelzap.WithVersion(config.ServiceVersion), // Versión del servicio
		)

		// Agregar el core OTEL a la lista de cores
		cores = append(cores, otelCore)

		// Función de shutdown para liberar recursos al cerrar la app
		shutdownFunc = func(ctx context.Context) error {
			if err := loggerProvider.Shutdown(ctx); err != nil {
				return fmt.Errorf("failed to shutdown logger provider: %w", err)
			}
			return nil
		}
	} else {
		// Si OTEL está deshabilitado, shutdown es no-op (no hace nada)
		shutdownFunc = func(context.Context) error { return nil }
	}

	// ========================================
	// 4. COMBINAR CORES CON TEE (Y)
	// ========================================
	// Tee = "Y" en plomería: un log se escribe a múltiples destinos
	//
	//       Zap Logger
	//           │
	//           ├─────────┬─────────┐
	//           │         │         │
	//        Console   OTEL    (otros)
	//       (stdout)  (gRPC)
	//
	core := zapcore.NewTee(cores...)

	// ========================================
	// 5. CREAR LOGGER DE ZAP
	// ========================================
	zapLogger := zap.New(core,
		zap.AddCaller(),                       // Agregar línea/archivo del log
		zap.AddStacktrace(zapcore.ErrorLevel), // Agregar stacktrace para ERROR+
	)

	// ========================================
	// 6. ENVOLVER EN NUESTRO LOGGER CUSTOM
	// ========================================
	logger := &Logger{
		Logger: zapLogger,
	}

	return logger, shutdownFunc, nil
}
