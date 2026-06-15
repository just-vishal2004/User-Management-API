package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Log is the global logger instance.
// Other packages call logger.Log.Info(), logger.Log.Error(), etc.
var Log *zap.Logger

// Init initializes the global Zap logger based on the environment.
// Call this once in main.go before starting the server.
// In production: outputs JSON (machine-readable).
// In development: outputs colorized console output (human-readable).
func Init(env string) {
	var err error

	if env == "production" {
		// Production logger:
		// - JSON format for log aggregation tools (Datadog, CloudWatch)
		// - Info level and above (no debug noise)
		// - Timestamps in ISO8601 format
		Log, err = zap.NewProduction()
	} else {
		// Development logger:
		// - Human-readable console format with colors
		// - Debug level and above (see everything)
		// - Caller information (which file/line logged this)
		Log, err = newDevelopmentLogger()
	}

	if err != nil {
		// If the logger itself fails to initialize,
		// we have no choice but to panic — the app
		// cannot run safely without logging.
		panic("failed to initialize logger: " + err.Error())
	}
}

// newDevelopmentLogger builds a Zap logger configured for
// local development — colorized, human-readable, with caller info.
func newDevelopmentLogger() (*zap.Logger, error) {
	cfg := zap.NewDevelopmentEncoderConfig()

	// Color the log level label (INFO=blue, ERROR=red, etc.)
	cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

	// Format timestamps as readable strings
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		// Console encoder: human-readable key=value format
		zapcore.NewConsoleEncoder(cfg),
		// Write to stdout
		zapcore.AddSync(os.Stdout),
		// Log everything at Debug level and above
		zapcore.DebugLevel,
	)

	return zap.New(core, zap.AddCaller()), nil
}

// Sync flushes any buffered log entries.
// Call this with defer in main.go to ensure all logs are written
// before the application exits.
func Sync() {
	if Log != nil {
		// We intentionally ignore the error here —
		// Sync can fail on some OS configurations when
		// stdout is not a regular file, and that's acceptable.
		_ = Log.Sync()
	}
}
