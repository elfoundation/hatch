package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthz(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()

	healthz(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	got := strings.TrimSpace(string(body))
	if got != "ok" {
		t.Fatalf("expected body 'ok', got '%s'", got)
	}
}

func TestHealthzSmoke(t *testing.T) {
	// Smoke test: boot the server on a random port, hit /healthz, verify 200 + "ok".
	srv := httptest.NewServer(http.HandlerFunc(healthz))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatalf("failed to GET /healthz: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	got := strings.TrimSpace(string(body))
	if got != "ok" {
		t.Fatalf("expected body 'ok', got '%s'", got)
	}
}
