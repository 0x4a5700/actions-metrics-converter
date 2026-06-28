package main

import (
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	randChars = "abcdefghijklmnopqrstuvwxyz0123456789"
	port      = 3018
)

func main() {
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

	filename := fmt.Sprintf("%d-%s.txt", time.Now().Unix(), randomString(5))
	f, err := os.Create(filename)
	if err != nil {
		slog.Warn("error creating file", slog.Any("error", err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
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
	decoded, err := url.QueryUnescape(string(body))
	if err != nil {
		slog.Warn("error url-decoding body, writing raw", slog.Any("error", err))
		decoded = string(body)
	}
	_, err = fmt.Fprint(f, decoded)
	if err != nil {
		slog.Warn("error writing to file", slog.Any("error", err))
	}
	slog.Info("request complete", slog.Any("url", r.URL.String()))
	w.WriteHeader(http.StatusNoContent)
}

func randomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = randChars[rand.Intn(len(randChars))]
	}
	return string(b)
}
