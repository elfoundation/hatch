# Hatch CLI Reference

The `hatch` CLI provides a powerful interface for interacting with Hatch servers. Use it to capture requests, inspect traffic, search logs, replay requests, configure mocks, and generate API documentation.

## Installation

### Pre-built Binaries (Recommended)

Download the latest binary for your platform from the [Releases page](https://github.com/elfoundation/hatch/releases):

- **Linux (x64)**: `hatch-linux-amd64`
- **Linux (ARM64)**: `hatch-linux-arm64`
- **macOS (Intel)**: `hatch-darwin-amd64`
- **macOS (Apple Silicon)**: `hatch-darwin-arm64`
- **Windows (x64)**: `hatch-windows-amd64.exe`

After downloading:

```bash
# Linux/macOS
chmod +x hatch-*
sudo mv hatch-* /usr/local/bin/hatch

# Windows (PowerShell)
Rename-Item hatch-windows-amd64.exe hatch.exe
# Move to a directory in your PATH
```

### Build from Source

Requires Go 1.25 or later:

```bash
git clone https://github.com/elfoundation/hatch.git
cd hatch
CGO_ENABLED=0 go build -o hatch ./cmd/hatch
sudo mv hatch /usr/local/bin/
```

### Docker

```bash
docker pull ghcr.io/elfoundation/hatch:latest
docker run --rm -it ghcr.io/elfoundation/hatch:latest --help
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `HATCH_URL` | Hatch server URL | `http://localhost:8080` |

### Global Options

All commands support these flags:

- `-h, --help`: Show help for the command
- `version`: Print version information

## Commands

### `hatch serve`

Start the Hatch server. This is the default command when no subcommand is specified.

```bash
# Start server on default port (8080)
hatch serve

# Start with custom port (use environment variable)
PORT=9090 hatch serve
```

### `hatch capture <url>`

Send a request to an endpoint and store it in Hatch.

**Options:**

- `-method METHOD`: HTTP method (default: `POST`)
- `-body BODY`: Request body as JSON string
- `-header KEY:VALUE`: Request header (repeatable)

**Examples:**

```bash
# Capture a simple POST request
hatch capture https://api.example.com/webhook

# Capture with custom method and body
hatch capture /my-endpoint -method PUT -body '{"status":"active"}'

# Capture with headers
hatch capture /api/events \
  -header 'Content-Type:application/json' \
  -header 'X-Webhook-Secret:abc123' \
  -body '{"event":"user.created","data":{"id":123}}'
```

**URL Handling:**

- Full URLs: `https://api.example.com/webhook` → endpoint ID derived from path
- Relative paths: `/my-endpoint` → used directly as endpoint ID
- Empty/`/` → defaults to `default` endpoint

### `hatch inspect <endpoint>`

Fetch captured requests as JSON.

**Options:**

- `-limit N`: Maximum number of requests to return (default: 100)

**Examples:**

```bash
# Inspect all requests for an endpoint
hatch inspect my-webhook

# Limit to 10 most recent requests
hatch inspect my-webhook -limit 10

# Pretty-print JSON output
hatch inspect my-webhook | jq .
```

**Output Format:**

```json
[
  {
    "id": "abc123",
    "method": "POST",
    "path": "/webhook",
    "headers": {"Content-Type": "application/json"},
    "body": "{\"event\":\"test\"}",
    "status": 200,
    "created_at": "2024-01-15T10:30:00Z"
  }
]
```

### `hatch search <endpoint>`

Search captured traffic with query syntax.

**Options:**

- `-query QUERY`: Search query (required)
- `-limit N`: Maximum number of results (default: 100)

**Query Syntax:**

| Query | Description |
|-------|-------------|
| `status:500` | Filter by HTTP status code |
| `POST` | Filter by HTTP method |
| `user.created` | Search in request body |
| `Content-Type:application/json` | Filter by header |

**Examples:**

```bash
# Find all 500 errors
hatch search my-webhook -query 'status:500'

# Find POST requests
hatch search my-webhook -query 'POST'

# Search for specific content in body
hatch search my-webhook -query 'user.created'

# Combine queries
hatch search my-webhook -query 'status:500 POST'
```

### `hatch replay <request-id>`

Replay a previously captured request to a target URL.

**Options:**

- `-endpoint ENDPOINT`: Source endpoint ID (required)
- `-target URL`: Target URL to replay to (required)

**Examples:**

```bash
# Replay a request
hatch replay abc123 -endpoint my-webhook -target https://httpbin.org/post

# Replay to a local development server
hatch replay def456 -endpoint api-calls -target http://localhost:3000/webhook

# Replay to a different environment
hatch replay ghi789 -endpoint production-webhook -target https://staging.example.com/webhook
```

**Use Cases:**

- Test webhook handlers in development
- Reproduce bugs by replaying failed requests
- Load testing with captured traffic patterns
- Migrating traffic between environments

### `hatch mock set <endpoint>`

Configure mock responses for an endpoint.

**Options:**

- `-status CODE`: HTTP status code (default: 200)
- `-body BODY`: Response body (string or base64)
- `-header KEY:VALUE`: Response header (repeatable)

**Examples:**

```bash
# Return 200 with JSON body
hatch mock set my-webhook -status 200 -body '{"ok":true}'

# Return error response
hatch mock set my-webhook -status 500 -body '{"error":"internal error"}'

# Return with custom headers
hatch mock set my-webhook \
  -status 200 \
  -header 'X-RateLimit-Remaining:100' \
  -header 'Cache-Control:no-cache'

# Simulate rate limiting
hatch mock set my-webhook \
  -status 429 \
  -header 'Retry-After:60' \
  -body '{"error":"rate limited"}'
```

**Mock Behavior:**

- Once configured, all requests to the endpoint return the mock response
- Mock configuration persists until server restart or explicit reconfiguration
- Original request capture continues alongside mock responses

### `hatch doc generate <endpoint>`

Generate OpenAPI 3.1 specification for an endpoint based on captured traffic.

**Examples:**

```bash
# Generate OpenAPI spec
hatch doc generate my-webhook

# Save to file
hatch doc generate my-webhook > openapi.json

# Use with Swagger UI
hatch doc generate my-webhook > openapi.json
docker run -p 8081:8080 -e SWAGGER_JSON=/openapi.json -v $(pwd):/usr/share/nginx/html/swagger swaggerapi/swagger-ui
```

**Output:**

Generates a complete OpenAPI 3.1 JSON specification including:

- All captured request/response patterns
- Schema definitions inferred from traffic
- Example values from real requests
- Endpoint documentation

## Common Workflows

### Development Environment Setup

```bash
# 1. Start Hatch server
hatch serve &

# 2. Capture incoming webhooks
hatch capture /api/webhooks -method POST -body '{"test":true}'

# 3. Inspect captured traffic
hatch inspect api/webhooks

# 4. Configure mock for frontend development
hatch mock set api/webhooks -status 200 -body '{"status":"ok"}'
```

### Bug Reproduction

```bash
# 1. Find the failing request
hatch search api/payments -query 'status:500'

# 2. Get request details
hatch inspect api/payments -limit 5

# 3. Replay to debug
hatch replay <request-id> -endpoint api/payments -target http://localhost:8080/debug
```

### API Documentation Generation

```bash
# 1. Capture representative traffic
for endpoint in users posts comments; do
  curl -X POST http://hatch:8080/$endpoint -d '{"sample":"data"}'
done

# 2. Generate OpenAPI specs
hatch doc generate users > users-api.json
hatch doc generate posts > posts-api.json

# 3. Combine into single spec (using jq)
jq -s '.[0] * .[1]' users-api.json posts-api.json > combined-api.json
```

### Load Testing Preparation

```bash
# 1. Capture production traffic patterns
hatch inspect api/critical-path -limit 1000 > traffic.json

# 2. Extract request patterns
jq -r '.[] | "\(.method) \(.path)"' traffic.json | sort | uniq -c | sort -rn

# 3. Replay to load testing tool
hatch inspect api/critical-path -limit 100 | \
  jq -r '.[] | "hatch replay \(.id) -endpoint api/critical-path -target https://loadtest.example.com"'
```

## Troubleshooting

### Connection Issues

**Problem:** `Error: request failed: connection refused`

**Solution:** Ensure the Hatch server is running:

```bash
# Check if server is running
curl http://localhost:8080/healthz

# Start server if not running
hatch serve
```

### Authentication Errors

**Problem:** `Error (HTTP 401): unauthorized`

**Solution:** Check if your Hatch instance requires authentication:

```bash
# For local development, ensure no auth is configured
# For production, check HATCH_AUTH_TOKEN environment variable
```

### Request Not Found

**Problem:** `Error (HTTP 404): endpoint not found`

**Solution:** Verify the endpoint ID exists:

```bash
# List available endpoints
curl http://localhost:8080/v1/endpoints

# Check if endpoint has captured requests
hatch inspect <endpoint-id> -limit 1
```

### JSON Parsing Errors

**Problem:** `Error: invalid character '}' looking for beginning of value`

**Solution:** Ensure request body is valid JSON:

```bash
# Use proper JSON escaping
hatch capture /api -body '{"key":"value"}'

# Or use a file
hatch capture /api -body @request.json
```

### Permission Denied on Binary

**Problem:** `Permission denied: ./hatch`

**Solution:** Make the binary executable:

```bash
chmod +x hatch
# Or move to a directory in PATH
sudo mv hatch /usr/local/bin/
```

## Environment Variables Reference

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `HATCH_URL` | Hatch server URL | `http://localhost:8080` | `https://hatch.example.com` |
| `PORT` | Server listening port | `8080` | `9090` |

## Exit Codes

| Code | Description |
|------|-------------|
| 0 | Success |
| 1 | General error |
| 2 | Invalid command or arguments |
| 3 | Network error |
| 4 | Authentication error |
| 5 | Resource not found |

## Getting Help

- **Command help:** `hatch <command> -h`
- **General help:** `hatch --help`
- **Version:** `hatch version`
- **Issues:** [GitHub Issues](https://github.com/elfoundation/hatch/issues)
- **Documentation:** [docs/](../README.md)

## Examples Repository

For more examples, see the [examples/](../../examples/) directory or visit our [documentation site](https://hatch.sh/examples).