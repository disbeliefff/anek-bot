package logger

import (
	"io"
	"log/slog"
	"os"
	"time"
)

var Log *slog.Logger

type Level string

const (
	DebugLevel Level = "debug"
	InfoLevel  Level = "info"
	WarnLevel  Level = "warn"
	ErrorLevel Level = "error"
)

func Init(level string, w io.Writer) {
	if w == nil {
		w = os.Stdout
	}

	opts := &slog.HandlerOptions{
		AddSource: true,
		Level:     parseLevel(level),
	}

	Log = slog.New(slog.NewJSONHandler(w, opts))
}

func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func Debug(msg string, args ...any) {
	Log.Debug(msg, args...)
}

func Info(msg string, args ...any) {
	Log.Info(msg, args...)
}

func Warn(msg string, args ...any) {
	Log.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	Log.Error(msg, args...)
}

func Err(err error) slog.Attr {
	return slog.Any("error", err)
}

func String(key, value string) slog.Attr {
	return slog.String(key, value)
}

func Int(key string, value int) slog.Attr {
	return slog.Int(key, value)
}

func Int64(key string, value int64) slog.Attr {
	return slog.Int64(key, value)
}

func Duration(key string, value time.Duration) slog.Attr {
	return slog.Duration(key, value)
}

func Any(key string, value any) slog.Attr {
	return slog.Any(key, value)
}
