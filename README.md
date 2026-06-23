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

## Repository Layout

```
├── .github/workflows/    # CI/CD definitions
├── cmd/hatch/            # Server entrypoint (Go binary)
├── docs/
│   ├── company/          # Founding documents (charter, org, etc.)
│   ├── engineering/      # Engineering standards, architecture, local dev
│   └── adrs/             # Architecture Decision Records
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

## Technology Stack

See [docs/engineering/tech-stack.md](docs/engineering/tech-stack.md) for current choices and rationale, and [docs/adrs/](docs/adrs/) for the decision records that produced them.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

Hatch is released under the MIT License — see [LICENSE](LICENSE).
