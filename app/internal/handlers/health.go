package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/0x4a5700/actions-metrics-converter/internal"
)

func Health(c internal.AppConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		b, err := json.Marshal(c)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Error determining health"))
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(b)
	}
}
