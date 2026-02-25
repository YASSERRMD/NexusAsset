package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a zap logger configured for JSON or console output.
// level: "debug" | "info" | "warn" | "error"
// format: "json" | "console"
func New(level, format string) (*zap.Logger, error) {
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		zapLevel = zapcore.InfoLevel
	}

	var cfg zap.Config
	if format == "console" {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
	}

	cfg.Level = zap.NewAtomicLevelAt(zapLevel)
	cfg.EncoderConfig.TimeKey = "ts"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.CallerKey = "caller"
	cfg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	logger, err := cfg.Build(zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	if err != nil {
		return nil, fmt.Errorf("building zap logger: %w", err)
	}

	return logger, nil
}
