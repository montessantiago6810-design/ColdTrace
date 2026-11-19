// Package config loads service configuration from environment variables.
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"time"
)

// Config holds the settings shared by the ColdTrace binaries.
type Config struct {
	// Env is one of local, test, staging or production.
	Env string
	// HTTPAddr is the listen address of the HTTP server (host:port).
	HTTPAddr string
	// LogLevel is the minimum level written to the log.
	LogLevel slog.Level
	// LogFormat is "json" (default) or "text".
	LogFormat string
	// ShutdownTimeout bounds graceful shutdown after SIGINT/SIGTERM.
	ShutdownTimeout time.Duration
}

// Defaults returns the configuration used when no variables are set.
// The listen address is loopback-only so a local run is never exposed.
func Defaults() Config {
	return Config{
		Env:             "local",
		HTTPAddr:        "127.0.0.1:8080",
		LogLevel:        slog.LevelInfo,
		LogFormat:       "json",
		ShutdownTimeout: 10 * time.Second,
	}
}

// LookupFunc matches os.LookupEnv so tests can pass a map instead.
type LookupFunc func(key string) (string, bool)

// Load reads COLDTRACE_* variables through lookup and validates them.
// Every invalid variable is reported, not just the first one.
func Load(lookup LookupFunc) (Config, error) {
	cfg := Defaults()
	var errs []error

	if v, ok := get(lookup, "COLDTRACE_ENV"); ok {
		switch v {
		case "local", "test", "staging", "production":
			cfg.Env = v
		default:
			errs = append(errs, fmt.Errorf("COLDTRACE_ENV: must be local, test, staging or production, got %q", v))
		}
	}

	if v, ok := get(lookup, "COLDTRACE_HTTP_ADDR"); ok {
		if _, port, err := net.SplitHostPort(v); err != nil || port == "" {
			errs = append(errs, fmt.Errorf("COLDTRACE_HTTP_ADDR: must be host:port, got %q", v))
		} else {
			cfg.HTTPAddr = v
		}
	}

	if v, ok := get(lookup, "COLDTRACE_LOG_LEVEL"); ok {
		var level slog.Level
		if err := level.UnmarshalText([]byte(v)); err != nil {
			errs = append(errs, fmt.Errorf("COLDTRACE_LOG_LEVEL: must be debug, info, warn or error, got %q", v))
		} else {
			cfg.LogLevel = level
		}
	}

	if v, ok := get(lookup, "COLDTRACE_LOG_FORMAT"); ok {
		switch strings.ToLower(v) {
		case "json", "text":
			cfg.LogFormat = strings.ToLower(v)
		default:
			errs = append(errs, fmt.Errorf("COLDTRACE_LOG_FORMAT: must be json or text, got %q", v))
		}
	}

	if v, ok := get(lookup, "COLDTRACE_SHUTDOWN_TIMEOUT"); ok {
		d, err := time.ParseDuration(v)
		switch {
		case err != nil:
			errs = append(errs, fmt.Errorf("COLDTRACE_SHUTDOWN_TIMEOUT: must be a duration such as 10s, got %q", v))
		case d <= 0 || d > 5*time.Minute:
			errs = append(errs, fmt.Errorf("COLDTRACE_SHUTDOWN_TIMEOUT: must be between 0s and 5m, got %s", d))
		default:
			cfg.ShutdownTimeout = d
		}
	}

	if err := errors.Join(errs...); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// get returns a trimmed value; empty values count as unset.
func get(lookup LookupFunc, key string) (string, bool) {
	v, ok := lookup(key)
	if !ok {
		return "", false
	}
	v = strings.TrimSpace(v)
	return v, v != ""
}
