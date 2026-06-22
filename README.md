# El Foundation

Engineering repository for El Foundation.

## About

El Foundation builds institutions that outlast their founders. We create technology and organizations that compound in value over time. This repository is the source of truth for our engineering work.

## Getting Started

1. Read the [company charter](docs/company/charter.md) to understand why we exist.
2. Read the [operating model](docs/company/operating-model.md) to understand how decisions are made.
3. Read [CONTRIBUTING.md](CONTRIBUTING.md) before making any changes.

## Repository Layout

```
├── .github/workflows/    # CI/CD definitions
├── docs/
│   ├── company/          # Founding documents (charter, org, etc.)
│   ├── engineering/      # Engineering standards and decisions
│   └── adrs/             # Architecture Decision Records
├── apps/                 # Application code (TBD — created when product work begins)
├── packages/             # Shared libraries and packages
└── scripts/              # Automation and utility scripts
```

## Hatch — Deploy in one command

Hatch is a self-hostable HTTP request inspector + mocker. Ship it to any VPS with Docker.

### Quick start (local dev, no HTTPS)

```bash
docker compose up --build
# Hatch UI: http://localhost:8080
# Capture endpoint: http://localhost:8080/{endpoint-id}
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

See [docs/engineering/tech-stack.md](docs/engineering/tech-stack.md) for current choices and rationale.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

Proprietary — All rights reserved.
