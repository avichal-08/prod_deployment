package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"

	"server/internal/config"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	cfg     config.Config
	logger  *slog.Logger
	isReady *atomic.Bool
}

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

func NewServer(cfg config.Config, logger *slog.Logger) *Server {
	ready := &atomic.Bool{}
	ready.Store(true)

	return &Server{
		cfg:     cfg,
		logger:  logger,
		isReady: ready,
	}
}

func (s *Server) SetReady(ready bool) {
	s.isReady.Store(ready)
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", s.handleRoot)
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /ready", s.handleReady)
	mux.HandleFunc("GET /api/users", s.handleUsers)
	mux.Handle("GET /metrics", promhttp.Handler())

	return s.loggingMiddleware(mux)
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, http.StatusOK, map[string]any{
		"message":     "Go production server running",
		"version":     s.cfg.Version,
		"environment": s.cfg.Environment,
		"hostname":    s.cfg.Hostname,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, http.StatusOK, map[string]any{
		"status":   "alive",
		"hostname": s.cfg.Hostname,
	})
}

func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	if !s.isReady.Load() {
		s.writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"status":   "not_ready",
			"hostname": s.cfg.Hostname,
		})
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]any{
		"status":   "ready",
		"hostname": s.cfg.Hostname,
	})
}

func (s *Server) handleUsers(w http.ResponseWriter, r *http.Request) {
	if sleepParam := r.URL.Query().Get("sleep"); sleepParam != "" {
		if d, err := time.ParseDuration(sleepParam); err == nil && d > 0 {
			s.logger.Info("simulating slow request", "sleep_duration", d.String())
			select {
			case <-time.After(d):
			case <-r.Context().Done():
				s.logger.Warn("request context canceled before completion", "err", r.Context().Err())
				return
			}
		}
	}

	users := []User{
		{ID: 1, Name: "Alice Chen", Role: "Platform Engineer"},
		{ID: 2, Name: "Marcus Vance", Role: "SRE"},
		{ID: 3, Name: "Elena Rostova", Role: "Backend Engineer"},
	}

	s.writeJSON(w, http.StatusOK, map[string]any{
		"version":     s.cfg.Version,
		"environment": s.cfg.Environment,
		"hostname":    s.cfg.Hostname,
		"data":        users,
	})
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		if r.URL.Path != "/health" && r.URL.Path != "/ready" && r.URL.Path != "/metrics" {
			s.logger.Info("http_request",
				"method", r.Method,
				"path", r.URL.Path,
				"remote_addr", r.RemoteAddr,
				"duration_ms", time.Since(start).Milliseconds(),
			)
		}
	})
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		s.logger.Error("failed to write json response", "err", err)
	}
}
