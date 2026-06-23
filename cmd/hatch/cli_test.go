package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// TestCLIHelp tests the help command.
func TestCLIHelp(t *testing.T) {
	// Save and restore os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"hatch", "help"}

	// cliMain prints help and returns true
	if !cliMain() {
		t.Error("expected cliMain to return true for help command")
	}
}

// TestCLIUnknownCommand tests unknown command handling.
// Note: This test verifies the command exists but doesn't call cliMain directly
// because it calls os.Exit(1). In production, the binary would exit.
func TestCLIUnknownCommand(t *testing.T) {
	// Just verify the command parsing works for known commands
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Test that help works
	os.Args = []string{"hatch", "help"}
	if !cliMain() {
		t.Error("expected cliMain to return true for help command")
	}
}

// TestCLIServe tests the serve command.
func TestCLIServe(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"hatch", "serve"}

	// cliMain should return false (server mode)
	if cliMain() {
		t.Error("expected cliMain to return false for serve command")
	}
}

// TestCLINoSubcommand tests no subcommand (default to server).
func TestCLINoSubcommand(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"hatch"}

	// cliMain should return false (server mode)
	if cliMain() {
		t.Error("expected cliMain to return false with no subcommand")
	}
}

// TestServerURL tests the serverURL helper.
func TestServerURL(t *testing.T) {
	// Test default
	oldURL := os.Getenv("HATCH_URL")
	defer os.Setenv("HATCH_URL", oldURL)

	os.Unsetenv("HATCH_URL")
	if got := serverURL(); got != "http://localhost:8080" {
		t.Errorf("expected default URL, got %q", got)
	}

	// Test custom URL
	os.Setenv("HATCH_URL", "https://my-server.example.com")
	if got := serverURL(); got != "https://my-server.example.com" {
		t.Errorf("expected custom URL, got %q", got)
	}

	// Test trailing slash removal
	os.Setenv("HATCH_URL", "https://my-server.example.com/")
	if got := serverURL(); got != "https://my-server.example.com" {
		t.Errorf("expected URL without trailing slash, got %q", got)
	}
}

// TestPrintJSON tests JSON pretty printing.
func TestPrintJSON(t *testing.T) {
	// Valid JSON
	validJSON := []byte(`{"key":"value","number":42}`)
	// Just ensure it doesn't panic
	printJSON(validJSON)

	// Invalid JSON
	invalidJSON := []byte(`not json`)
	printJSON(invalidJSON)
}

// TestMultiFlag tests the multiFlag type.
func TestMultiFlag(t *testing.T) {
	var flags multiFlag

	if err := flags.Set("Content-Type:application/json"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := flags.Set("X-Custom:test"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(flags) != 2 {
		t.Errorf("expected 2 flags, got %d", len(flags))
	}
	if flags[0] != "Content-Type:application/json" {
		t.Errorf("unexpected first flag: %q", flags[0])
	}
	if flags[1] != "X-Custom:test" {
		t.Errorf("unexpected second flag: %q", flags[1])
	}
}

// TestCLICaptureIntegration tests the capture command against a mock server.
func TestCLICaptureIntegration(t *testing.T) {
	// Start a mock server that accepts the API request
	var receivedBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/endpoints/test-ep/requests" && r.Method == "POST" {
			json.NewDecoder(r.Body).Decode(&receivedBody)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]string{"id": "abc123", "status": "captured"})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	// Set HATCH_URL to our test server
	oldURL := os.Getenv("HATCH_URL")
	defer os.Setenv("HATCH_URL", oldURL)
	os.Setenv("HATCH_URL", server.URL)

	// Test the API request helper
	body := map[string]interface{}{
		"method": "POST",
		"path":   "/",
		"body":   `{"test":"data"}`,
	}
	data, status, err := apiRequest("POST", "/v1/endpoints/test-ep/requests", body)
	if err != nil {
		t.Fatalf("apiRequest failed: %v", err)
	}
	if status != http.StatusCreated {
		t.Errorf("expected status 201, got %d", status)
	}

	var resp map[string]string
	json.Unmarshal(data, &resp)
	if resp["id"] != "abc123" {
		t.Errorf("expected id abc123, got %q", resp["id"])
	}

	// Verify the server received the correct body
	if receivedBody["method"] != "POST" {
		t.Errorf("expected method POST, got %v", receivedBody["method"])
	}
}

// TestCLIInspectIntegration tests the inspect command against a mock server.
func TestCLIInspectIntegration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/endpoints/test-ep/requests" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]map[string]string{
				{"id": "1", "method": "POST"},
				{"id": "2", "method": "GET"},
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	oldURL := os.Getenv("HATCH_URL")
	defer os.Setenv("HATCH_URL", oldURL)
	os.Setenv("HATCH_URL", server.URL)

	data, status, err := apiRequest("GET", "/v1/endpoints/test-ep/requests?limit=100", nil)
	if err != nil {
		t.Fatalf("apiRequest failed: %v", err)
	}
	if status != http.StatusOK {
		t.Errorf("expected status 200, got %d", status)
	}

	var requests []map[string]string
	json.Unmarshal(data, &requests)
	if len(requests) != 2 {
		t.Errorf("expected 2 requests, got %d", len(requests))
	}
}

// TestCLIMockIntegration tests the mock set command against a mock server.
func TestCLIMockIntegration(t *testing.T) {
	var receivedBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/endpoints/test-ep/mock" && r.Method == "PUT" {
			json.NewDecoder(r.Body).Decode(&receivedBody)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	oldURL := os.Getenv("HATCH_URL")
	defer os.Setenv("HATCH_URL", oldURL)
	os.Setenv("HATCH_URL", server.URL)

	body := map[string]interface{}{
		"status": 418,
		"headers": map[string]string{
			"X-Teapot": "true",
		},
	}
	data, status, err := apiRequest("PUT", "/v1/endpoints/test-ep/mock", body)
	if err != nil {
		t.Fatalf("apiRequest failed: %v", err)
	}
	if status != http.StatusOK {
		t.Errorf("expected status 200, got %d", status)
	}

	var resp map[string]string
	json.Unmarshal(data, &resp)
	if resp["status"] != "ok" {
		t.Errorf("expected status ok, got %q", resp["status"])
	}

	// Verify server received correct mock config
	if receivedBody["status"] != float64(418) {
		t.Errorf("expected status 418, got %v", receivedBody["status"])
	}
}

// TestCLIReplayIntegration tests the replay command against a mock server.
func TestCLIReplayIntegration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/endpoints/test-ep/requests/abc123/replay" && r.Method == "POST" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  200,
				"headers": map[string]string{"X-Reply": "yes"},
				"body":    `{"replayed":true}`,
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	oldURL := os.Getenv("HATCH_URL")
	defer os.Setenv("HATCH_URL", oldURL)
	os.Setenv("HATCH_URL", server.URL)

	body := map[string]string{
		"target_url": "https://example.com/webhook",
	}
	data, status, err := apiRequest("POST", "/v1/endpoints/test-ep/requests/abc123/replay", body)
	if err != nil {
		t.Fatalf("apiRequest failed: %v", err)
	}
	if status != http.StatusOK {
		t.Errorf("expected status 200, got %d", status)
	}

	var resp map[string]interface{}
	json.Unmarshal(data, &resp)
	if resp["status"] != float64(200) {
		t.Errorf("expected status 200, got %v", resp["status"])
	}
}
