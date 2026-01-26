package logger

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.Logger
}

// New creates a new logger instance with clean JSON format for Loki
func New(level string, isDevelopment bool) (*Logger, error) {
	// Custom encoder config for clean, parseable JSON logs
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  zapcore.OmitKey, // Omit automatic stacktraces (makes logs huge)
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder, // "info", "error" (lowercase)
		EncodeTime:     zapcore.EpochTimeEncoder,      // Unix timestamp (compatible with Promtail)
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder, // "file.go:123"
	}

	config := zap.Config{
		Level:            zap.NewAtomicLevel(),
		Development:      isDevelopment,
		Encoding:         "json",
		EncoderConfig:    encoderConfig,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	// Parse log level
	logLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		return nil, err
	}
	config.Level = zap.NewAtomicLevelAt(logLevel)

	// Disable sampling in development (see all logs)
	if isDevelopment {
		config.Sampling = nil
	} else {
		config.Sampling = &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		}
	}

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return &Logger{
		Logger: logger,
	}, nil
}

// WithFields añade campos estructurados al logger
func (l *Logger) WithFields(fields ...zap.Field) *Logger {
	return &Logger{
		Logger: l.Logger.With(fields...),
	}
}

// WithComponent adds the actual component
func (l *Logger) WithComponent(component string) *Logger {
	return l.WithFields(zap.String("component", component))
}

// WithRequestID add requestId useful for tracing
func (l *Logger) WithRequestID(requestID string) *Logger {
	return l.WithFields(zap.String("request_id", requestID))
}

// Sync force log flush
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}

// WithTraceContext enhance  logger with context trace_id and span_id
// useful to correlate logs with traces ein Jaeger
func (l *Logger) WithTraceContext(ctx context.Context) *Logger {
	//extract span from context
	span := trace.SpanFromContext(ctx)

	if !span.SpanContext().IsValid() {
		// No  span, return logger without chhanges
		return l
	}

	// get trace_id and span_id
	spanCtx := span.SpanContext()
	traceID := spanCtx.TraceID().String()
	spanID := spanCtx.SpanID().String()

	return l.WithFields(
		zap.String("trace_id", traceID),
		zap.String("span_id", spanID),
	)
}

// FromContext extrae el logger del context, o retorna uno por defecto
func FromContext(ctx context.Context) *Logger {
	// Intentar extraer logger del context
	if log := ctx.Value("logger"); log != nil {
		if l, ok := log.(*Logger); ok {
			return l
		}
	}

	// Fallback: retornar logger con nivel Info (para evitar nil panics)
	fallbackLogger, _ := zap.NewProduction()
	return &Logger{Logger: fallbackLogger.Sugar().Desugar()}
}
