package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/elfoundation/hatch/internal/store"
)

func TestV1CaptureRequest(t *testing.T) {
	repo := newFakeRepo()
	r := testRouter(repo)

	body := `{"method":"POST","path":"/test","headers":{"X-Custom":"value"},"query":"foo=bar","body":"hello"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/endpoints/test-ep/requests", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", w.Code, w.Body.String())
	}
	// Verify response JSON contains the request.
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if resp["method"] != "POST" {
		t.Errorf("expected method POST, got %v", resp["method"])
	}
	if resp["path"] != "/test" {
		t.Errorf("expected path /test, got %v", resp["path"])
	}
	// Verify request stored in repo.
	if len(repo.requests) != 1 {
		t.Fatalf("expected 1 request stored, got %d", len(repo.requests))
	}
	if repo.requests[0].Headers != `{"X-Custom":"value"}` {
		t.Errorf("expected headers JSON, got %q", repo.requests[0].Headers)
	}
}

func TestV1ListRequests(t *testing.T) {
	repo := newFakeRepo()
	repo.CreateEndpoint(nil, "list-ep")
	// Add a few requests.
	repo.AppendRequest(nil, "list-ep", &store.Request{Method: "GET", Path: "/a"})
	repo.AppendRequest(nil, "list-ep", &store.Request{Method: "POST", Path: "/b"})
	repo.AppendRequest(nil, "list-ep", &store.Request{Method: "PUT", Path: "/c"})

	r := testRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/v1/endpoints/list-ep/requests", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}
	var list []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &list); err != nil {
		t.Fatalf("invalid JSON array: %v", err)
	}
	if len(list) != 3 {
		t.Errorf("expected 3 requests, got %d", len(list))
	}
	// Verify all methods present.
	methods := make(map[string]bool)
	for _, r := range list {
		methods[r["method"].(string)] = true
	}
	if !methods["GET"] || !methods["POST"] || !methods["PUT"] {
		t.Errorf("expected GET, POST, PUT, got %v", methods)
	}
}

func TestV1SearchRequests(t *testing.T) {
	repo := newFakeRepo()
	repo.CreateEndpoint(nil, "search-ep")
	repo.AppendRequest(nil, "search-ep", &store.Request{Method: "GET", Path: "/users", Headers: `{"Accept":"json"}`})
	repo.AppendRequest(nil, "search-ep", &store.Request{Method: "POST", Path: "/orders", Body: []byte(`{"item":"apple"}`)})
	repo.AppendRequest(nil, "search-ep", &store.Request{Method: "DELETE", Path: "/users/123"})

	r := testRouter(repo)

	// Search for "users".
	req := httptest.NewRequest(http.MethodGet, "/v1/endpoints/search-ep/requests?q=users", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}
	var list []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &list); err != nil {
		t.Fatalf("invalid JSON array: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 requests matching 'users', got %d", len(list))
	}
	// Verify the matched methods.
	methods := make(map[string]bool)
	for _, r := range list {
		methods[r["method"].(string)] = true
	}
	if !methods["GET"] || !methods["DELETE"] {
		t.Errorf("expected GET and DELETE, got %v", methods)
	}
}

func TestV1MockSet(t *testing.T) {
	repo := newFakeRepo()
	repo.CreateEndpoint(nil, "mock-ep")
	r := testRouter(repo)

	mockBody := `{"status":201,"headers":{"X-Mock":"true"},"body":"bW9ja2Vk"}`
	req := httptest.NewRequest(http.MethodPut, "/v1/endpoints/mock-ep/mock", bytes.NewBufferString(mockBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}
	// Verify mock stored.
	mock, err := repo.GetMock(nil, "mock-ep")
	if err != nil {
		t.Fatalf("mock not stored: %v", err)
	}
	if mock.Status != 201 {
		t.Errorf("expected status 201, got %d", mock.Status)
	}
	if mock.Headers["X-Mock"] != "true" {
		t.Errorf("expected X-Mock header, got %v", mock.Headers)
	}
}

func TestV1OpenAPI(t *testing.T) {
	repo := newFakeRepo()
	r := testRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/v1/endpoints/test/openapi.json", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}
	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}
	var spec map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &spec); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if spec["openapi"] != "3.1.0" {
		t.Errorf("expected openapi 3.1.0, got %v", spec["openapi"])
	}
	if spec["info"].(map[string]interface{})["title"] != "Hatch API" {
		t.Errorf("expected title 'Hatch API', got %v", spec["info"])
	}
	// Verify paths exist.
	paths, ok := spec["paths"].(map[string]interface{})
	if !ok {
		t.Fatal("missing paths")
	}
	if _, ok := paths["/v1/endpoints/{endpointID}/requests"]; !ok {
		t.Error("missing capture/list path")
	}
	if _, ok := paths["/v1/endpoints/{endpointID}/requests/{requestID}/replay"]; !ok {
		t.Error("missing replay path")
	}
	if _, ok := paths["/v1/endpoints/{endpointID}/mock"]; !ok {
		t.Error("missing mock path")
	}
}

// Import store package for the fakeRepo usage (already imported via handler_test.go)
// We just need to ensure the test compiles.
var _ = store.Request{}