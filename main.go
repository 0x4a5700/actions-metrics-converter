package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/0x4a5700/actions-metrics-converter/internal/telemetry"
	"github.com/0x4a5700/actions-metrics-converter/internal/workflow"
	github "github.com/0x4a5700/actions-metrics-converter/pkg/github"
	"go.opentelemetry.io/otel"
)

const (
	randChars = "abcdefghijklmnopqrstuvwxyz0123456789"
	port      = 3018
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

	http.HandleFunc("/", handleAny)
	slog.Info("starting server", slog.Int("port", port))
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		slog.Error("problem listening for connections", slog.Any("error", err))
	}
}

func handleAny(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("error reading request body", slog.Any("error", err))
		http.Error(w, "failed to read body", http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			slog.Error("unable to close http body")
		}
	}()

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

	if payload.WorkflowRun.Id != 0 {
		workflowRun(payload)
	} else {
		spans := workflow.ProcessJob(payload)
		telemetry.Emit(r.Context(), otel.Tracer("actions-metrics-converter"), spans)
	}

	slog.Info("request complete", slog.Any("url", r.URL.String()))
	w.WriteHeader(http.StatusNoContent)
}

func workflowRun(payload github.WorkflowJobPayload) {
	slog.Info("workflow run", slog.String("name", payload.WorkflowRun.Name), slog.String("status", payload.WorkflowRun.Status))
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
