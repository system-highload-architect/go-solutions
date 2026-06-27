// Package logger provides a simple way to create structured loggers using
// the standard library's slog package. It supports two formats (text, json)
// and four levels (debug, info, warn, error).
//
// Example usage:
//
//	package main
//
//	import (
//	    "github.com/system-highload-architect/go-solutions/logger"
//	    "log/slog"
//	)
//
//	func main() {
//	    log := logger.New("info", "json", slog.String("service", "myapp"))
//	    log.Info("server started", "port", 8080)
//	}
package logger

import (
	"log/slog"
	"os"
)

// New creates a new slog.Logger with the given level, format, and default attributes.
// level must be one of: "debug", "info", "warn", "error".
// format must be one of: "text", "json".
// attrs are key‑value pairs added to every log line.
func New(level, format string, attrs ...slog.Attr) *slog.Logger {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	var handler slog.Handler
	if format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	if len(attrs) > 0 {
		handler = handler.WithAttrs(attrs)
	}

	return slog.New(handler)
}
