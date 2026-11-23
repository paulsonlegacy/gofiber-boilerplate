// Package logger provides a Zap-based logger with automatic log rotation via lumberjack.
// Call InitLogger once at startup — it replaces the global zap logger so
// zap.L() and zap.S() work everywhere without passing the logger around.
package logger

import (
	"os"
	"path/filepath"

	"boilerplate/internal/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger is the globally accessible logger instance.
var Logger *zap.Logger

func newRotatingWriter(path string) *lumberjack.Logger {
	_ = os.MkdirAll(filepath.Dir(path), os.ModePerm)
	return &lumberjack.Logger{
		Filename: path,
		MaxSize:  10,   // MB before rotation
		MaxAge:   7,    // days to keep rotated files
		Compress: true, // gzip rotated files
	}
}

// InitLogger initializes a production-grade Zap logger and registers it globally.
// LOCAL-DEV → Debug level (verbose). All other envs → Info level.
func InitLogger(cfg config.LoggingConfig) *zap.Logger {
	logLevel := zapcore.InfoLevel
	if cfg.EnvType == "LOCAL-DEV" {
		logLevel = zapcore.DebugLevel
	}

	generalWriter := zapcore.AddSync(newRotatingWriter(cfg.LogFilePath))
	errorWriter := zapcore.AddSync(newRotatingWriter(cfg.ErrorLogFilePath))

	// Fan out: all logs → general file, errors only → error file
	core := zapcore.NewTee(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
			generalWriter,
			logLevel,
		),
		zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
			errorWriter,
			zapcore.ErrorLevel,
		),
	)

	logger := zap.New(core, zap.AddCaller())
	Logger = logger
	zap.ReplaceGlobals(logger)

	return logger
}
