package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/elfoundation/hatch/internal/handler"
	"github.com/elfoundation/hatch/internal/store"
	"github.com/go-chi/chi/v5"
)

func TestHealthzViaRouter(t *testing.T) {
	repo, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer repo.Close()

	r := chi.NewRouter()
	h := handler.New(repo)
	h.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}
	body, _ := io.ReadAll(w.Result().Body)
	w.Result().Body.Close()
	got := strings.TrimSpace(string(body))
	if got != "ok" {
		t.Fatalf("expected body 'ok', got %q", got)
	}
}

func TestHealthzSmoke(t *testing.T) {
	// Smoke: boot server on random port, hit /healthz.
	// Use :memory: store so no filesystem dependency.
	t.Setenv("HATCH_DB_PATH", ":memory:")

	repo, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer repo.Close()

	r := chi.NewRouter()
	h := handler.New(repo)
	h.RegisterRoutes(r)

	srv := httptest.NewServer(r)
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
		t.Fatalf("expected body 'ok', got %q", got)
	}
}
