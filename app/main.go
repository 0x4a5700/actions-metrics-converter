package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/0x4a5700/actions-metrics-converter/internal"
	"github.com/0x4a5700/actions-metrics-converter/internal/apperr"
	"github.com/0x4a5700/actions-metrics-converter/internal/handlers"
	telemetry2 "github.com/0x4a5700/actions-metrics-converter/internal/telemetry"
)

const (
	port = 3018
	// GitHub caps webhook payloads at 25 MB.
)

var (
	buildDate    string
	buildVersion string
	gitHash      string
	goVersion    string
)

func main() {
	if err := run(); err != nil {
		slog.Error("exiting", slog.Any("error", err))
		os.Exit(1)
	}
}

// run holds main's logic so its defers (tracer shutdown, flushing spans) run
// on every exit path before main decides the exit code.
func run() error {
	ctx := context.Background()

	shutdown, err := telemetry2.InitProvider(ctx, "actions-metrics-converter")
	if err != nil {
		return fmt.Errorf("initialise tracer provider: %w", err)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			slog.Error("error shutting down tracer provider", slog.Any("error", err))
		}
	}()

	secret, err := getSecretBytes()
	if err != nil {
		return err
	}

	c := internal.AppConfig{
		BuildDate:    buildDate,
		BuildVersion: buildVersion,
		GitHash:      gitHash,
		GoVersion:    goVersion,
	}

	// GitHub gives webhook deliveries ~10s before marking them failed, so
	// there is no value in letting requests linger much longer than that.
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	http.HandleFunc("/", handlers.Webhook(secret))
	http.HandleFunc("/health", handlers.Health(c))

	errCh := make(chan error, 1)
	go func() {
		slog.Info("starting server", slog.Int("port", port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return fmt.Errorf("listening for connections: %w", err)
	case <-quit:
	}

	slog.Info("shutting down server")
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}
	return nil
}

// getSecretBytes reads the webhook secret from the environment, failing if it
// is unset so the server never starts up accepting unsigned requests.
func getSecretBytes() ([]byte, error) {
	secret := os.Getenv("GITHUB_WEBHOOK_SECRET")
	if secret == "" {
		return nil, apperr.MissingEnvVar{VarName: "GITHUB_WEBHOOK_SECRET", Fatal: true}
	}
	return []byte(secret), nil
}
