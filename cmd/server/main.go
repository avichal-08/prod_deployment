package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"server/internal/config"
	"server/internal/handlers"
	"server/internal/metrics"
)

func main() {
	cfg := config.Load()

	baseLogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	logger := baseLogger.With(
		"version", cfg.Version,
		"env", cfg.Environment,
		"hostname", cfg.Hostname,
	)

	appServer := handlers.NewServer(cfg, logger)
	handler := metrics.InstrumentMiddleware(appServer.Routes())

	httpServer := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("starting HTTP server", "port", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	select {
	case err := <-serverErrors:
		logger.Error("server failed to start", "err", err)
		os.Exit(1)
	case <-ctx.Done():
		logger.Info("shutdown signal received; marking server not ready and draining connections")
	}

	appServer.SetReady(false)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown timed out or failed; forcing close", "err", err)
		_ = httpServer.Close()
		os.Exit(1)
	}

	logger.Info("server exited cleanly")
}
