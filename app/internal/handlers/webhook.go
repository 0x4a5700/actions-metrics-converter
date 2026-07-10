package handlers

import (
	"context"
	"crypto/hmac"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/0x4a5700/actions-metrics-converter/internal/telemetry"
	"github.com/0x4a5700/actions-metrics-converter/internal/workflow"
	"github.com/0x4a5700/actions-metrics-converter/pkg/github"
	"go.opentelemetry.io/otel"
)

const (
	maxBodyBytes = 25 << 20
	randChars    = "abcdefghijklmnopqrstuvwxyz0123456789"
)

func Webhook(secret []byte) http.HandlerFunc {
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

	// Webhooks configured with content type application/x-www-form-urlencoded
	// deliver the JSON document as a form field: payload=<url-encoded-json>.
	decoded := string(body)
	if ct := r.Header.Get("Content-Type"); strings.HasPrefix(ct, "application/x-www-form-urlencoded") {
		form, err := url.ParseQuery(decoded)
		if err != nil {
			slog.Warn("error parsing form-encoded body, treating as raw", slog.Any("error", err))
		} else {
			decoded = form.Get("payload")
		}
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
	dir := os.Getenv("FAILED_PAYLOAD_DIR")
	if dir == "" {
		dir = "."
	}
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", err
	}
	filename := filepath.Join(dir, fmt.Sprintf("%d-%s.txt", time.Now().Unix(), randomString(5)))
	f, err := os.Create(filename) // #nosec G304 -- path is env-configured and cleaned by filepath.Join
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
		idx, _ := crand.Int(crand.Reader, big.NewInt(int64(len(randChars))))
		b[i] = randChars[idx.Int64()]
	}
	return string(b)
}
