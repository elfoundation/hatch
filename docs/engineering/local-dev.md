# Local Development

This is the day-to-day workflow for working on Hatch. If something here disagrees with what you actually observe, update this doc — the docs are a living artifact.

## Prerequisites

- **Go 1.25 or newer** — `go version` should report 1.25+. Install from <https://go.dev/doc/install> or via your package manager.
- **Docker** (optional but recommended) — only required if you want to exercise the `docker compose` flow.
- **`curl`** — for smoke tests against the running server.
- **`sqlite3`** CLI (optional) — for poking at the database file directly. Not required for the normal workflow.

## Clone and Build

```bash
git clone https://github.com/elfoundation/hatch.git
cd hatch
go mod download
```

## Run the Server

The fastest path is `go run`:

```bash
go run ./cmd/hatch
# hatch starting on :8080
```

In another terminal:

```bash
curl -fsS http://localhost:8080/healthz
# ok
```

The server reads its port from the `PORT` environment variable and defaults to `:8080`.

```bash
PORT=9090 go run ./cmd/hatch
# hatch starting on :9090
```

## Run the Tests

```bash
go test ./...
```

This runs the full test suite, including the smoke test that boots the HTTP server in-process and hits `/healthz`. Add `-v` for verbose output, `-race` for the race detector.

```bash
go test ./... -v -race
```

## Vet and Format

```bash
go vet ./...
gofmt -l .
```

CI runs `go vet ./...`. `gofmt -l .` lists any unformatted files (no output means everything is formatted). Run `gofmt -w .` to fix formatting in place.

## Build a Static Binary

```bash
CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/hatch ./cmd/hatch
ls -lh bin/hatch
```

The resulting binary is fully static, has no libc dependency, and runs on any Linux x86_64 with the same kernel.

## Docker Compose

For the full local stack (Hatch + optional Caddy reverse proxy for TLS):

```bash
# Plain HTTP — Hatch only
docker compose up --build

# With Caddy (HTTPS, self-signed in dev)
docker compose --profile with-caddy up --build
```

The compose file reads `.env` for `HATCH_HOSTNAME`. Copy `.env.example` to `.env` and adjust:

```bash
cp .env.example .env
# Edit HATCH_HOSTNAME=localhost  (or your real domain)
```

## Where Things Live

| Path | Purpose |
|---|---|
| `cmd/hatch/` | Server entrypoint. `main.go` wires up `http.ServeMux`, `main_test.go` is the smoke test. |
| `internal/handler/` | HTTP handlers (Capture, Inspect, Mock). One file per route group. |
| `internal/store/` | Storage layer. `schema.sql` is the canonical DDL, `sqlite_repo.go` is the SQLite-backed implementation of the `Repository` interface. |
| `docs/engineering/` | Engineering standards and architecture docs. |
| `docs/adrs/` | Architecture Decision Records. |
| `Dockerfile` | Multi-stage build: `golang:1.25-alpine` → `scratch` with the static binary. |
| `docker-compose.yml` | Local stack: Hatch on `:8080`, optional Caddy sidecar for TLS. |
| `.env.example` | Documented environment variables. |
| `hatch.db` (gitignored) | SQLite database file. Created on first start in the working directory. |

## SQLite Tips

The database is a single file. By default it lives at `./hatch.db` in the process working directory.

```bash
# Inspect the schema
sqlite3 hatch.db '.schema'

# List endpoints
sqlite3 hatch.db 'SELECT * FROM endpoints;'

# Reset (destructive)
rm hatch.db
# Schema is recreated on next start
```

For tests that need a clean database, use the `:memory:` SQLite database — see `internal/store/sqlite_repo_test.go` for the pattern. Do not point tests at the on-disk database file.

## Common Tasks

| Task | Command |
|---|---|
| Run the server | `go run ./cmd/hatch` |
| Run all tests | `go test ./...` |
| Run a single test | `go test ./internal/handler -run TestCapture -v` |
| Vet | `go vet ./...` |
| Format | `gofmt -w .` |
| Build a binary | `CGO_ENABLED=0 go build -o bin/hatch ./cmd/hatch` |
| Docker stack | `docker compose up --build` |
| Reset the database | `rm hatch.db` |

## Troubleshooting

**`go: command not found`** — Install Go 1.25+ and ensure `$HOME/go/bin` (or wherever `go install` puts binaries) is on your `PATH`.

**`bind: address already in use`** — Another process is on `:8080`. Either stop it or run with `PORT=9090 go run ./cmd/hatch`.

**Tests fail with `database is locked`** — A previous test process left a handle. Look for stray `hatch` processes (`ps aux | grep hatch`) and kill them. Tests should use `:memory:` databases and not share the on-disk file.

**Docker build fails on `go.sum`** — Run `go mod tidy` locally, commit `go.sum`, rebuild. The Dockerfile copies `go.sum` before `go mod download` for reproducible builds.

**`CGO_ENABLED` warning during `go build`** — The static-binary build sets `CGO_ENABLED=0` explicitly. If you forget, the binary will dynamically link libc and break the "single static binary" promise.
