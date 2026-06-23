package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// CLI commands:
//   hatch serve                  Start the server (default if no subcommand)
//   hatch capture <url>          Send a request to an endpoint and store it
//   hatch inspect <endpoint>     Fetch requests as JSON
//   hatch search <endpoint>      Search captured traffic
//   hatch replay <request-id>    Replay a request
//   hatch mock set <endpoint>    Configure mock
//   hatch doc generate <endpoint> Output OpenAPI spec

// cliMain parses os.Args and dispatches to the appropriate CLI command.
// Returns true if a CLI command was handled, false if the caller should
// start the server (no subcommand or "serve").
func cliMain() bool {
	if len(os.Args) < 2 {
		return false // no subcommand → start server
	}

	cmd := os.Args[1]
	switch cmd {
	case "serve":
		return false // explicit serve → start server
	case "capture":
		cmdCapture(os.Args[2:])
	case "inspect":
		cmdInspect(os.Args[2:])
	case "search":
		cmdSearch(os.Args[2:])
	case "replay":
		cmdReplay(os.Args[2:])
	case "mock":
		cmdMock(os.Args[2:])
	case "doc":
		cmdDoc(os.Args[2:])
	case "help", "-h", "--help":
		printUsage()
	case "version":
		fmt.Println("hatch v0.1.0")
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
	return true
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: hatch <command> [options]

Commands:
  serve                       Start the server (default)
  capture <url>               Send a request to an endpoint and store it
  inspect <endpoint>          Fetch requests as JSON
  search <endpoint>           Search captured traffic
  replay <request-id>         Replay a request
  mock set <endpoint>         Configure mock response
  doc generate <endpoint>     Output OpenAPI spec
  version                     Print version

Run 'hatch <command> -h' for command-specific help.
`)
}

// serverURL returns the Hatch server URL from HATCH_URL env or default.
func serverURL() string {
	if u := os.Getenv("HATCH_URL"); u != "" {
		return strings.TrimRight(u, "/")
	}
	return "http://localhost:8080"
}

// apiRequest is a helper to make API requests.
func apiRequest(method, path string, body interface{}) ([]byte, int, error) {
	url := serverURL() + path
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("read body: %w", err)
	}

	return data, resp.StatusCode, nil
}

// printJSON pretty-prints JSON data.
func printJSON(data []byte) {
	var buf bytes.Buffer
	if err := json.Indent(&buf, data, "", "  "); err != nil {
		// Not valid JSON, print as-is
		fmt.Println(string(data))
		return
	}
	fmt.Println(buf.String())
}

// cmdCapture handles: hatch capture <url> [-method METHOD] [-body BODY] [-header KEY:VALUE]
func cmdCapture(args []string) {
	fs := flag.NewFlagSet("capture", flag.ExitOnError)
	method := fs.String("method", "POST", "HTTP method")
	body := fs.String("body", "", "Request body (JSON string)")
	var headers multiFlag
	fs.Var(&headers, "header", "Header in KEY:VALUE format (repeatable)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: hatch capture <url> [options]\n\nSend a request to an endpoint and store it.\n\nOptions:\n")
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  hatch capture https://api.example.com/webhook\n")
		fmt.Fprintf(os.Stderr, "  hatch capture /my-endpoint -method POST -body '{\"event\":\"test\"}'\n")
		fmt.Fprintf(os.Stderr, "  hatch capture /ep -header 'Content-Type:application/json' -header 'X-Custom:test'\n")
	}
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Error: URL is required\n\n")
		fs.Usage()
		os.Exit(1)
	}

	url := fs.Arg(0)

	// Parse headers into map.
	headerMap := make(map[string]string)
	for _, h := range headers {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) != 2 {
			fmt.Fprintf(os.Stderr, "Error: invalid header format %q (expected KEY:VALUE)\n", h)
			os.Exit(1)
		}
		headerMap[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}

	// Build the API request body.
	apiBody := map[string]interface{}{
		"method": *method,
		"path":   "/",
	}
	if *body != "" {
		apiBody["body"] = *body
	}
	if len(headerMap) > 0 {
		apiBody["headers"] = headerMap
	}

	// Determine endpoint ID from URL.
	// If it looks like a full URL, extract the path. Otherwise use as-is.
	endpointID := strings.TrimLeft(url, "/")
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		// Extract path from URL.
		if idx := strings.Index(url[8:], "/"); idx != -1 {
			endpointID = url[8+idx+1:]
		} else {
			endpointID = "default"
		}
	}
	endpointID = strings.TrimRight(endpointID, "/")
	if endpointID == "" {
		endpointID = "default"
	}

	path := fmt.Sprintf("/v1/endpoints/%s/requests", endpointID)
	data, status, err := apiRequest("POST", path, apiBody)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if status >= 400 {
		fmt.Fprintf(os.Stderr, "Error (HTTP %d): %s\n", status, string(data))
		os.Exit(1)
	}

	fmt.Println("Request captured successfully:")
	printJSON(data)
}

// cmdInspect handles: hatch inspect <endpoint> [-limit N]
func cmdInspect(args []string) {
	fs := flag.NewFlagSet("inspect", flag.ExitOnError)
	limit := fs.Int("limit", 100, "Maximum number of requests to return")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: hatch inspect <endpoint> [options]\n\nFetch requests as JSON.\n\nOptions:\n")
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  hatch inspect my-webhook\n")
		fmt.Fprintf(os.Stderr, "  hatch inspect my-webhook -limit 10\n")
	}
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Error: endpoint ID is required\n\n")
		fs.Usage()
		os.Exit(1)
	}

	endpointID := fs.Arg(0)
	path := fmt.Sprintf("/v1/endpoints/%s/requests?limit=%d", endpointID, *limit)
	data, status, err := apiRequest("GET", path, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if status >= 400 {
		fmt.Fprintf(os.Stderr, "Error (HTTP %d): %s\n", status, string(data))
		os.Exit(1)
	}

	printJSON(data)
}

// cmdSearch handles: hatch search <endpoint> -query <query> [-limit N]
func cmdSearch(args []string) {
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	query := fs.String("query", "", "Search query")
	limit := fs.Int("limit", 100, "Maximum number of results")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: hatch search <endpoint> [options]\n\nSearch captured traffic.\n\nOptions:\n")
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  hatch search my-webhook -query 'status:500'\n")
		fmt.Fprintf(os.Stderr, "  hatch search my-webhook -query 'POST' -limit 5\n")
	}
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Error: endpoint ID is required\n\n")
		fs.Usage()
		os.Exit(1)
	}

	if *query == "" {
		fmt.Fprintf(os.Stderr, "Error: -query is required\n\n")
		fs.Usage()
		os.Exit(1)
	}

	endpointID := fs.Arg(0)
	path := fmt.Sprintf("/v1/endpoints/%s/requests?q=%s&limit=%d", endpointID, *query, *limit)
	data, status, err := apiRequest("GET", path, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if status >= 400 {
		fmt.Fprintf(os.Stderr, "Error (HTTP %d): %s\n", status, string(data))
		os.Exit(1)
	}

	printJSON(data)
}

// cmdReplay handles: hatch replay <request-id> -endpoint <endpoint> -target <url>
func cmdReplay(args []string) {
	fs := flag.NewFlagSet("replay", flag.ExitOnError)
	endpoint := fs.String("endpoint", "", "Endpoint ID")
	target := fs.String("target", "", "Target URL to replay to")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: hatch replay <request-id> [options]\n\nReplay a captured request.\n\nOptions:\n")
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  hatch replay abc123 -endpoint my-webhook -target https://httpbin.org/post\n")
	}
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Error: request ID is required\n\n")
		fs.Usage()
		os.Exit(1)
	}

	if *endpoint == "" {
		fmt.Fprintf(os.Stderr, "Error: -endpoint is required\n\n")
		fs.Usage()
		os.Exit(1)
	}

	if *target == "" {
		fmt.Fprintf(os.Stderr, "Error: -target is required\n\n")
		fs.Usage()
		os.Exit(1)
	}

	requestID := fs.Arg(0)
	apiBody := map[string]string{
		"target_url": *target,
	}

	path := fmt.Sprintf("/v1/endpoints/%s/requests/%s/replay", *endpoint, requestID)
	data, status, err := apiRequest("POST", path, apiBody)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if status >= 400 {
		fmt.Fprintf(os.Stderr, "Error (HTTP %d): %s\n", status, string(data))
		os.Exit(1)
	}

	fmt.Println("Replay result:")
	printJSON(data)
}

// cmdMock handles: hatch mock set <endpoint> -status <code> [-body BODY] [-header KEY:VALUE]
func cmdMock(args []string) {
	if len(args) < 1 || args[0] != "set" {
		fmt.Fprintf(os.Stderr, "Usage: hatch mock set <endpoint> [options]\n\nConfigure mock response.\n\nSubcommands:\n  set    Set mock configuration\n")
		os.Exit(1)
	}

	fs := flag.NewFlagSet("mock set", flag.ExitOnError)
	statusCode := fs.Int("status", 200, "HTTP status code")
	body := fs.String("body", "", "Response body (string or base64)")
	var headers multiFlag
	fs.Var(&headers, "header", "Response header in KEY:VALUE format (repeatable)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: hatch mock set <endpoint> [options]\n\nConfigure mock response.\n\nOptions:\n")
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  hatch mock set my-webhook -status 200 -body '{\"ok\":true}'\n")
		fmt.Fprintf(os.Stderr, "  hatch mock set my-webhook -status 418 -header 'X-Teapot:true'\n")
	}
	fs.Parse(args[1:]) // skip "set"

	if fs.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Error: endpoint ID is required\n\n")
		fs.Usage()
		os.Exit(1)
	}

	endpointID := fs.Arg(0)

	// Parse headers into map.
	headerMap := make(map[string]string)
	for _, h := range headers {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) != 2 {
			fmt.Fprintf(os.Stderr, "Error: invalid header format %q (expected KEY:VALUE)\n", h)
			os.Exit(1)
		}
		headerMap[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}

	// Build mock config body.
	mockBody := map[string]interface{}{
		"status": *statusCode,
	}
	if *body != "" {
		mockBody["body"] = *body
	}
	if len(headerMap) > 0 {
		mockBody["headers"] = headerMap
	}

	path := fmt.Sprintf("/v1/endpoints/%s/mock", endpointID)
	data, status, err := apiRequest("PUT", path, mockBody)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if status >= 400 {
		fmt.Fprintf(os.Stderr, "Error (HTTP %d): %s\n", status, string(data))
		os.Exit(1)
	}

	fmt.Println("Mock configured successfully:")
	printJSON(data)
}

// cmdDoc handles: hatch doc generate <endpoint>
func cmdDoc(args []string) {
	if len(args) < 1 || args[0] != "generate" {
		fmt.Fprintf(os.Stderr, "Usage: hatch doc generate <endpoint>\n\nOutput OpenAPI spec.\n\nSubcommands:\n  generate   Generate OpenAPI spec for an endpoint\n")
		os.Exit(1)
	}

	fs := flag.NewFlagSet("doc generate", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: hatch doc generate <endpoint>\n\nOutput OpenAPI 3.1 spec for the Hatch API.\n\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  hatch doc generate my-webhook\n")
		fmt.Fprintf(os.Stderr, "  hatch doc generate my-webhook > openapi.json\n")
	}
	fs.Parse(args[1:]) // skip "generate"

	if fs.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Error: endpoint ID is required\n\n")
		fs.Usage()
		os.Exit(1)
	}

	endpointID := fs.Arg(0)
	path := fmt.Sprintf("/v1/endpoints/%s/openapi.json", endpointID)
	data, status, err := apiRequest("GET", path, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if status >= 400 {
		fmt.Fprintf(os.Stderr, "Error (HTTP %d): %s\n", status, string(data))
		os.Exit(1)
	}

	printJSON(data)
}

// multiFlag allows repeated -header flags.
type multiFlag []string

func (m *multiFlag) String() string { return strings.Join(*m, ", ") }
func (m *multiFlag) Set(v string) error {
	*m = append(*m, v)
	return nil
}
