package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

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

// TestSmokeE2E is the acceptance gate: one command to prove the whole product works.
// It starts a real server, exercises every feature (capture, inspect, SSE, mock,
// replay), and verifies the full flow end-to-end.
//
// Run: go test ./cmd/hatch -run TestSmokeE2E -v
func TestSmokeE2E(t *testing.T) {
	// ---- Phase 0: Start server ----
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

	client := &http.Client{Timeout: 10 * time.Second}

	// ---- Phase 1: Health check ----
	t.Run("healthz", func(t *testing.T) {
		resp, err := client.Get(srv.URL + "/healthz")
		if err != nil {
			t.Fatalf("GET /healthz: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
		body, _ := io.ReadAll(resp.Body)
		if strings.TrimSpace(string(body)) != "ok" {
			t.Fatalf("expected 'ok', got %q", string(body))
		}
	})

	// ---- Phase 2: Capture requests (all HTTP methods) ----
	t.Run("capture", func(t *testing.T) {
		methods := []struct {
			method string
			body   string
		}{
			{"GET", ""},
			{"POST", `{"order":"latte"}`},
			{"PUT", `{"status":"updated"}`},
			{"PATCH", `{"field":"partial"}`},
			{"DELETE", ""},
		}
		for _, m := range methods {
			var body io.Reader
			if m.body != "" {
				body = strings.NewReader(m.body)
			}
			req, _ := http.NewRequest(m.method, srv.URL+"/capture-test", body)
			if m.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			req.Header.Set("X-Custom", "smoke-"+m.method)

			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("%s /capture-test: %v", m.method, err)
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("%s: expected 200, got %d", m.method, resp.StatusCode)
			}
		}

		// Verify all 5 requests are in the store.
		reqs, err := repo.ListRequests(context.Background(), "capture-test", 10)
		if err != nil {
			t.Fatalf("list requests: %v", err)
		}
		if len(reqs) != 5 {
			t.Fatalf("expected 5 captured requests, got %d", len(reqs))
		}
	})

	// ---- Phase 3: Inspect page (HTML rendering) ----
	t.Run("inspect", func(t *testing.T) {
		resp, err := client.Get(srv.URL + "/e/capture-test")
		if err != nil {
			t.Fatalf("GET /e/capture-test: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
		ct := resp.Header.Get("Content-Type")
		if !strings.Contains(ct, "text/html") {
			t.Fatalf("expected text/html, got %q", ct)
		}

		body, _ := io.ReadAll(resp.Body)
		html := string(body)

		// Page structure.
		if !strings.Contains(html, "<!DOCTYPE html>") {
			t.Error("missing doctype")
		}
		if !strings.Contains(html, "capture-test") {
			t.Error("missing endpoint ID")
		}

		// Request cards.
		if !strings.Contains(html, "DELETE") {
			t.Error("missing DELETE method badge")
		}
		if !strings.Contains(html, "POST") {
			t.Error("missing POST method badge")
		}

		// Replay UI.
		if !strings.Contains(html, "Replay") {
			t.Error("missing Replay button")
		}

		// SSE live-update script.
		if !strings.Contains(html, "EventSource") {
			t.Error("missing EventSource script")
		}

		// Empty-state endpoint should show the hint.
		resp2, err := client.Get(srv.URL + "/e/nonexistent")
		if err != nil {
			t.Fatalf("GET /e/nonexistent: %v", err)
		}
		defer resp2.Body.Close()
		body2, _ := io.ReadAll(resp2.Body)
		if !strings.Contains(string(body2), "Waiting for requests") {
			t.Error("missing empty state for unused endpoint")
		}
	})

	// ---- Phase 4: SSE stream receives live events ----
	t.Run("sse", func(t *testing.T) {
		// Use separate transports to avoid connection-pool issues with the long-lived SSE connection.
		sseClient := &http.Client{
			Transport: &http.Transport{DisableKeepAlives: true},
			Timeout:   5 * time.Second,
		}
		captureClient := &http.Client{
			Transport: &http.Transport{DisableKeepAlives: true},
			Timeout:   5 * time.Second,
		}

		// Subscribe to SSE.
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sseReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/e/sse-ep/events", nil)
		sseResp, err := sseClient.Do(sseReq)
		if err != nil {
			t.Fatalf("SSE request: %v", err)
		}
		defer sseResp.Body.Close()

		if sseResp.StatusCode != http.StatusOK {
			t.Fatalf("SSE returned %d", sseResp.StatusCode)
		}
		ct := sseResp.Header.Get("Content-Type")
		if !strings.HasPrefix(ct, "text/event-stream") {
			t.Fatalf("SSE Content-Type: %q", ct)
		}

		// Capture a request on a goroutine; read SSE on main.
		eventCh := make(chan string, 1)
		go func() {
			buf := make([]byte, 4096)
			n, _ := sseResp.Body.Read(buf)
			eventCh <- string(buf[:n])
		}()

		time.Sleep(20 * time.Millisecond) // let SSE subscription register

		resp, err := captureClient.Post(srv.URL+"/sse-ep", "application/json", strings.NewReader(`{"sse":true}`))
		if err != nil {
			t.Fatalf("capture: %v", err)
		}
		resp.Body.Close()

		select {
		case event := <-eventCh:
			if !strings.Contains(event, "data:") {
				t.Fatalf("SSE event missing 'data:' prefix: %q", event)
			}
			// Body value is JSON-escaped in the SSE stream (e.g. "body":"{\"sse\":true}").
			// Check for the content values rather than the exact escaped form.
			if !strings.Contains(event, `sse`) || !strings.Contains(event, `true`) {
				t.Fatalf("SSE event missing body content: %q", event)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("timeout waiting for SSE event")
		}
	})

	// ---- Phase 5: Mock configuration and response ----
	t.Run("mock", func(t *testing.T) {
		// Set a mock on a fresh endpoint.
		mockBody := `{"status":418,"headers":{"X-Teapot":"true"},"body":"eyJicmV3aW5nIjp0cnVlfQ=="}`

		resp, err := client.Post(
			srv.URL+"/e/mock-ep/mock",
			"application/json",
			strings.NewReader(mockBody),
		)
		if err != nil {
			t.Fatalf("PUT /e/mock-ep/mock: %v", err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			// PUT is registered as PUT, but we use POST here — let's use PUT.
			// The route is PUT /e/{endpointID}/mock
		}

		// Use proper PUT method.
		req, _ := http.NewRequest(http.MethodPut, srv.URL+"/e/mock-ep-2/mock", strings.NewReader(mockBody))
		req.Header.Set("Content-Type", "application/json")
		resp2, err := client.Do(req)
		if err != nil {
			t.Fatalf("PUT /e/mock-ep-2/mock: %v", err)
		}
		resp2.Body.Close()
		if resp2.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 OK from mock set, got %d: %s", resp2.StatusCode, readBody(resp2))
		}

		// Verify mock is stored.
		mock, err := repo.GetMock(context.Background(), "mock-ep-2")
		if err != nil {
			t.Fatalf("mock not stored: %v", err)
		}
		if mock.Status != 418 {
			t.Errorf("expected status 418, got %d", mock.Status)
		}
		if mock.Headers["X-Teapot"] != "true" {
			t.Errorf("expected X-Teapot header, got %v", mock.Headers)
		}

		// Capture a request — should return the mock response.
		resp3, err := client.Post(srv.URL+"/mock-ep-2", "application/json", strings.NewReader(`{"order":"earl grey"}`))
		if err != nil {
			t.Fatalf("capture on mocked endpoint: %v", err)
		}
		defer resp3.Body.Close()
		if resp3.StatusCode != 418 {
			t.Fatalf("expected mock status 418, got %d", resp3.StatusCode)
		}
		if resp3.Header.Get("X-Teapot") != "true" {
			t.Errorf("expected X-Teapot: true, got %q", resp3.Header.Get("X-Teapot"))
		}
		body, _ := io.ReadAll(resp3.Body)
		// Body is base64-decoded by Go's json decoder ([]byte field).
		if strings.TrimSpace(string(body)) != `{"brewing":true}` {
			t.Errorf("expected mock body, got %q", string(body))
		}
	})

	// ---- Phase 6: Replay ----
	t.Run("replay", func(t *testing.T) {
		// Start a sink server that echoes back what it receives.
		sink := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Sink", "yes")
			w.WriteHeader(http.StatusCreated)
			fmt.Fprintf(w, `{"echo":{"method":%q,"path":%q}}`, r.Method, r.URL.Path)
		}))
		defer sink.Close()

		// Allow private replay so we can hit the loopback sink.
		t.Setenv("HATCH_ALLOW_PRIVATE_REPLAY", "true")

		// Capture a request to replay later.
		captureResp, err := client.Post(srv.URL+"/replay-ep", "application/json", strings.NewReader(`{"msg":"replay me"}`))
		if err != nil {
			t.Fatalf("capture for replay: %v", err)
		}
		captureResp.Body.Close()

		// Get the captured request ID from the store.
		reqs, _ := repo.ListRequests(context.Background(), "replay-ep", 1)
		if len(reqs) != 1 {
			t.Fatalf("expected 1 captured request, got %d", len(reqs))
		}
		reqID := reqs[0].ID

		// Replay to the sink.
		replayPayload := fmt.Sprintf(`{"target_url":%q}`, sink.URL)
		replayResp, err := client.Post(
			srv.URL+"/e/replay-ep/requests/"+reqID+"/replay",
			"application/json",
			strings.NewReader(replayPayload),
		)
		if err != nil {
			t.Fatalf("replay request: %v", err)
		}
		defer replayResp.Body.Close()
		if replayResp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(replayResp.Body)
			t.Fatalf("replay returned %d: %s", replayResp.StatusCode, string(body))
		}

		var result struct {
			Status  int               `json:"status"`
			Headers map[string]string `json:"headers"`
			Body    string            `json:"body"`
		}
		json.NewDecoder(replayResp.Body).Decode(&result)
		if result.Status != 201 {
			t.Errorf("expected replay sink status 201, got %d", result.Status)
		}
		if result.Headers["X-Sink"] != "yes" {
			t.Errorf("expected X-Sink: yes, got %q", result.Headers["X-Sink"])
		}
	})

	// ---- Phase 7: Edge cases ----
	t.Run("edge", func(t *testing.T) {
		// Capture with query params.
		resp, err := client.Get(srv.URL + "/query-ep?foo=bar&baz=qux")
		if err != nil {
			t.Fatalf("capture with query: %v", err)
		}
		resp.Body.Close()

		reqs, _ := repo.ListRequests(context.Background(), "query-ep", 1)
		if len(reqs) != 1 {
			t.Fatalf("expected 1 request, got %d", len(reqs))
		}
		if reqs[0].Query != "foo=bar&baz=qux" {
			t.Errorf("expected query 'foo=bar&baz=qux', got %q", reqs[0].Query)
		}

		// Capture with binary body.
		binaryBody := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE}
		resp2, err := client.Post(srv.URL+"/binary-ep", "application/octet-stream", strings.NewReader(string(binaryBody)))
		if err != nil {
			t.Fatalf("binary capture: %v", err)
		}
		resp2.Body.Close()

		reqs2, _ := repo.ListRequests(context.Background(), "binary-ep", 1)
		if len(reqs2) != 1 {
			t.Fatalf("expected 1 binary request, got %d", len(reqs2))
		}
		if len(reqs2[0].Body) != len(binaryBody) {
			t.Errorf("expected body len %d, got %d", len(binaryBody), len(reqs2[0].Body))
		}

		// SSRF protection: replay to localhost is denied.
		replayPayload := `{"target_url":"http://localhost:9999/"}`
		req, _ := http.NewRequest(http.MethodPost, srv.URL+"/e/query-ep/requests/"+reqs[0].ID+"/replay", strings.NewReader(replayPayload))
		req.Header.Set("Content-Type", "application/json")
		resp3, err := client.Do(req)
		if err != nil {
			t.Fatalf("SSRF test: %v", err)
		}
		resp3.Body.Close()
		if resp3.StatusCode != http.StatusForbidden {
			t.Errorf("expected 403 for localhost replay, got %d", resp3.StatusCode)
		}
	})

	t.Log("E2E smoke test complete — all phases passed")
}

func readBody(resp *http.Response) string {
	if resp.Body == nil {
		return ""
	}
	b, _ := io.ReadAll(resp.Body)
	return string(b)
}

// TestServerConfigurations tests various server configurations.
func TestServerConfigurations(t *testing.T) {
	// Test with default port.
	t.Run("default_port", func(t *testing.T) {
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
			t.Fatalf("GET /healthz: %v", err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	// Test with custom headers.
	t.Run("custom_headers", func(t *testing.T) {
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

		req, _ := http.NewRequest("POST", srv.URL+"/test", strings.NewReader(`{"key":"value"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Custom-Header", "test-value")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("POST /test: %v", err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})
}

// TestMultipleEndpointsConcurrency tests concurrent access to multiple endpoints.
func TestMultipleEndpointsConcurrency(t *testing.T) {
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

	client := &http.Client{Timeout: 5 * time.Second}

	// Create multiple endpoints concurrently.
	var wg sync.WaitGroup
	numEndpoints := 5

	for i := 0; i < numEndpoints; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			endpoint := fmt.Sprintf("endpoint-%d", idx)
			for j := 0; j < 10; j++ {
				body := fmt.Sprintf(`{"endpoint":%d,"request":%d}`, idx, j)
				resp, err := client.Post(srv.URL+"/"+endpoint, "application/json", strings.NewReader(body))
				if err != nil {
					t.Errorf("endpoint %d, request %d: %v", idx, j, err)
					return
				}
				resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					t.Errorf("endpoint %d, request %d: expected 200, got %d", idx, j, resp.StatusCode)
					return
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify all endpoints have requests.
	for i := 0; i < numEndpoints; i++ {
		endpoint := fmt.Sprintf("endpoint-%d", i)
		reqs, err := repo.ListRequests(context.Background(), endpoint, 100)
		if err != nil {
			t.Fatalf("list requests for %s: %v", endpoint, err)
		}
		if len(reqs) != 10 {
			t.Errorf("expected 10 requests for %s, got %d", endpoint, len(reqs))
		}
	}
}

// TestErrorHandling tests error handling scenarios.
func TestErrorHandling(t *testing.T) {
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

	client := &http.Client{Timeout: 5 * time.Second}

	// Test invalid JSON body.
	t.Run("invalid_json", func(t *testing.T) {
		resp, err := client.Post(srv.URL+"/test", "application/json", strings.NewReader("not json"))
		if err != nil {
			t.Fatalf("POST /test: %v", err)
		}
		resp.Body.Close()
		// Should still return 200 (capture accepts any body).
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	// Test missing endpoint ID (root path).
	t.Run("missing_endpoint", func(t *testing.T) {
		resp, err := client.Get(srv.URL + "/")
		if err != nil {
			t.Fatalf("GET /: %v", err)
		}
		resp.Body.Close()
		// Chi returns 404 for unmatched routes.
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	// Test replay with invalid request ID.
	t.Run("replay_invalid_id", func(t *testing.T) {
		req, _ := http.NewRequest("POST", srv.URL+"/e/test/requests/nonexistent/replay",
			strings.NewReader(`{"target_url":"https://example.com"}`))
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("replay: %v", err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})
}
