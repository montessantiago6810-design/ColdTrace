package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/montessantiago6810-design/coldtrace/internal/platform/httpx"
)

func TestHealth(t *testing.T) {
	h := NewHandler(Options{Version: "1.2.3"})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/health", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("status %d, want 200", rr.Code)
	}
	if rr.Header().Get(httpx.RequestIDHeader) == "" {
		t.Fatal("missing request id header")
	}
	var body HealthResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	want := HealthResponse{Status: "ok", Service: "coldtrace-api", Version: "1.2.3"}
	if body != want {
		t.Fatalf("got %+v, want %+v", body, want)
	}
}

func TestHealthDefaultsVersion(t *testing.T) {
	rr := httptest.NewRecorder()
	NewHandler(Options{}).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/health", nil))

	var body HealthResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Version != "dev" {
		t.Fatalf("version %q, want dev", body.Version)
	}
}

func TestHealthRejectsOtherMethods(t *testing.T) {
	rr := httptest.NewRecorder()
	NewHandler(Options{}).ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/v1/health", nil))

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status %d, want 405", rr.Code)
	}
}

func TestUnknownRouteIs404(t *testing.T) {
	rr := httptest.NewRecorder()
	NewHandler(Options{}).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/nope", nil))

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status %d, want 404", rr.Code)
	}
}
