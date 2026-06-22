package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/elfoundation/hatch/internal/store"
	"github.com/go-chi/chi/v5"
)

// fakeRepo implements store.Repository for tests.
type fakeRepo struct {
	endpoints map[string]*store.Endpoint
	requests  []*store.Request
	mocks     map[string]*store.MockConfig
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		endpoints: map[string]*store.Endpoint{},
		mocks:     map[string]*store.MockConfig{},
	}
}

func (f *fakeRepo) CreateEndpoint(_ context.Context, u string) (*store.Endpoint, error) {
	e := &store.Endpoint{ID: u, URL: u, CreatedAt: "t", UpdatedAt: "t"}
	f.endpoints[u] = e
	return e, nil
}
func (f *fakeRepo) GetEndpoint(_ context.Context, id string) (*store.Endpoint, error) {
	e, ok := f.endpoints[id]
	if !ok {
		return nil, errNotFound
	}
	return e, nil
}
func (f *fakeRepo) AppendRequest(_ context.Context, eid string, r *store.Request) error {
	r.ID = "req-" + string(rune(len(f.requests)+'0'))
	r.EndpointID = eid
	f.requests = append(f.requests, r)
	return nil
}
func (f *fakeRepo) GetRequest(_ context.Context, id string) (*store.Request, error) {
	for _, r := range f.requests {
		if r.ID == id {
			return r, nil
		}
	}
	return nil, errNotFound
}
func (f *fakeRepo) ListRequests(_ context.Context, _ string, _ int) ([]*store.Request, error) {
	return f.requests, nil
}
func (f *fakeRepo) GetMock(_ context.Context, endpointID string) (*store.MockConfig, error) {
	m, ok := f.mocks[endpointID]
	if !ok {
		return nil, errNotFound
	}
	return m, nil
}
func (f *fakeRepo) SetMock(_ context.Context, mock *store.MockConfig) error {
	f.mocks[mock.EndpointID] = mock
	return nil
}
func (f *fakeRepo) Close() error { return nil }

var errNotFound = &se{"nf"}

type se struct{ m string }

func (e *se) Error() string { return e.m }

// testRouter creates a chi router with all routes registered using a fake repo.
func testRouter(repo store.Repository) chi.Router {
	r := chi.NewRouter()
	h := New(repo)
	h.RegisterRoutes(r)
	return r
}

func TestHealthz(t *testing.T) {
	r := testRouter(newFakeRepo())
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := strings.TrimSpace(w.Body.String())
	if body != "ok" {
		t.Fatalf("expected 'ok', got %q", body)
	}
}

func TestCaptureRecordsAllVerbs(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE"}
	for _, m := range methods {
		repo := newFakeRepo()
		r := testRouter(repo)

		body := ""
		if m == "POST" || m == "PUT" || m == "PATCH" {
			body = `{"k":"v"}`
		}
		req := httptest.NewRequest(m, "/ep", strings.NewReader(body))
		if m == "GET" {
			req.Header.Set("Accept", "text/html")
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("%s: expected 200, got %d", m, w.Code)
		}
		if len(repo.requests) != 1 {
			t.Errorf("%s: expected 1 request, got %d", m, len(repo.requests))
			continue
		}
		reqCaptured := repo.requests[0]
		if reqCaptured.Method != m {
			t.Errorf("%s: wrong method %s", m, reqCaptured.Method)
		}
	}
}

func TestSSEStreamReceivesEventOnCapture(t *testing.T) {
	repo := newFakeRepo()
	srv := httptest.NewServer(testRouter(repo))
	defer srv.Close()

	// Use separate clients with separate transports to avoid
	// connection-pool contention with the long-lived SSE connection.
	sseClient := &http.Client{
		Transport: &http.Transport{DisableKeepAlives: true},
	}
	captureClient := &http.Client{
		Transport: &http.Transport{DisableKeepAlives: true},
	}

	// Subscribe to SSE with a cancellable context.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sseReq, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/e/ep/events", nil)
	if err != nil {
		t.Fatalf("SSE request create failed: %v", err)
	}
	sseResp, err := sseClient.Do(sseReq)
	if err != nil {
		t.Fatalf("SSE GET failed: %v", err)
	}
	defer sseResp.Body.Close()

	if sseResp.StatusCode != http.StatusOK {
		t.Fatalf("SSE returned %d", sseResp.StatusCode)
	}
	ct := sseResp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "text/event-stream") {
		t.Fatalf("SSE Content-Type is %q, expected text/event-stream", ct)
	}

	// Read SSE body in a goroutine; capture in main goroutine.
	bodyCh := make(chan string, 1)
	go func() {
		buf := make([]byte, 4096)
		n, _ := sseResp.Body.Read(buf)
		bodyCh <- string(buf[:n])
	}()

	// Small sleep to let the SSE subscription register.
	time.Sleep(10 * time.Millisecond)

	// Capture a request on the same server.
	captureResp, err := captureClient.Post(srv.URL+"/ep", "application/json", strings.NewReader(`{"msg":"hello"}`))
	if err != nil {
		t.Fatalf("capture POST failed: %v", err)
	}
	captureResp.Body.Close()
	if captureResp.StatusCode != http.StatusOK {
		t.Fatalf("capture returned %d", captureResp.StatusCode)
	}

	// Wait for the SSE event with a timeout.
	select {
	case body := <-bodyCh:
		if !strings.Contains(body, "data:") {
			t.Fatalf("SSE stream missing 'data:' prefix: %q", body)
		}
		if !strings.Contains(body, "\"method\":\"POST\"") {
			t.Fatalf("SSE stream missing POST method: %q", body)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for SSE event")
	}

	// Cancel context to close the SSE connection.
	cancel()
}

func TestInspectReturnsHTML(t *testing.T) {
	repo := newFakeRepo()
	repo.CreateEndpoint(nil, "ep")
	repo.AppendRequest(nil, "ep", &store.Request{
		Method:  "POST",
		Path:    "/ep/webhook",
		Headers: `{"Content-Type":"application/json"}`,
		Query:   "foo=bar",
		Body:    []byte(`{"msg":"hello"}`),
	})

	r := testRouter(repo)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/e/ep", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Fatalf("expected text/html Content-Type, got %q", ct)
	}

	body := w.Body.String()

	// Should contain the endpoint ID.
	if !strings.Contains(body, "ep") {
		t.Error("missing endpoint ID in HTML")
	}

	// Should contain the request method badge.
	if !strings.Contains(body, "POST") {
		t.Error("missing POST method in HTML")
	}

	// Should contain the request path.
	if !strings.Contains(body, "/ep/webhook") {
		t.Error("missing request path in HTML")
	}

	// Should contain the header JSON.
	if !strings.Contains(body, "Content-Type") {
		t.Error("missing Content-Type header in HTML")
	}

	// Should contain the query string.
	if !strings.Contains(body, "foo=bar") {
		t.Error("missing query string in HTML")
	}

	// Should contain the body content (html/template escapes quotes).
	if !strings.Contains(body, "msg") || !strings.Contains(body, "hello") {
		t.Error("missing request body content in HTML")
	}

	// Should contain the replay button.
	if !strings.Contains(body, "Replay") {
		t.Error("missing Replay button in HTML")
	}

	// Should contain SSE EventSource script.
	if !strings.Contains(body, "EventSource") {
		t.Error("missing EventSource in HTML")
	}
}

func TestInspectEmptyState(t *testing.T) {
	repo := newFakeRepo()
	repo.CreateEndpoint(nil, "new-ep")

	r := testRouter(repo)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/e/new-ep", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	body := w.Body.String()

	// Should show the "waiting for requests" empty state.
	if !strings.Contains(body, "Waiting for requests") {
		t.Error("missing empty state message")
	}

	// Should show the usage hint with the endpoint ID.
	if !strings.Contains(body, "/new-ep") {
		t.Error("missing usage hint with endpoint ID")
	}

	// No request cards should be rendered.
	if strings.Contains(body, `class="request"`) {
		t.Error("unexpected request card in empty state")
	}
}

func TestInspectAutoCreatesEndpoint(t *testing.T) {
	repo := newFakeRepo()

	// The endpoint doesn't exist yet.
	r := testRouter(repo)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/e/auto-ep", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// The endpoint should now exist in the repo.
	ep, err := repo.GetEndpoint(nil, "auto-ep")
	if err != nil {
		t.Fatalf("endpoint was not auto-created: %v", err)
	}
	if ep.ID != "auto-ep" {
		t.Errorf("expected endpoint ID 'auto-ep', got %q", ep.ID)
	}

	body := w.Body.String()
	if !strings.Contains(body, "auto-ep") {
		t.Error("missing endpoint ID in auto-created page")
	}
	if !strings.Contains(body, "Waiting for requests") {
		t.Error("missing empty state for auto-created endpoint")
	}
}

func TestCaptureReturnsJSON200(t *testing.T) {
	r := testRouter(newFakeRepo())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/t", nil))

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Header().Get("Content-Type"), "application/json") {
		t.Error("not json")
	}
	data, _ := io.ReadAll(w.Result().Body)
	w.Result().Body.Close()
	var v map[string]interface{}
	if json.Unmarshal(data, &v) != nil {
		t.Fatal("invalid json")
	}
}
