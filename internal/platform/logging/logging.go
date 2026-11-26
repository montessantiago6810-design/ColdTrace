// Package logging builds the structured slog logger used by every binary.
package logging

import (
	"io"
	"log/slog"
)

// New returns a logger writing to w in the given format ("text" or, by
// default, "json") at the given minimum level. Every record carries the
// service name so logs from several binaries can share one sink.
func New(w io.Writer, format string, level slog.Level, service string) *slog.Logger {
	opts := &slog.HandlerOptions{Level: level}
	var h slog.Handler
	if format == "text" {
		h = slog.NewTextHandler(w, opts)
	} else {
		h = slog.NewJSONHandler(w, opts)
	}
	return slog.New(h).With(slog.String("service", service))
}
