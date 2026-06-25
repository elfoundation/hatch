# El Foundation

Engineering repository for El Foundation.

## About

El Foundation builds institutions that outlast their founders. We create technology and organizations that compound in value over time. This repository is the source of truth for our engineering work.

## Getting Started

1. Read the [company charter](docs/company/charter.md) to understand why we exist.
2. Read the [operating model](docs/company/operating-model.md) to understand how decisions are made.
3. Read [CONTRIBUTING.md](CONTRIBUTING.md) before making any changes.
4. Read [docs/engineering/local-dev.md](docs/engineering/local-dev.md) for the day-to-day workflow.
5. Read [docs/engineering/hatch-architecture.md](docs/engineering/hatch-architecture.md) for the component map.

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

For details on the release process and how to create new releases, see [Release Process](docs/engineering/release-process.md).

### Build from Source

Requires Go 1.25 or later:

```bash
git clone https://github.com/elfoundation/hatch.git
cd hatch
CGO_ENABLED=0 go build -o hatch ./cmd/hatch
sudo mv hatch /usr/local/bin/
```

## Repository Layout

```
├── .github/workflows/    # CI/CD definitions
├── cmd/hatch/            # Server entrypoint (Go binary)
├── docs/
│   ├── company/          # Founding documents (charter, org, etc.)
│   ├── engineering/      # Engineering standards, architecture, local dev
│   └── adrs/             # Architecture Decision Records
├── examples/             # Usage examples and integration guides
├── internal/             # Go packages (handler, store, ...)
├── Dockerfile            # Multi-stage static binary build (golang → scratch)
├── docker-compose.yml    # Local stack with optional Caddy sidecar
├── Caddyfile             # TLS reverse proxy for the demo host
└── go.mod                # Go module definition
```

## Hatch — Deploy in one command

Hatch is a self-hostable HTTP request inspector + mocker. Ship it to any VPS with Docker.

**[Read why we are building Hatch →](https://hatch.sh/blog/why-we-are-building-hatch)**

### Quick start (local dev, no HTTPS)

```bash
docker compose up --build
# Hatch UI: http://localhost:8080
# Health check: http://localhost:8080/healthz
```

Or run the binary directly:

```bash
go run ./cmd/hatch
# Health check: http://localhost:8080/healthz
```

### Production (with HTTPS via Caddy)

```bash
# Set your domain name
cp .env.example .env
# Edit HATCH_HOSTNAME in .env to your real domain

# Start Hatch + Caddy (auto-issues Let's Encrypt cert)
docker compose --profile with-caddy up -d --build
# Hatch UI: https://{your-domain}
# Capture endpoint: https://{your-domain}/{endpoint-id}
```

### Architecture

```
Internet → :443 (Caddy) → hatch:8080 (Go binary, internal network)
                 │
                 ├─ Auto TLS (Let's Encrypt, or self-signed for localhost)
                 ├─ Reverse proxy with security headers
                 └─ JSON access logs to stdout
```

Caddy terminates TLS and reverse-proxies to the Hatch Go binary. The Hatch container only listens on `127.0.0.1:8080` — it's never directly exposed to the internet.

## CLI Reference

Hatch includes a powerful CLI for interacting with the server from the command line:

```bash
# Capture requests
hatch capture https://api.example.com/webhook

# Inspect captured traffic
hatch inspect my-webhook

# Search for specific requests
hatch search my-webhook -query 'status:500'

# Replay requests to other services
hatch replay <request-id> -endpoint my-webhook -target https://httpbin.org/post

# Configure mock responses
hatch mock set my-webhook -status 200 -body '{"ok":true}'

# Generate OpenAPI documentation
hatch doc generate my-webhook > openapi.json
```

For detailed CLI documentation, see [docs/engineering/cli.md](docs/engineering/cli.md).

## Technology Stack

See [docs/engineering/tech-stack.md](docs/engineering/tech-stack.md) for current choices and rationale, and [docs/adrs/](docs/adrs/) for the decision records that produced them.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

Hatch is released under the MIT License — see [LICENSE](LICENSE).
