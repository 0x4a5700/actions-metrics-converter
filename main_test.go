package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func sign(secret, body string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(body))
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestVerifySignature(t *testing.T) {
	secret := []byte("s3cret")
	body := []byte(`{"action":"completed"}`)

	tests := []struct {
		name   string
		header string
		want   bool
	}{
		{"valid signature", sign("s3cret", `{"action":"completed"}`), true},
		{"wrong secret", sign("wrong", `{"action":"completed"}`), false},
		{"signature of different body", sign("s3cret", `{}`), false},
		{"missing header", "", false},
		{"missing sha256 prefix", strings.TrimPrefix(sign("s3cret", `{"action":"completed"}`), "sha256="), false},
		{"sha1 prefix", "sha1=deadbeef", false},
		{"invalid hex", "sha256=not-hex!", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, verifySignature(secret, body, tt.header))
		})
	}
}

func TestHandleWebhookSignature(t *testing.T) {
	// A non-completed action so accepted requests stop at the event filter
	// without emitting spans or writing files.
	body := `{"action":"queued"}`

	tests := []struct {
		name       string
		header     string
		wantStatus int
	}{
		{"valid signature is accepted", sign("s3cret", body), http.StatusNoContent},
		{"wrong secret is rejected", sign("wrong", body), http.StatusUnauthorized},
		{"missing signature is rejected", "", http.StatusUnauthorized},
	}

	handler := handleWebhook([]byte("s3cret"))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
			if tt.header != "" {
				req.Header.Set("X-Hub-Signature-256", tt.header)
			}
			rec := httptest.NewRecorder()

			handler(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}
