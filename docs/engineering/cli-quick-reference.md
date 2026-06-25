# Hatch CLI Quick Reference

## Installation

```bash
# Download binary (Linux example)
curl -L https://github.com/elfoundation/hatch/releases/latest/download/hatch-linux-amd64 -o hatch
chmod +x hatch
sudo mv hatch /usr/local/bin/
```

## Commands

| Command | Description | Example |
|---------|-------------|---------|
| `hatch serve` | Start server | `hatch serve` |
| `hatch capture <url>` | Capture request | `hatch capture /api/webhook` |
| `hatch inspect <endpoint>` | View requests | `hatch inspect my-webhook` |
| `hatch search <endpoint>` | Search traffic | `hatch search my-webhook -query 'status:500'` |
| `hatch replay <id>` | Replay request | `hatch replay abc123 -endpoint api -target http://localhost:3000` |
| `hatch mock set <endpoint>` | Configure mock | `hatch mock set api -status 200 -body '{}'` |
| `hatch doc generate <endpoint>` | Generate OpenAPI | `hatch doc generate api > openapi.json` |
| `hatch version` | Print version | `hatch version` |

## Common Flags

| Flag | Description | Example |
|------|-------------|---------|
| `-method METHOD` | HTTP method | `hatch capture /api -method PUT` |
| `-body BODY` | Request body | `hatch capture /api -body '{"key":"value"}'` |
| `-header KEY:VALUE` | Add header | `hatch capture /api -header 'Auth:token'` |
| `-limit N` | Limit results | `hatch inspect api -limit 10` |
| `-query QUERY` | Search query | `hatch search api -query 'status:500'` |
| `-endpoint ENDPOINT` | Source endpoint | `hatch replay abc -endpoint api` |
| `-target URL` | Target URL | `hatch replay abc -target http://localhost:3000` |
| `-status CODE` | HTTP status | `hatch mock set api -status 200` |

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `HATCH_URL` | Server URL | `http://localhost:8080` |
| `PORT` | Server port | `8080` |

## Quick Examples

### Capture a Webhook

```bash
hatch capture /webhooks/stripe \
  -method POST \
  -header 'Stripe-Signature:whsec_test' \
  -body '{"type":"payment_intent.succeeded"}'
```

### Debug Failed Requests

```bash
# Find 500 errors
hatch search api -query 'status:500'

# Get details
hatch inspect api -limit 5

# Replay to debug
hatch replay <id> -endpoint api -target http://localhost:8080/debug
```

### Mock API for Frontend

```bash
hatch mock set api/users \
  -status 200 \
  -header 'Content-Type:application/json' \
  -body '[{"id":1,"name":"John"}]'
```

### Generate API Docs

```bash
hatch doc generate api/users > users-api.json
hatch doc generate api/posts > posts-api.json
```

## Query Syntax

| Query | Description |
|-------|-------------|
| `status:500` | Filter by status code |
| `POST` | Filter by HTTP method |
| `user.created` | Search in body |
| `Content-Type:application/json` | Filter by header |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Invalid arguments |
| 3 | Network error |
| 4 | Auth error |
| 5 | Not found |

## Troubleshooting

```bash
# Check server health
curl http://localhost:8080/healthz

# Verify binary works
hatch version

# Test connection
hatch inspect default -limit 1
```

For detailed troubleshooting, see [cli-troubleshooting.md](cli-troubleshooting.md).