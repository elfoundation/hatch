package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Version is set at build time via ldflags.
var Version = "dev"

// Exit codes for consistent error handling.
const (
	ExitOK           = 0
	ExitGeneralError = 1
	ExitUsageError   = 2
	ExitNetworkError = 3
	ExitServerError  = 4
	ExitConfigError  = 5
)

// ANSI color codes for terminal output.
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
)

// isTerminal checks if stdout is a terminal for color support.
func isTerminal() bool {
	f := os.Stdout
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// colorize applies color only if terminal supports it.
func colorize(color, text string) string {
	if isTerminal() && os.Getenv("NO_COLOR") == "" {
		return color + text + colorReset
	}
	return text
}

// CLIError represents a structured CLI error with exit code.
type CLIError struct {
	Code    int
	Message string
	Detail  string
}

func (e *CLIError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.Detail)
	}
	return e.Message
}

// Config holds CLI configuration from file and environment.
type Config struct {
	ServerURL string `json:"server_url"`
	Format    string `json:"format"` // json, table, compact
	NoColor   bool   `json:"no_color"`
	Timeout   int    `json:"timeout_seconds"`
}

// LoadConfig reads configuration from file and environment variables.
// Priority: env vars > config file > defaults.
func LoadConfig() (*Config, error) {
	cfg := &Config{
		ServerURL: "http://localhost:8080",
		Format:    "json",
		NoColor:   false,
		Timeout:   30,
	}

	// Try to load config file.
	configPath := getConfigPath()
	if configPath != "" {
		if data, err := os.ReadFile(configPath); err == nil {
			if err := json.Unmarshal(data, cfg); err != nil {
				return nil, &CLIError{Code: ExitConfigError, Message: "invalid config file", Detail: err.Error()}
			}
		}
	}

	// Environment variables override config file.
	if v := os.Getenv("HATCH_URL"); v != "" {
		cfg.ServerURL = strings.TrimRight(v, "/")
	}
	if v := os.Getenv("HATCH_FORMAT"); v != "" {
		cfg.Format = v
	}
	if v := os.Getenv("NO_COLOR"); v != "" {
		cfg.NoColor = true
	}

	return cfg, nil
}

// getConfigPath returns the path to the config file.
func getConfigPath() string {
	// Check XDG_CONFIG_HOME first.
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "hatch", "config.json")
	}

	// Check home directory.
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	// macOS: ~/Library/Application Support/hatch/config.json
	if runtime.GOOS == "darwin" {
		return filepath.Join(home, "Library", "Application Support", "hatch", "config.json")
	}

	// Linux/Unix: ~/.config/hatch/config.json
	return filepath.Join(home, ".config", "hatch", "config.json")
}

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
	case "config":
		cmdConfig(os.Args[2:])
	case "completions":
		cmdCompletions(os.Args[2:])
	case "help", "-h", "--help":
		printUsage()
	case "version", "--version":
		fmt.Printf("hatch %s\n", Version)
	default:
		fmt.Fprintf(os.Stderr, "%s Unknown command: %s\n", colorize(colorRed, "error:"), cmd)
		fmt.Fprintln(os.Stderr)
		printUsage()
		os.Exit(ExitUsageError)
	}
	return true
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `%s - Webhook capture and replay tool

%s hatch <command> [options]

%s
  serve                       Start the server (default)
  capture <url>               Send a request to an endpoint and store it
  inspect <endpoint>          Fetch requests as JSON
  search <endpoint>           Search captured traffic
  replay <request-id>         Replay a request
  mock set <endpoint>         Configure mock response
  doc generate <endpoint>     Output OpenAPI spec
  config                      Manage configuration
  completions                 Generate shell completions
  version                     Print version

Run 'hatch <command> -h' for command-specific help.
Run 'hatch completions' to set up shell completions.
`,
		colorize(colorCyan, "hatch"),
		colorize(colorYellow, "Usage:"),
		colorize(colorGreen, "Commands:"),
	)
}

// serverURL returns the Hatch server URL from config or environment.
func serverURL() string {
	cfg, err := LoadConfig()
	if err != nil {
		// Fallback to env var or default.
		if u := os.Getenv("HATCH_URL"); u != "" {
			return strings.TrimRight(u, "/")
		}
		return "http://localhost:8080"
	}
	return cfg.ServerURL
}

// apiRequest is a helper to make API requests.
func apiRequest(method, path string, body interface{}) ([]byte, int, error) {
	url := serverURL() + path
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, 0, &CLIError{Code: ExitGeneralError, Message: "failed to marshal request body", Detail: err.Error()}
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, 0, &CLIError{Code: ExitGeneralError, Message: "failed to create request", Detail: err.Error()}
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			return nil, 0, &CLIError{
				Code:    ExitNetworkError,
				Message: "cannot connect to Hatch server",
				Detail:  fmt.Sprintf("is the server running at %s?", serverURL()),
			}
		}
		return nil, 0, &CLIError{Code: ExitNetworkError, Message: "request failed", Detail: err.Error()}
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, &CLIError{Code: ExitGeneralError, Message: "failed to read response", Detail: err.Error()}
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

// printSuccess prints a success message with green checkmark.
func printSuccess(msg string) {
	fmt.Fprintf(os.Stderr, "%s %s\n", colorize(colorGreen, "✓"), msg)
}

// printError prints an error message with red X.
func printError(msg string) {
	fmt.Fprintf(os.Stderr, "%s %s\n", colorize(colorRed, "✗"), msg)
}

// printWarning prints a warning message with yellow !.
func printWarning(msg string) {
	fmt.Fprintf(os.Stderr, "%s %s\n", colorize(colorYellow, "!"), msg)
}

// handleError processes CLI errors and exits with appropriate code.
func handleError(err error) {
	if cliErr, ok := err.(*CLIError); ok {
		fmt.Fprintf(os.Stderr, "%s %s\n", colorize(colorRed, "error:"), cliErr.Message)
		if cliErr.Detail != "" {
			fmt.Fprintf(os.Stderr, "  %s\n", colorize(colorGray, cliErr.Detail))
		}
		os.Exit(cliErr.Code)
	}
	printError(err.Error())
	os.Exit(ExitGeneralError)
}

// cmdCapture handles: hatch capture <url> [-method METHOD] [-body BODY] [-header KEY:VALUE]
func cmdCapture(args []string) {
	fs := flag.NewFlagSet("capture", flag.ContinueOnError)
	method := fs.String("method", "POST", "HTTP method")
	body := fs.String("body", "", "Request body (JSON string)")
	output := fs.String("output", "", "Output format: json, table, compact")
	var headers multiFlag
	fs.Var(&headers, "header", "Header in KEY:VALUE format (repeatable)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `%s Send a request to an endpoint and store it.

%s hatch capture <url> [options]

%s
  -method string    HTTP method (default "POST")
  -body string      Request body (JSON string)
  -header string    Header in KEY:VALUE format (repeatable)
  -output string    Output format: json, table, compact (default json)

%s
  hatch capture https://api.example.com/webhook
  hatch capture /my-endpoint -method POST -body '{"event":"test"}'
  hatch capture /ep -header 'Content-Type:application/json' -header 'X-Custom:test'
`,
			colorize(colorCyan, "Capture a webhook request."),
			colorize(colorYellow, "Usage:"),
			colorize(colorGreen, "Options:"),
			colorize(colorGreen, "Examples:"),
		)
	}
	if err := fs.Parse(args); err != nil {
		os.Exit(ExitUsageError)
	}

	if fs.NArg() < 1 {
		printError("URL is required")
		fs.Usage()
		os.Exit(ExitUsageError)
	}

	url := fs.Arg(0)

	// Parse headers into map.
	headerMap := make(map[string]string)
	for _, h := range headers {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) != 2 {
			fmt.Fprintf(os.Stderr, "%s invalid header format %q (expected KEY:VALUE)\n",
				colorize(colorRed, "error:"), h)
			os.Exit(ExitUsageError)
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
	endpointID := extractEndpointID(url)

	path := fmt.Sprintf("/v1/endpoints/%s/requests", endpointID)
	data, status, err := apiRequest("POST", path, apiBody)
	if err != nil {
		handleError(err)
	}
	if status >= 400 {
		fmt.Fprintf(os.Stderr, "%s HTTP %d\n", colorize(colorRed, "error:"), status)
		if len(data) > 0 {
			var errResp struct {
				Error string `json:"error"`
			}
			if json.Unmarshal(data, &errResp) == nil && errResp.Error != "" {
				fmt.Fprintf(os.Stderr, "  %s\n", errResp.Error)
			}
		}
		os.Exit(ExitServerError)
	}

	printSuccess("Request captured successfully")
	if *output == "compact" {
		fmt.Println(string(data))
	} else {
		printJSON(data)
	}
}

// cmdInspect handles: hatch inspect <endpoint> [-limit N] [-output FORMAT]
func cmdInspect(args []string) {
	fs := flag.NewFlagSet("inspect", flag.ContinueOnError)
	limit := fs.Int("limit", 100, "Maximum number of requests to return")
	output := fs.String("output", "", "Output format: json, table, compact")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `%s Fetch captured requests for an endpoint.

%s hatch inspect <endpoint> [options]

%s
  -limit int        Maximum number of requests to return (default 100)
  -output string    Output format: json, table, compact (default json)

%s
  hatch inspect my-webhook
  hatch inspect my-webhook -limit 10
  hatch inspect my-webhook -output table
`,
			colorize(colorCyan, "Inspect captured requests."),
			colorize(colorYellow, "Usage:"),
			colorize(colorGreen, "Options:"),
			colorize(colorGreen, "Examples:"),
		)
	}
	if err := fs.Parse(args); err != nil {
		os.Exit(ExitUsageError)
	}

	if fs.NArg() < 1 {
		printError("endpoint ID is required")
		fs.Usage()
		os.Exit(ExitUsageError)
	}

	endpointID := fs.Arg(0)
	path := fmt.Sprintf("/v1/endpoints/%s/requests?limit=%d", endpointID, *limit)
	data, status, err := apiRequest("GET", path, nil)
	if err != nil {
		handleError(err)
	}
	if status >= 400 {
		fmt.Fprintf(os.Stderr, "%s HTTP %d\n", colorize(colorRed, "error:"), status)
		os.Exit(ExitServerError)
	}

	if *output == "compact" {
		fmt.Println(string(data))
	} else {
		printJSON(data)
	}
}

// cmdSearch handles: hatch search <endpoint> -query <query> [-limit N]
func cmdSearch(args []string) {
	fs := flag.NewFlagSet("search", flag.ContinueOnError)
	query := fs.String("query", "", "Search query")
	limit := fs.Int("limit", 100, "Maximum number of results")
	output := fs.String("output", "", "Output format: json, table, compact")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `%s Search captured traffic.

%s hatch search <endpoint> [options]

%s
  -query string     Search query (required)
  -limit int        Maximum number of results (default 100)
  -output string    Output format: json, table, compact (default json)

%s
  hatch search my-webhook -query 'status:500'
  hatch search my-webhook -query 'POST' -limit 5
`,
			colorize(colorCyan, "Search captured traffic."),
			colorize(colorYellow, "Usage:"),
			colorize(colorGreen, "Options:"),
			colorize(colorGreen, "Examples:"),
		)
	}
	if err := fs.Parse(args); err != nil {
		os.Exit(ExitUsageError)
	}

	if fs.NArg() < 1 {
		printError("endpoint ID is required")
		fs.Usage()
		os.Exit(ExitUsageError)
	}

	if *query == "" {
		printError("-query is required")
		fs.Usage()
		os.Exit(ExitUsageError)
	}

	endpointID := fs.Arg(0)
	path := fmt.Sprintf("/v1/endpoints/%s/requests?q=%s&limit=%d", endpointID, *query, *limit)
	data, status, err := apiRequest("GET", path, nil)
	if err != nil {
		handleError(err)
	}
	if status >= 400 {
		fmt.Fprintf(os.Stderr, "%s HTTP %d\n", colorize(colorRed, "error:"), status)
		os.Exit(ExitServerError)
	}

	if *output == "compact" {
		fmt.Println(string(data))
	} else {
		printJSON(data)
	}
}

// cmdReplay handles: hatch replay <request-id> -endpoint <endpoint> -target <url>
func cmdReplay(args []string) {
	fs := flag.NewFlagSet("replay", flag.ContinueOnError)
	endpoint := fs.String("endpoint", "", "Endpoint ID")
	target := fs.String("target", "", "Target URL to replay to")
	output := fs.String("output", "", "Output format: json, table, compact")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `%s Replay a captured request.

%s hatch replay <request-id> [options]

%s
  -endpoint string    Endpoint ID (required)
  -target string      Target URL to replay to (required)
  -output string      Output format: json, table, compact (default json)

%s
  hatch replay abc123 -endpoint my-webhook -target https://httpbin.org/post
`,
			colorize(colorCyan, "Replay a captured request."),
			colorize(colorYellow, "Usage:"),
			colorize(colorGreen, "Options:"),
			colorize(colorGreen, "Examples:"),
		)
	}
	if err := fs.Parse(args); err != nil {
		os.Exit(ExitUsageError)
	}

	if fs.NArg() < 1 {
		printError("request ID is required")
		fs.Usage()
		os.Exit(ExitUsageError)
	}

	if *endpoint == "" {
		printError("-endpoint is required")
		fs.Usage()
		os.Exit(ExitUsageError)
	}

	if *target == "" {
		printError("-target is required")
		fs.Usage()
		os.Exit(ExitUsageError)
	}

	requestID := fs.Arg(0)
	apiBody := map[string]string{
		"target_url": *target,
	}

	path := fmt.Sprintf("/v1/endpoints/%s/requests/%s/replay", *endpoint, requestID)
	data, status, err := apiRequest("POST", path, apiBody)
	if err != nil {
		handleError(err)
	}
	if status >= 400 {
		fmt.Fprintf(os.Stderr, "%s HTTP %d\n", colorize(colorRed, "error:"), status)
		os.Exit(ExitServerError)
	}

	printSuccess("Replay completed")
	if *output == "compact" {
		fmt.Println(string(data))
	} else {
		printJSON(data)
	}
}

// cmdMock handles: hatch mock set <endpoint> -status <code> [-body BODY] [-header KEY:VALUE]
func cmdMock(args []string) {
	if len(args) < 1 || args[0] != "set" {
		fmt.Fprintf(os.Stderr, `%s Configure mock responses.

%s hatch mock set <endpoint> [options]

%s
  set    Set mock configuration
`,
			colorize(colorCyan, "Mock configuration."),
			colorize(colorYellow, "Usage:"),
			colorize(colorGreen, "Subcommands:"),
		)
		os.Exit(ExitUsageError)
	}

	fs := flag.NewFlagSet("mock set", flag.ContinueOnError)
	statusCode := fs.Int("status", 200, "HTTP status code")
	body := fs.String("body", "", "Response body (string or base64)")
	output := fs.String("output", "", "Output format: json, table, compact")
	var headers multiFlag
	fs.Var(&headers, "header", "Response header in KEY:VALUE format (repeatable)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `%s Configure a mock response for an endpoint.

%s hatch mock set <endpoint> [options]

%s
  -status int        HTTP status code (default 200)
  -body string       Response body (string or base64)
  -header string     Response header in KEY:VALUE format (repeatable)
  -output string     Output format: json, table, compact (default json)

%s
  hatch mock set my-webhook -status 200 -body '{"ok":true}'
  hatch mock set my-webhook -status 418 -header 'X-Teapot:true'
`,
			colorize(colorCyan, "Set mock response."),
			colorize(colorYellow, "Usage:"),
			colorize(colorGreen, "Options:"),
			colorize(colorGreen, "Examples:"),
		)
	}
	if err := fs.Parse(args[1:]); err != nil {
		os.Exit(ExitUsageError)
	}

	if fs.NArg() < 1 {
		printError("endpoint ID is required")
		fs.Usage()
		os.Exit(ExitUsageError)
	}

	endpointID := fs.Arg(0)

	// Parse headers into map.
	headerMap := make(map[string]string)
	for _, h := range headers {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) != 2 {
			fmt.Fprintf(os.Stderr, "%s invalid header format %q (expected KEY:VALUE)\n",
				colorize(colorRed, "error:"), h)
			os.Exit(ExitUsageError)
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
		handleError(err)
	}
	if status >= 400 {
		fmt.Fprintf(os.Stderr, "%s HTTP %d\n", colorize(colorRed, "error:"), status)
		os.Exit(ExitServerError)
	}

	printSuccess("Mock configured successfully")
	if *output == "compact" {
		fmt.Println(string(data))
	} else {
		printJSON(data)
	}
}

// cmdDoc handles: hatch doc generate <endpoint>
func cmdDoc(args []string) {
	if len(args) < 1 || args[0] != "generate" {
		fmt.Fprintf(os.Stderr, `%s Generate API documentation.

%s hatch doc generate <endpoint>

%s
  generate   Generate OpenAPI spec for an endpoint
`,
			colorize(colorCyan, "Documentation generation."),
			colorize(colorYellow, "Usage:"),
			colorize(colorGreen, "Subcommands:"),
		)
		os.Exit(ExitUsageError)
	}

	fs := flag.NewFlagSet("doc generate", flag.ContinueOnError)
	output := fs.String("output", "", "Output format: json, compact")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `%s Generate OpenAPI 3.1 spec for an endpoint.

%s hatch doc generate <endpoint>

%s
  -output string    Output format: json, compact (default json)

%s
  hatch doc generate my-webhook
  hatch doc generate my-webhook > openapi.json
`,
			colorize(colorCyan, "Generate OpenAPI spec."),
			colorize(colorYellow, "Usage:"),
			colorize(colorGreen, "Options:"),
			colorize(colorGreen, "Examples:"),
		)
	}
	if err := fs.Parse(args[1:]); err != nil {
		os.Exit(ExitUsageError)
	}

	if fs.NArg() < 1 {
		printError("endpoint ID is required")
		fs.Usage()
		os.Exit(ExitUsageError)
	}

	endpointID := fs.Arg(0)
	path := fmt.Sprintf("/v1/endpoints/%s/openapi.json", endpointID)
	data, status, err := apiRequest("GET", path, nil)
	if err != nil {
		handleError(err)
	}
	if status >= 400 {
		fmt.Fprintf(os.Stderr, "%s HTTP %d\n", colorize(colorRed, "error:"), status)
		os.Exit(ExitServerError)
	}

	if *output == "compact" {
		fmt.Println(string(data))
	} else {
		printJSON(data)
	}
}

// cmdConfig handles: hatch config [show|set|get|init]
func cmdConfig(args []string) {
	if len(args) < 1 {
		printConfigUsage()
		os.Exit(ExitUsageError)
	}

	subcmd := args[0]
	switch subcmd {
	case "show":
		cmdConfigShow()
	case "set":
		cmdConfigSet(args[1:])
	case "get":
		cmdConfigGet(args[1:])
	case "init":
		cmdConfigInit()
	case "help", "-h", "--help":
		printConfigUsage()
	default:
		fmt.Fprintf(os.Stderr, "%s Unknown config subcommand: %s\n", colorize(colorRed, "error:"), subcmd)
		printConfigUsage()
		os.Exit(ExitUsageError)
	}
}

func printConfigUsage() {
	fmt.Fprintf(os.Stderr, `%s Manage hatch configuration.

%s hatch config <command>

%s
  show              Show current configuration
  set <key> <value> Set a configuration value
  get <key>         Get a configuration value
  init              Create default config file

%s
  server_url        Hatch server URL (default: http://localhost:8080)
  format            Output format: json, table, compact (default: json)
  no_color          Disable colored output (default: false)

%s
  hatch config show
  hatch config set server_url http://hatch.example.com:9000
  hatch config set format table
  hatch config get server_url
  hatch config init
`,
		colorize(colorCyan, "Configuration management."),
		colorize(colorYellow, "Usage:"),
		colorize(colorGreen, "Subcommands:"),
		colorize(colorGreen, "Keys:"),
		colorize(colorGreen, "Examples:"),
	)
}

func cmdConfigShow() {
	cfg, err := LoadConfig()
	if err != nil {
		handleError(err)
	}

	fmt.Fprintf(os.Stderr, "%s Current configuration:\n\n", colorize(colorCyan, "Config:"))
	fmt.Fprintf(os.Stderr, "  %s %s\n", colorize(colorGray, "server_url:"), cfg.ServerURL)
	fmt.Fprintf(os.Stderr, "  %s %s\n", colorize(colorGray, "format:"), cfg.Format)
	fmt.Fprintf(os.Stderr, "  %s %v\n", colorize(colorGray, "no_color:"), cfg.NoColor)
	fmt.Fprintf(os.Stderr, "  %s %ds\n", colorize(colorGray, "timeout:"), cfg.Timeout)
	fmt.Fprintf(os.Stderr, "\n  %s %s\n", colorize(colorGray, "config_file:"), getConfigPath())
}

func cmdConfigSet(args []string) {
	if len(args) < 2 {
		printError("key and value are required")
		fmt.Fprintf(os.Stderr, "Usage: hatch config set <key> <value>\n")
		os.Exit(ExitUsageError)
	}

	key, value := args[0], args[1]

	cfg, err := LoadConfig()
	if err != nil {
		handleError(err)
	}

	switch key {
	case "server_url":
		cfg.ServerURL = strings.TrimRight(value, "/")
	case "format":
		if value != "json" && value != "table" && value != "compact" {
			printError("invalid format (must be: json, table, compact)")
			os.Exit(ExitUsageError)
		}
		cfg.Format = value
	case "no_color":
		cfg.NoColor = value == "true" || value == "1"
	case "timeout":
		var timeout int
		if _, err := fmt.Sscanf(value, "%d", &timeout); err != nil || timeout < 1 {
			printError("invalid timeout (must be a positive integer)")
			os.Exit(ExitUsageError)
		}
		cfg.Timeout = timeout
	default:
		fmt.Fprintf(os.Stderr, "%s Unknown key: %s\n", colorize(colorRed, "error:"), key)
		fmt.Fprintf(os.Stderr, "Valid keys: server_url, format, no_color, timeout\n")
		os.Exit(ExitUsageError)
	}

	// Save config.
	configPath := getConfigPath()
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		handleError(&CLIError{Code: ExitConfigError, Message: "failed to create config directory", Detail: err.Error()})
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		handleError(&CLIError{Code: ExitConfigError, Message: "failed to marshal config", Detail: err.Error()})
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		handleError(&CLIError{Code: ExitConfigError, Message: "failed to write config file", Detail: err.Error()})
	}

	printSuccess(fmt.Sprintf("Set %s = %s", key, value))
}

func cmdConfigGet(args []string) {
	if len(args) < 1 {
		printError("key is required")
		fmt.Fprintf(os.Stderr, "Usage: hatch config get <key>\n")
		os.Exit(ExitUsageError)
	}

	key := args[0]

	cfg, err := LoadConfig()
	if err != nil {
		handleError(err)
	}

	switch key {
	case "server_url":
		fmt.Println(cfg.ServerURL)
	case "format":
		fmt.Println(cfg.Format)
	case "no_color":
		fmt.Println(cfg.NoColor)
	case "timeout":
		fmt.Println(cfg.Timeout)
	default:
		fmt.Fprintf(os.Stderr, "%s Unknown key: %s\n", colorize(colorRed, "error:"), key)
		os.Exit(ExitUsageError)
	}
}

func cmdConfigInit() {
	configPath := getConfigPath()
	if _, err := os.Stat(configPath); err == nil {
		printWarning("Config file already exists")
		fmt.Fprintf(os.Stderr, "  %s\n", configPath)
		return
	}

	cfg := &Config{
		ServerURL: "http://localhost:8080",
		Format:    "json",
		NoColor:   false,
		Timeout:   30,
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		handleError(&CLIError{Code: ExitConfigError, Message: "failed to create config", Detail: err.Error()})
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		handleError(&CLIError{Code: ExitConfigError, Message: "failed to create config directory", Detail: err.Error()})
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		handleError(&CLIError{Code: ExitConfigError, Message: "failed to write config file", Detail: err.Error()})
	}

	printSuccess(fmt.Sprintf("Created config file: %s", configPath))
}

// cmdCompletions handles shell completion generation.
func cmdCompletions(args []string) {
	shell := "bash"
	if len(args) > 0 {
		shell = args[0]
	}

	switch shell {
	case "bash":
		printBashCompletions()
	case "zsh":
		printZshCompletions()
	case "fish":
		printFishCompletions()
	case "powershell":
		printPowershellCompletions()
	case "help", "-h", "--help":
		printCompletionsUsage()
	default:
		fmt.Fprintf(os.Stderr, "%s Unknown shell: %s\n", colorize(colorRed, "error:"), shell)
		printCompletionsUsage()
		os.Exit(ExitUsageError)
	}
}

func printCompletionsUsage() {
	fmt.Fprintf(os.Stderr, `%s Generate shell completions for hatch.

%s hatch completions [shell]

%s
  bash          Bash completions (default)
  zsh           Zsh completions
  fish          Fish completions
  powershell    PowerShell completions

%s
  # Bash: add to ~/.bashrc
  eval "$(hatch completions bash)"

  # Zsh: add to ~/.zshrc
  eval "$(hatch completions zsh)"

  # Fish: save to completions directory
  hatch completions fish > ~/.config/fish/completions/hatch.fish

  # PowerShell: add to profile
  hatch completions powershell | Out-String | Invoke-Expression
`,
		colorize(colorCyan, "Shell completions."),
		colorize(colorYellow, "Usage:"),
		colorize(colorGreen, "Shells:"),
		colorize(colorGreen, "Setup:"),
	)
}

func printBashCompletions() {
	fmt.Print(`# Bash completions for hatch
# Usage: eval "$(hatch completions bash)"

_hatch_completions() {
    local cur prev words cword
    _init_completion || return

    # Commands
    local commands="serve capture inspect search replay mock doc config completions help version"

    # Global flags
    local global_flags="--help --version -h"

    if [[ ${cword} -eq 1 ]]; then
        COMPREPLY=($(compgen -W "${commands} ${global_flags}" -- "${cur}"))
        return
    fi

    local command="${words[1]}"

    case "${command}" in
        capture)
            local flags="--method --body --header --output -h"
            COMPREPLY=($(compgen -W "${flags}" -- "${cur}"))
            ;;
        inspect)
            local flags="--limit --output -h"
            COMPREPLY=($(compgen -W "${flags}" -- "${cur}"))
            ;;
        search)
            local flags="--query --limit --output -h"
            COMPREPLY=($(compgen -W "${flags}" -- "${cur}"))
            ;;
        replay)
            local flags="--endpoint --target --output -h"
            COMPREPLY=($(compgen -W "${flags}" -- "${cur}"))
            ;;
        mock)
            if [[ ${cword} -eq 2 ]]; then
                COMPREPLY=($(compgen -W "set help" -- "${cur}"))
            elif [[ "${words[2]}" == "set" ]]; then
                local flags="--status --body --header --output -h"
                COMPREPLY=($(compgen -W "${flags}" -- "${cur}"))
            fi
            ;;
        doc)
            if [[ ${cword} -eq 2 ]]; then
                COMPREPLY=($(compgen -W "generate help" -- "${cur}"))
            fi
            ;;
        config)
            if [[ ${cword} -eq 2 ]]; then
                local subcommands="show set get init help"
                COMPREPLY=($(compgen -W "${subcommands}" -- "${cur}"))
            elif [[ "${words[2]}" == "set" ]]; then
                local keys="server_url format no_color timeout"
                if [[ ${cword} -eq 4 ]]; then
                    case "${words[3]}" in
                        format)
                            COMPREPLY=($(compgen -W "json table compact" -- "${cur}"))
                            ;;
                        no_color)
                            COMPREPLY=($(compgen -W "true false" -- "${cur}"))
                            ;;
                    esac
                elif [[ ${cword} -eq 3 ]]; then
                    COMPREPLY=($(compgen -W "${keys}" -- "${cur}"))
                fi
            elif [[ "${words[2]}" == "get" ]]; then
                local keys="server_url format no_color timeout"
                COMPREPLY=($(compgen -W "${keys}" -- "${cur}"))
            fi
            ;;
        completions)
            local shells="bash zsh fish powershell"
            COMPREPLY=($(compgen -W "${shells}" -- "${cur}"))
            ;;
    esac
}

complete -F _hatch_completions hatch
`)
}

func printZshCompletions() {
	fmt.Print(`# Zsh completions for hatch
# Usage: eval "$(hatch completions zsh)"

#compdef hatch

_hatch() {
    local -a commands
    commands=(
        'serve:Start the server (default)'
        'capture:Send a request to an endpoint and store it'
        'inspect:Fetch captured requests for an endpoint'
        'search:Search captured traffic'
        'replay:Replay a captured request'
        'mock:Configure mock responses'
        'doc:Generate API documentation'
        'config:Manage configuration'
        'completions:Generate shell completions'
        'help:Show help'
        'version:Print version'
    )

    local -a capture_flags
    capture_flags=(
        '--method[HTTP method]:method:(GET POST PUT PATCH DELETE)'
        '--body[Request body]:body:'
        '--header[Header in KEY:VALUE format]:header:'
        '--output[Output format]:format:(json table compact)'
    )

    local -a inspect_flags
    inspect_flags=(
        '--limit[Maximum number of requests]:limit:'
        '--output[Output format]:format:(json table compact)'
    )

    local -a search_flags
    search_flags=(
        '--query[Search query]:query:'
        '--limit[Maximum number of results]:limit:'
        '--output[Output format]:format:(json table compact)'
    )

    local -a replay_flags
    replay_flags=(
        '--endpoint[Endpoint ID]:endpoint:'
        '--target[Target URL to replay to]:target:'
        '--output[Output format]:format:(json table compact)'
    )

    local -a mock_flags
    mock_flags=(
        '--status[HTTP status code]:status:'
        '--body[Response body]:body:'
        '--header[Response header in KEY:VALUE format]:header:'
        '--output[Output format]:format:(json table compact)'
    )

    local -a config_subcommands
    config_subcommands=(
        'show:Show current configuration'
        'set:Set a configuration value'
        'get:Get a configuration value'
        'init:Create default config file'
        'help:Show config help'
    )

    local -a config_set_keys
    config_set_keys=(
        'server_url:Hatch server URL'
        'format:Output format'
        'no_color:Disable colored output'
        'timeout:Request timeout in seconds'
    )

    local -a completions_shells
    completions_shells=(
        'bash:Bash completions'
        'zsh:Zsh completions'
        'fish:Fish completions'
        'powershell:PowerShell completions'
    )

    _arguments -C \
        '1:command:->command' \
        '*::arg:->args'

    case $state in
        command)
            _describe 'command' commands
            ;;
        args)
            case ${words[1]} in
                capture)
                    _arguments $capture_flags
                    ;;
                inspect)
                    _arguments $inspect_flags
                    ;;
                search)
                    _arguments $search_flags
                    ;;
                replay)
                    _arguments $replay_flags
                    ;;
                mock)
                    _arguments '1:subcommand:(set)' $mock_flags
                    ;;
                config)
                    _arguments '1:subcommand:->config_subcmd' '*::arg:->config_args'
                    ;;
                completions)
                    _describe 'shell' completions_shells
                    ;;
            esac
            ;;
        config_subcmd)
            _describe 'subcommand' config_subcommands
            ;;
        config_args)
            case ${words[1]} in
                set)
                    _arguments '1:key:->config_key' '2:value:'
                    ;;
                get)
                    _arguments '1:key:->config_key'
                    ;;
            esac
            ;;
        config_key)
            _describe 'key' config_set_keys
            ;;
    esac
}

_hatch "$@"
`)
}

func printFishCompletions() {
	fmt.Print(`# Fish completions for hatch
# Usage: hatch completions fish > ~/.config/fish/completions/hatch.fish

# Helper function
function __hatch_needs_command
    set cmd (commandline -opc)
    if test (count $cmd) -eq 1
        return 0
    end
    return 1
end

function __hatch_using_command
    set cmd (commandline -opc)
    if test (count $cmd) -gt 1
        if test $argv[1] = $cmd[2]
            return 0
        end
    end
    return 1
end

function __hatch_complete_with_subcommands
    set cmd (commandline -opc)
    if test (count $cmd) -eq 2
        switch $cmd[2]
            case mock
                echo -e "set\tConfigure mock response"
            case doc
                echo -e "generate\tGenerate OpenAPI spec"
            case config
                echo -e "show\tShow current configuration"
                echo -e "set\tSet a configuration value"
                echo -e "get\tGet a configuration value"
                echo -e "init\tCreate default config file"
            case completions
                echo -e "bash\tBash completions"
                echo -e "zsh\tZsh completions"
                echo -e "fish\tFish completions"
                echo -e "powershell\tPowerShell completions"
        end
    end
end

# Main command completions
complete -c hatch -n __hatch_needs_command -a serve -d 'Start the server (default)'
complete -c hatch -n __hatch_needs_command -a capture -d 'Send a request to an endpoint and store it'
complete -c hatch -n __hatch_needs_command -a inspect -d 'Fetch captured requests for an endpoint'
complete -c hatch -n __hatch_needs_command -a search -d 'Search captured traffic'
complete -c hatch -n __hatch_needs_command -a replay -d 'Replay a captured request'
complete -c hatch -n __hatch_needs_command -a mock -d 'Configure mock responses'
complete -c hatch -n __hatch_needs_command -a doc -d 'Generate API documentation'
complete -c hatch -n __hatch_needs_command -a config -d 'Manage configuration'
complete -c hatch -n __hatch_needs_command -a completions -d 'Generate shell completions'
complete -c hatch -n __hatch_needs_command -a help -d 'Show help'
complete -c hatch -n __hatch_needs_command -a version -d 'Print version'

# Subcommand completions
complete -c hatch -n __hatch_complete_with_subcommands

# capture flags
complete -c hatch -n __hatch_using_command capture -l method -d 'HTTP method' -r
complete -c hatch -n __hatch_using_command capture -l body -d 'Request body (JSON string)' -r
complete -c hatch -n __hatch_using_command capture -l header -d 'Header in KEY:VALUE format (repeatable)' -r
complete -c hatch -n __hatch_using_command capture -l output -d 'Output format: json, table, compact' -r -a 'json table compact'

# inspect flags
complete -c hatch -n __hatch_using_command inspect -l limit -d 'Maximum number of requests to return' -r
complete -c hatch -n __hatch_using_command inspect -l output -d 'Output format: json, table, compact' -r -a 'json table compact'

# search flags
complete -c hatch -n __hatch_using_command search -l query -d 'Search query' -r
complete -c hatch -n __hatch_using_command search -l limit -d 'Maximum number of results' -r
complete -c hatch -n __hatch_using_command search -l output -d 'Output format: json, table, compact' -r -a 'json table compact'

# replay flags
complete -c hatch -n __hatch_using_command replay -l endpoint -d 'Endpoint ID' -r
complete -c hatch -n __hatch_using_command replay -l target -d 'Target URL to replay to' -r
complete -c hatch -n __hatch_using_command replay -l output -d 'Output format: json, table, compact' -r -a 'json table compact'

# mock set flags
complete -c hatch -n __hatch_using_command mock -l status -d 'HTTP status code' -r
complete -c hatch -n __hatch_using_command mock -l body -d 'Response body (string or base64)' -r
complete -c hatch -n __hatch_using_command mock -l header -d 'Response header in KEY:VALUE format (repeatable)' -r
complete -c hatch -n __hatch_using_command mock -l output -d 'Output format: json, table, compact' -r -a 'json table compact'

# config set keys
complete -c hatch -n __hatch_using_command config -l server_url -d 'Hatch server URL' -r
complete -c hatch -n __hatch_using_command config -l format -d 'Output format' -r -a 'json table compact'
complete -c hatch -n __hatch_using_command config -l no_color -d 'Disable colored output' -r -a 'true false'
complete -c hatch -n __hatch_using_command config -l timeout -d 'Request timeout in seconds' -r
`)
}

func printPowershellCompletions() {
	fmt.Print(`# PowerShell completions for hatch
# Usage: hatch completions powershell | Out-String | Invoke-Expression

Register-ArgumentCompleter -Native -CommandName 'hatch' -ScriptBlock {
    param($wordToComplete, $commandAst, $cursorPosition)

    $commandElements = $commandAst.CommandElements
    $command = @($commandElements)[0].CommandElement.Text

    # Commands
    $commands = @(
        'serve'
        'capture'
        'inspect'
        'search'
        'replay'
        'mock'
        'doc'
        'config'
        'completions'
        'help'
        'version'
    )

    # If we're typing the command itself
    if ($commandElements.Count -eq 1) {
        $commands | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
            [System.Management.Automation.CompletionResult]::new(
                $_,
                $_,
                'ParameterValue',
                $_
            )
        }
        return
    }

    # Handle subcommands
    $subCommand = $commandElements[1].CommandElement.Text

    switch ($subCommand) {
        'capture' {
            $flags = @('--method', '--body', '--header', '--output')
            $flags | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
                [System.Management.Automation.CompletionResult]::new(
                    $_,
                    $_,
                    'ParameterValue',
                    $_
                )
            }
        }
        'inspect' {
            $flags = @('--limit', '--output')
            $flags | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
                [System.Management.Automation.CompletionResult]::new(
                    $_,
                    $_,
                    'ParameterValue',
                    $_
                )
            }
        }
        'search' {
            $flags = @('--query', '--limit', '--output')
            $flags | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
                [System.Management.Automation.CompletionResult]::new(
                    $_,
                    $_,
                    'ParameterValue',
                    $_
                )
            }
        }
        'replay' {
            $flags = @('--endpoint', '--target', '--output')
            $flags | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
                [System.Management.Automation.CompletionResult]::new(
                    $_,
                    $_,
                    'ParameterValue',
                    $_
                )
            }
        }
        'mock' {
            if ($commandElements.Count -eq 2) {
                @('set') | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
                    [System.Management.Automation.CompletionResult]::new(
                        $_,
                        $_,
                        'ParameterValue',
                        $_
                    )
                }
            } else {
                $flags = @('--status', '--body', '--header', '--output')
                $flags | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
                    [System.Management.Automation.CompletionResult]::new(
                        $_,
                        $_,
                        'ParameterValue',
                        $_
                    )
                }
            }
        }
        'doc' {
            if ($commandElements.Count -eq 2) {
                @('generate') | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
                    [System.Management.Automation.CompletionResult]::new(
                        $_,
                        $_,
                        'ParameterValue',
                        $_
                    )
                }
            }
        }
        'config' {
            if ($commandElements.Count -eq 2) {
                @('show', 'set', 'get', 'init') | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
                    [System.Management.Automation.CompletionResult]::new(
                        $_,
                        $_,
                        'ParameterValue',
                        $_
                    )
                }
            } elseif ($commandElements.Count -eq 3 -and $commandElements[2].CommandElement.Text -eq 'set') {
                @('server_url', 'format', 'no_color', 'timeout') | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
                    [System.Management.Automation.CompletionResult]::new(
                        $_,
                        $_,
                        'ParameterValue',
                        $_
                    )
                }
            }
        }
        'completions' {
            @('bash', 'zsh', 'fish', 'powershell') | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
                [System.Management.Automation.CompletionResult]::new(
                    $_,
                    $_,
                    'ParameterValue',
                    $_
                )
            }
        }
    }
}
`)
}

// extractEndpointID extracts the endpoint ID from a URL or path.
func extractEndpointID(url string) string {
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
	return endpointID
}

// multiFlag allows repeated -header flags.
type multiFlag []string

func (m *multiFlag) String() string { return strings.Join(*m, ", ") }
func (m *multiFlag) Set(v string) error {
	*m = append(*m, v)
	return nil
}
