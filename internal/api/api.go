// Package api wires the HTTP routes served by cmd/api.
package api

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/montessantiago6810-design/coldtrace/internal/platform/httpx"
)

// Options configures the API handler.
type Options struct {
	Logger  *slog.Logger
	Version string
}

// HealthResponse is the body of GET /api/v1/health.
type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Version string `json:"version"`
}

// NewHandler returns the API root handler with request IDs and access
// logging applied to every route.
func NewHandler(opts Options) http.Handler {
	logger := opts.Logger
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	version := opts.Version
	if version == "" {
		version = "dev"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, _ *http.Request) {
		httpx.WriteJSON(w, http.StatusOK, HealthResponse{
			Status:  "ok",
			Service: "coldtrace-api",
			Version: version,
		})
	})

	return httpx.RequestID(httpx.AccessLog(logger)(mux))
}
