package logging

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func TestNewJSONIncludesServiceAndRespectsLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, "json", slog.LevelInfo, "api")

	logger.Debug("hidden")
	logger.Info("started", "addr", "127.0.0.1:8080")

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("got %d lines, want 1 (debug must be filtered): %q", len(lines), buf.String())
	}
	var rec map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &rec); err != nil {
		t.Fatalf("not JSON: %v", err)
	}
	if rec["service"] != "api" || rec["msg"] != "started" || rec["addr"] != "127.0.0.1:8080" {
		t.Fatalf("unexpected record: %v", rec)
	}
}

func TestNewText(t *testing.T) {
	var buf bytes.Buffer
	New(&buf, "text", slog.LevelDebug, "ingestor").Debug("tick")
	out := buf.String()
	if !strings.Contains(out, "service=ingestor") || !strings.Contains(out, "msg=tick") {
		t.Fatalf("unexpected text output: %q", out)
	}
}
