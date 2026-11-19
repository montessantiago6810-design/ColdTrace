package config

import (
	"log/slog"
	"strings"
	"testing"
	"time"
)

func env(m map[string]string) LookupFunc {
	return func(key string) (string, bool) {
		v, ok := m[key]
		return v, ok
	}
}

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load(env(nil))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg != Defaults() {
		t.Fatalf("got %+v, want defaults %+v", cfg, Defaults())
	}
	if !strings.HasPrefix(cfg.HTTPAddr, "127.0.0.1:") {
		t.Fatalf("default address must be loopback, got %q", cfg.HTTPAddr)
	}
}

func TestLoadOverrides(t *testing.T) {
	cfg, err := Load(env(map[string]string{
		"COLDTRACE_ENV":              "staging",
		"COLDTRACE_HTTP_ADDR":        ":9090",
		"COLDTRACE_LOG_LEVEL":        "debug",
		"COLDTRACE_LOG_FORMAT":       " TEXT ",
		"COLDTRACE_SHUTDOWN_TIMEOUT": "30s",
	}))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	want := Config{
		Env:             "staging",
		HTTPAddr:        ":9090",
		LogLevel:        slog.LevelDebug,
		LogFormat:       "text",
		ShutdownTimeout: 30 * time.Second,
	}
	if cfg != want {
		t.Fatalf("got %+v, want %+v", cfg, want)
	}
}

func TestLoadTreatsEmptyAsUnset(t *testing.T) {
	cfg, err := Load(env(map[string]string{"COLDTRACE_ENV": "  ", "COLDTRACE_HTTP_ADDR": ""}))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg != Defaults() {
		t.Fatalf("got %+v, want defaults", cfg)
	}
}

func TestLoadReportsEveryInvalidVariable(t *testing.T) {
	_, err := Load(env(map[string]string{
		"COLDTRACE_ENV":              "prod",
		"COLDTRACE_HTTP_ADDR":        "8080",
		"COLDTRACE_LOG_LEVEL":        "loud",
		"COLDTRACE_LOG_FORMAT":       "xml",
		"COLDTRACE_SHUTDOWN_TIMEOUT": "-1s",
	}))
	if err == nil {
		t.Fatal("Load: expected an error")
	}
	for _, key := range []string{
		"COLDTRACE_ENV", "COLDTRACE_HTTP_ADDR", "COLDTRACE_LOG_LEVEL",
		"COLDTRACE_LOG_FORMAT", "COLDTRACE_SHUTDOWN_TIMEOUT",
	} {
		if !strings.Contains(err.Error(), key) {
			t.Errorf("error does not mention %s: %v", key, err)
		}
	}
}

func TestLoadRejectsUnparsableTimeout(t *testing.T) {
	for _, v := range []string{"ten seconds", "0s", "10m"} {
		if _, err := Load(env(map[string]string{"COLDTRACE_SHUTDOWN_TIMEOUT": v})); err == nil {
			t.Errorf("timeout %q: expected an error", v)
		}
	}
}
