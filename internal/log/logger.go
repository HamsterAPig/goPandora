package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	log *zap.Logger
)

// InitLogger initializes the logger with the specified log level.
// Valid log levels are "debug", "info", "warn", "error", and "fatal".
// The log level is case-insensitive.
func InitLogger(level string) error {
	logLevel := parseLogLevel(level)
	if logLevel == zapcore.Level(-1) {
		return fmt.Errorf("invalid log level: %s", level)
	}

	cfg := zap.Config{
		Encoding:         "json",
		Level:            zap.NewAtomicLevelAt(logLevel),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig:    zap.NewProductionEncoderConfig(),
	}

	cfg.EncoderConfig.EncodeCaller = zapcore.FullCallerEncoder

	var err error
	log, err = cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		return err
	}

	zap.RedirectStdLog(log)

	return nil
}

// parseLogLevel converts a log level string to the corresponding zapcore.Level.
// It returns -1 if the log level is invalid.
func parseLogLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.Level(-1)
	}
}

// Debug logs a debug level message.
func Debug(msg string, fields ...zap.Field) {
	log.Debug(msg, fields...)
}

// Info logs an info level message.
func Info(msg string, fields ...zap.Field) {
	log.Info(msg, fields...)
}

// Warn logs a warning level message.
func Warn(msg string, fields ...zap.Field) {
	log.Warn(msg, fields...)
}

// Error logs an error level message.
func Error(msg string, fields ...zap.Field) {
	log.Error(msg, fields...)
}

// Fatal logs a fatal level message and then exits the program.
func Fatal(msg string, fields ...zap.Field) {
	log.Fatal(msg, fields...)
	os.Exit(1)
}
