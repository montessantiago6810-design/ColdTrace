package httpx

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func echoID() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(RequestIDFrom(r.Context())))
	})
}

func TestRequestIDGeneratesWhenMissing(t *testing.T) {
	rr := httptest.NewRecorder()
	RequestID(echoID()).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	id := rr.Header().Get(RequestIDHeader)
	if len(id) != 32 {
		t.Fatalf("generated id %q: want 32 hex characters", id)
	}
	if rr.Body.String() != id {
		t.Fatalf("context id %q differs from header id %q", rr.Body.String(), id)
	}
}

func TestRequestIDReusesValidIncoming(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(RequestIDHeader, "gw-T17_0304.backfill-1")
	rr := httptest.NewRecorder()
	RequestID(echoID()).ServeHTTP(rr, req)

	if got := rr.Header().Get(RequestIDHeader); got != "gw-T17_0304.backfill-1" {
		t.Fatalf("got %q, want the incoming id", got)
	}
}

func TestRequestIDReplacesInvalidIncoming(t *testing.T) {
	for _, bad := range []string{
		"has space",
		"line\nbreak",
		`quote"`,
		strings.Repeat("a", 65),
	} {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(RequestIDHeader, bad)
		rr := httptest.NewRecorder()
		RequestID(echoID()).ServeHTTP(rr, req)

		if got := rr.Header().Get(RequestIDHeader); got == bad || len(got) != 32 {
			t.Errorf("incoming %q: got %q, want a fresh generated id", bad, got)
		}
	}
}

func TestAccessLogRecordsStatusAndOmitsQuery(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	h := RequestID(AccessLog(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusTeapot)
	})))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/shipments?token=secret", nil)
	req.Header.Set(RequestIDHeader, "req-1")
	h.ServeHTTP(httptest.NewRecorder(), req)

	var rec map[string]any
	if err := json.Unmarshal(buf.Bytes(), &rec); err != nil {
		t.Fatalf("log line is not JSON: %v (%q)", err, buf.String())
	}
	if rec["status"] != float64(http.StatusTeapot) || rec["path"] != "/api/v1/shipments" ||
		rec["method"] != "GET" || rec["request_id"] != "req-1" {
		t.Fatalf("unexpected access log record: %v", rec)
	}
	if strings.Contains(buf.String(), "secret") {
		t.Fatalf("query string leaked into the log: %q", buf.String())
	}
}

func TestAccessLogDefaultsTo200WhenHandlerWritesNothing(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	AccessLog(logger)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})).
		ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	if !strings.Contains(buf.String(), `"status":200`) {
		t.Fatalf("want status 200 in %q", buf.String())
	}
}

func TestWriteJSON(t *testing.T) {
	rr := httptest.NewRecorder()
	WriteJSON(rr, http.StatusCreated, map[string]int{"tempCMilli": 8420})

	if rr.Code != http.StatusCreated {
		t.Fatalf("status %d, want 201", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("content type %q", ct)
	}
	if cc := rr.Header().Get("Cache-Control"); cc != "no-store" {
		t.Fatalf("cache control %q", cc)
	}
	if rr.Body.String() != "{\"tempCMilli\":8420}\n" {
		t.Fatalf("body %q", rr.Body.String())
	}
}

func TestWriteJSONEncodingFailure(t *testing.T) {
	rr := httptest.NewRecorder()
	WriteJSON(rr, http.StatusOK, map[string]any{"bad": make(chan int)})
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status %d, want 500", rr.Code)
	}
}
