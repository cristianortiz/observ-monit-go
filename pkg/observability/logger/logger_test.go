package logger

import (
	"testing"

	"go.uber.org/zap"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		level       string
		development bool
		wantErr     bool
	}{
		{
			name:        "development logger with debug level",
			level:       "debug",
			development: true,
			wantErr:     false,
		},
		{
			name:        "production logger with info level",
			level:       "info",
			development: false,
			wantErr:     false,
		},
		{
			name:        "production logger with error level",
			level:       "error",
			development: false,
			wantErr:     false,
		},
		{
			name:        "invalid log level",
			level:       "invalid",
			development: false,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := New(tt.level, tt.development)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && logger == nil {
				t.Error("New() returned nil logger")
			}
			if logger != nil {
				// Test that we can actually log
				logger.Info("test message", zap.String("test", "value"))
				logger.Sync()
			}
		})
	}
}

func TestWithFields(t *testing.T) {
	logger, err := New("info", true)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	// Test adding fields
	loggerWithFields := logger.WithFields(
		zap.String("service", "test"),
		zap.Int("version", 1),
	)

	if loggerWithFields == nil {
		t.Error("WithFields() returned nil")
	}

	// Test logging with fields
	loggerWithFields.Info("test message with fields")
}

func TestWithComponent(t *testing.T) {
	logger, err := New("info", true)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	componentLogger := logger.WithComponent("user-service")
	if componentLogger == nil {
		t.Error("WithComponent() returned nil")
	}

	componentLogger.Info("test component message")
}

func TestWithRequestID(t *testing.T) {
	logger, err := New("info", true)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	requestLogger := logger.WithRequestID("req-123")
	if requestLogger == nil {
		t.Error("WithRequestID() returned nil")
	}

	requestLogger.Info("test request message")
}

func TestLoggerLevels(t *testing.T) {
	logger, err := New("debug", true)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	// Test different log levels
	logger.Debug("debug message", zap.String("level", "debug"))
	logger.Info("info message", zap.String("level", "info"))
	logger.Warn("warn message", zap.String("level", "warn"))
	logger.Error("error message", zap.String("level", "error"))
}
