package log

import (
	"log/slog"
	"os"
)

var defaultLogger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

func SetDebug(debug bool) {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}
	defaultLogger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}

func Debug(msg string, args ...any) {
	defaultLogger.Debug(msg, args...)
}

func Info(msg string, args ...any) {
	defaultLogger.Info(msg, args...)
}

func Error(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
}

func Get() *slog.Logger {
	return defaultLogger
}
