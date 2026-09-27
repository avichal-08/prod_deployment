package handlers_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"server/internal/config"
	"server/internal/handlers"
)

func newTestServer() *handlers.Server {
	cfg := config.Config{
		Port:        "8080",
		Environment: "test",
		Version:     "1.0.0-test",
		Hostname:    "test-pod-01",
	}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	return handlers.NewServer(cfg, logger)
}

func TestHealthAndReadinessToggle(t *testing.T) {
	srv := newTestServer()
	handler := srv.Routes()

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	srv.SetReady(false)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req)

	if rec2.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503 after SetReady(false), got %d", rec2.Code)
	}

	healthReq := httptest.NewRequest(http.MethodGet, "/health", nil)
	healthRec := httptest.NewRecorder()
	handler.ServeHTTP(healthRec, healthReq)

	if healthRec.Code != http.StatusOK {
		t.Fatalf("expected /health to remain 200 OK during drain, got %d", healthRec.Code)
	}
}

func TestUsersEndpointMetadata(t *testing.T) {
	srv := newTestServer()
	handler := srv.Routes()

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response JSON: %v", err)
	}

	if body["hostname"] != "test-pod-01" {
		t.Errorf("expected hostname test-pod-01, got %v", body["hostname"])
	}
	if body["version"] != "1.0.0-test" {
		t.Errorf("expected version 1.0.0-test, got %v", body["version"])
	}
}
