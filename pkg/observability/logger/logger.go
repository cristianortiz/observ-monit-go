package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.Logger
}

// New creates a new logger instance
func New(level string, isDevelopment bool) (*Logger, error) {

	var config zap.Config
	if isDevelopment {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		config = zap.NewProductionConfig()
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	// Parse log level
	logLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		return nil, err
	}
	config.Level = zap.NewAtomicLevelAt(logLevel)

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
