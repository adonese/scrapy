package logger

import (
	"log/slog"
	"os"
)

var Log *slog.Logger

// Init initializes the global structured logger
func Init() {
	Log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

// Info logs an informational message
func Info(msg string, args ...any) {
	Log.Info(msg, args...)
}

// Error logs an error message
func Error(msg string, args ...any) {
	Log.Error(msg, args...)
}

// Debug logs a debug message
func Debug(msg string, args ...any) {
	Log.Debug(msg, args...)
}

// Warn logs a warning message
func Warn(msg string, args ...any) {
	Log.Warn(msg, args...)
}
