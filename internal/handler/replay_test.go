package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/elfoundation/hatch/internal/store"
	"github.com/elfoundation/hatch/internal/testutil"
)

func TestReplayMissingTargetURL(t *testing.T) {
	repo := testutil.NewFakeRepository()
	repo.CreateEndpoint(nil, "ep")
	repo.AppendRequest(nil, "ep", &store.Request{Method: "POST", Path: "/webhook", Headers: "{}"})
	reqID := repo.Requests[0].ID

	r := testRouter(repo)
	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/e/ep/requests/"+reqID+"/replay", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if !strings.Contains(resp["error"], "target_url") {
		t.Errorf("expected target_url error, got %v", resp)
	}
}

func TestReplayInvalidScheme(t *testing.T) {
	repo := testutil.NewFakeRepository()
	repo.CreateEndpoint(nil, "ep")
	repo.AppendRequest(nil, "ep", &store.Request{Method: "GET", Path: "/", Headers: "{}"})
	reqID := repo.Requests[0].ID

	r := testRouter(repo)
	body := `{"target_url": "ftp://bad.scheme/"}`
	req := httptest.NewRequest(http.MethodPost, "/e/ep/requests/"+reqID+"/replay", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestReplaySSRFBlocksLocalhost(t *testing.T) {
	repo := testutil.NewFakeRepository()
	repo.CreateEndpoint(nil, "ep")
	repo.AppendRequest(nil, "ep", &store.Request{Method: "GET", Path: "/", Headers: "{}"})
	reqID := repo.Requests[0].ID

	r := testRouter(repo)
	body := `{"target_url": "http://localhost:8080/"}`
	req := httptest.NewRequest(http.MethodPost, "/e/ep/requests/"+reqID+"/replay", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for localhost, got %d", w.Code)
	}
}

func TestReplaySSRFBlocksPrivate(t *testing.T) {
	repo := testutil.NewFakeRepository()
	repo.CreateEndpoint(nil, "ep")
	repo.AppendRequest(nil, "ep", &store.Request{Method: "GET", Path: "/", Headers: "{}"})
	reqID := repo.Requests[0].ID

	r := testRouter(repo)
	for _, addr := range []string{
		"http://10.0.0.1:80/",
		"http://172.16.0.1:80/",
		"http://192.168.1.1:80/",
		"http://127.0.0.1:80/",
		"http://[::1]:80/",
		"http://0.0.0.0:80/",
	} {
		body := `{"target_url": "` + addr + `"}`
		req := httptest.NewRequest(http.MethodPost, "/e/ep/requests/"+reqID+"/replay", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Errorf("expected 403 for %s, got %d", addr, w.Code)
		}
	}
}

func TestReplayNotFound(t *testing.T) {
	repo := testutil.NewFakeRepository()
	r := testRouter(repo)
	body := `{"target_url": "https://example.com/"}`
	req := httptest.NewRequest(http.MethodPost, "/e/ep/requests/nonexistent/replay", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestReplaySuccess(t *testing.T) {
	// Start a local sink server that echoes back what it receives.
	sink := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Sink", "yes")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"received": true}`))
	}))
	defer sink.Close()

	repo := testutil.NewFakeRepository()
	repo.CreateEndpoint(nil, "ep")
	repo.AppendRequest(nil, "ep", &store.Request{
		Method:  "POST",
		Path:    "/webhook",
		Headers: `{"Content-Type":"application/json","X-Custom":"val"}`,
		Query:   "foo=bar",
		Body:    []byte(`{"msg":"hello"}`),
	})
	reqID := repo.Requests[0].ID

	// Set env to allow private replay since httptest server is on loopback.
	t.Setenv("HATCH_ALLOW_PRIVATE_REPLAY", "true")

	r := testRouter(repo)
	body := `{"target_url": "` + sink.URL + `"}`
	req := httptest.NewRequest(http.MethodPost, "/e/ep/requests/"+reqID+"/replay", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp replayResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Status != 201 {
		t.Errorf("expected status 201, got %d", resp.Status)
	}
	if resp.Headers["X-Sink"] != "yes" {
		t.Errorf("expected X-Sink header, got %v", resp.Headers)
	}
	if resp.Body != `{"received": true}` {
		t.Errorf("expected body, got %q", resp.Body)
	}
}

func TestReplayE2ECaptureThenReplay(t *testing.T) {
	// E2E: capture a request, then replay it to a sink, verify the sink received it.

	// Step 1: Start a sink.
	sink := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"method": r.Method,
			"path":   r.URL.Path,
			"query":  r.URL.RawQuery,
			"body":   "received",
		})
	}))
	defer sink.Close()

	// Step 2: Capture a request.
	repo := testutil.NewFakeRepository()
	rt := testRouter(repo)

	captureReq := httptest.NewRequest("POST", "/e2e-test", strings.NewReader(`{"order":"1"}`))
	captureReq.Header.Set("X-Test", "hello")
	captureW := httptest.NewRecorder()
	rt.ServeHTTP(captureW, captureReq)

	if len(repo.Requests) != 1 {
		t.Fatalf("expected 1 captured request, got %d", len(repo.Requests))
	}
	captured := repo.Requests[0]
	if captured.Method != "POST" {
		t.Errorf("captured method: %s", captured.Method)
	}

	// Step 3: Replay it to the sink.
	t.Setenv("HATCH_ALLOW_PRIVATE_REPLAY", "true")

	replayBody := `{"target_url": "` + sink.URL + `"}`
	replayReq := httptest.NewRequest("POST", "/e/e2e-test/requests/"+captured.ID+"/replay", strings.NewReader(replayBody))
	replayReq.Header.Set("Content-Type", "application/json")
	replayW := httptest.NewRecorder()
	rt.ServeHTTP(replayW, replayReq)

	if replayW.Code != http.StatusOK {
		t.Fatalf("replay failed with %d: %s", replayW.Code, replayW.Body.String())
	}

	var resp replayResponse
	json.NewDecoder(replayW.Body).Decode(&resp)
	if resp.Status != 200 {
		t.Errorf("replay response status: %d", resp.Status)
	}
}

func TestIsPrivateAddr(t *testing.T) {
	tests := []struct {
		addr string
		want bool
	}{
		{"localhost:8080", true},
		{"127.0.0.1:9090", true},
		{"[::1]:80", true},
		{"10.0.0.5:443", true},
		{"172.16.0.1:80", true},
		{"192.168.1.1:3000", true},
		{"0.0.0.0:8080", true},
		{"example.com:443", false},
		{"93.184.216.34:80", false},
		{"8.8.8.8:53", false},
	}
	for _, tc := range tests {
		got := isPrivateAddr(tc.addr)
		if got != tc.want {
			t.Errorf("isPrivateAddr(%q) = %v, want %v", tc.addr, tc.want, got)
		}
	}
}
