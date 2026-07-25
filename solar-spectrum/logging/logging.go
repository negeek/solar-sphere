// Package logging sets up the structured (log/slog) logger shared by every
// solar-sphere service, so log level and format are controlled the same way
// everywhere via the LOG_LEVEL/LOG_FORMAT environment variables.
package logging

import (
	"log/slog"
	"os"
	"strings"
)

// New builds a slog.Logger for service, reading LOG_LEVEL (debug, info,
// warn, error; default info) and LOG_FORMAT (json, text; default json) from
// the environment. The returned logger tags every record with "service".
func New(service string) *slog.Logger {
	opts := &slog.HandlerOptions{Level: level()}

	var handler slog.Handler
	if strings.EqualFold(os.Getenv("LOG_FORMAT"), "text") {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(handler).With("service", service)
}

func level() slog.Level {
	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
