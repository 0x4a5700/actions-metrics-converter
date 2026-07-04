package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/0x4a5700/actions-metrics-converter/internal"
	"github.com/stretchr/testify/assert"
)

func TestHealth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	c := internal.AppConfig{
		BuildDate:    "2026-07-04T11:32:32Z",
		BuildVersion: "development",
		GitHash:      "1234567890abcdef1234567890abcdef",
		GoVersion:    "1.26.4",
	}

	Health(c)(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var got internal.AppConfig
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal response body %q: %v", rec.Body.String(), err)
	}
	assert.Equal(t, c, got)
}
