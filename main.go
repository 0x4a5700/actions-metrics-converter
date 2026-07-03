package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/0x4a5700/actions-metrics-converter/internal/telemetry"
	"github.com/0x4a5700/actions-metrics-converter/internal/workflow"
	github "github.com/0x4a5700/actions-metrics-converter/pkg/github"
	"go.opentelemetry.io/otel"
)

const (
	randChars = "abcdefghijklmnopqrstuvwxyz0123456789"
	port      = 3018
	// GitHub caps webhook payloads at 25 MB.
	maxBodyBytes = 25 << 20
)

func main() {
	ctx := context.Background()

	shutdown, err := telemetry.InitProvider(ctx, "actions-metrics-converter")
	if err != nil {
		slog.Error("failed to initialise tracer provider", slog.Any("error", err))
		os.Exit(1)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			slog.Error("error shutting down tracer provider", slog.Any("error", err))
		}
	}()

	secret := os.Getenv("GITHUB_WEBHOOK_SECRET")
	if secret == "" {
		slog.Error("GITHUB_WEBHOOK_SECRET must be set")
		os.Exit(1)
	}

	srv := &http.Server{Addr: fmt.Sprintf(":%d", port)}
	http.HandleFunc("/", handleWebhook([]byte(secret)))

	go func() {
		slog.Info("starting server", slog.Int("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("problem listening for connections", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server")
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown error", slog.Any("error", err))
	}
}

func handleWebhook(secret []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			var maxErr *http.MaxBytesError
			if errors.As(err, &maxErr) {
				slog.Warn("rejecting oversized request body", slog.String("remote_addr", r.RemoteAddr))
				http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
				return
			}
			slog.Error("error reading request body", slog.Any("error", err))
			http.Error(w, "failed to read body", http.StatusInternalServerError)
			return
		}
		defer func() {
			if err := r.Body.Close(); err != nil {
				slog.Error("unable to close http body")
			}
		}()

		if !verifySignature(secret, body, r.Header.Get("X-Hub-Signature-256")) {
			slog.Warn("rejecting request with invalid signature", slog.String("remote_addr", r.RemoteAddr))
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}

		handlePayload(w, r, body)
	}
}

// verifySignature checks the X-Hub-Signature-256 header against an HMAC-SHA256
// of the raw request body, as sent by GitHub webhooks.
func verifySignature(secret, body []byte, header string) bool {
	hexSig, ok := strings.CutPrefix(header, "sha256=")
	if !ok {
		return false
	}
	sig, err := hex.DecodeString(hexSig)
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, secret)
	mac.Write(body)
	return hmac.Equal(mac.Sum(nil), sig)
}

func handlePayload(w http.ResponseWriter, r *http.Request, body []byte) {
	event := r.Header.Get("X-GitHub-Event")
	if event != "workflow_job" && event != "workflow_run" {
		slog.Info("ignoring event", slog.String("event", event))
		w.WriteHeader(http.StatusNoContent)
		return
	}

	decoded, err := url.QueryUnescape(string(body))
	if err != nil {
		slog.Warn("error url-decoding body, treating as raw", slog.Any("error", err))
		decoded = string(body)
	}

	var payload github.WorkflowJobPayload
	if err := json.Unmarshal([]byte(decoded), &payload); err != nil {
		filename, writeErr := writeRequestToFile(r, decoded)
		if writeErr != nil {
			slog.Error("error unmarshalling payload and failed to write to file", slog.Any("error", err), slog.Any("write_error", writeErr))
		} else {
			slog.Error("error unmarshalling payload", slog.Any("error", err), slog.String("file", filename))
		}
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	if payload.Action != "completed" {
		slog.Info("ignoring non-completed event", slog.String("action", payload.Action))
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if event == "workflow_run" {
		workflowRun(r.Context(), payload)
	} else {
		spans := workflow.ProcessJob(payload)
		telemetry.Emit(r.Context(), otel.Tracer("actions-metrics-converter"), spans)
	}

	slog.Info("request complete", slog.Any("url", r.URL.String()))
	w.WriteHeader(http.StatusNoContent)
}

func workflowRun(ctx context.Context, payload github.WorkflowJobPayload) {
	spans := workflow.ProcessRun(payload)
	telemetry.EmitRun(ctx, otel.Tracer("actions-metrics-converter"), spans)
}

func writeRequestToFile(r *http.Request, body string) (string, error) {
	filename := fmt.Sprintf("%d-%s.txt", time.Now().Unix(), randomString(5))
	f, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := f.Close(); err != nil {
			slog.Warn("error closing file", slog.Any("error", err))
		}
	}()

	_, _ = fmt.Fprintf(f, "%s %s %s\n", r.Method, r.URL.String(), r.Proto)
	for name, values := range r.Header {
		for _, v := range values {
			_, _ = fmt.Fprintf(f, "%s: %s\n", name, v)
		}
	}
	_, _ = fmt.Fprintln(f)
	_, err = fmt.Fprint(f, body)
	return filename, err
}

func randomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = randChars[rand.Intn(len(randChars))]
	}
	return string(b)
}
