# Technology Stack

## Overview

These choices were made by the CTO on 2026-06-22 to align the engineering foundation with [ADR-0001](https://github.com/elfoundation/hatch/blob/main/docs/adrs/0001-stack.md) and the Hatch product thesis. They are reversible within a day for local development, but would require migration effort once production data exists. All choices default to boring, well-supported technology over novelty.

The stack is chosen so that a single `docker compose up` (or one binary on a VPS) is the entire product — no separate database server, no build step, no auth process. This is the v0.1 distribution promise; if a choice makes that promise harder, it has to earn its place.

## Backend & Data

| Layer | Choice | Rationale |
|---|---|---|
| Language | Go 1.25 | Single static binary, fast iteration, strong stdlib for HTTP and SQL. "One command on a VPS" requires a self-contained binary. |
| HTTP server | stdlib `net/http` (Go 1.22+ method-based routing) | No router dependency for a v0.1 surface. `go-chi/chi` is allowed where middleware is genuinely needed (SSE). |
| Database | SQLite via `modernc.org/sqlite` | Pure-Go driver keeps the "no CGO" promise. Survives restarts, zero-ops, single file you can `cp` to back up. |
| Migrations | Hand-rolled runner that applies `internal/storage/schema.sql` on first start, idempotently | Premature tooling is paid complexity. One SQL file, one `CREATE TABLE IF NOT EXISTS`, ship it. |
| Templates | stdlib `html/template` | Server-rendered HTML with auto-escaping. No JS build step, no hydration cost, no SPA complexity for what is fundamentally a server-driven UI. |
| Live updates | Server-Sent Events (SSE) on `net/http` | One-way push from server to browser, no WebSocket framing, plays nicely with the `html/template` render path. |

## Frontend

| Layer | Choice | Rationale |
|---|---|---|
| UI architecture | Server-rendered HTML + a little vanilla JS | The UI is a small request list with one live update. JS is loaded only where it earns its bandwidth (the SSE client). No bundler. |
| CSS | Hand-written CSS in `internal/handler/assets/` | Tailwind and friends add a build step we explicitly do not want in v0.1. A small `style.css` is enough for the surfaces we ship. |
| Client JS | Vanilla JS (no framework, no transpiler) | One `events.js` for the SSE subscription. If we ever need more, we will earn a toolchain. |

## Packaging & Distribution

| Layer | Choice | Rationale |
|---|---|---|
| Container | Multi-stage `Dockerfile` (golang:1.25 → scratch) | Produces an 8–10 MB static binary image. No runtime, no shell, no package manager on the final image. |
| Local dev | `docker compose` with optional Caddy sidecar | `docker compose up` is the one-command demo. Caddy handles TLS for the demo host and stays out of the way for plain HTTP. |
| Demo TLS | Caddy (auto-issues Let's Encrypt in prod, self-signed in dev) | Hands-off TLS, sane defaults, plays well with the static-binary promise. |
| Config | Env vars (read with `os.Getenv`) | Twelve-factor. No YAML/JSON config files for v0.1. Defaults documented in `.env.example`. |

## Tooling

| Layer | Choice | Rationale |
|---|---|---|
| Build | `go build` (with `CGO_ENABLED=0` for the final binary) | No Makefile needed for the common path. |
| Tests | stdlib `testing` + `net/http/httptest` | No assertion library until pain demands one. |
| Vet / lint | `go vet ./...` in CI | Stdlib tooling, no `golangci-lint` config to maintain. |
| CI | GitHub Actions (`.github/workflows/ci.yml`) | `go vet` → `go test` → `docker build` → smoke-test the running image. |
| Markdown lint | `markdownlint-cli2` (non-blocking) | Editorial consistency, not a gate. |
| Module path | `github.com/elfoundation/hatch` | Matches the GitHub org. Easy to change before any external consumer depends on it. |

## Principles

These are stack-agnostic and outlive any choice in the tables above. They are the cultural defaults; the tables are the implementation.

- **Server-rendered by default.** Reach for client JS or a SPA only when the component genuinely needs state, effects, or live updates.
- **Pure logic in `internal/`, I/O in adapters.** Business logic does not call `http.*` or `database/sql` directly. Storage and HTTP are replaced with interfaces in tests.
- **Store timestamps as UTC.** Render in local time only at the edge.
- **Single static binary, single file database.** If a dependency breaks that, justify it.
- **Observability before optimization.** Measure before fixing. No tuning without metrics.
- **Idempotency.** Operations should be safe to retry. Migrations, request capture, and SSE reconnect are all idempotent.
- **Boring technology.** A new dependency must earn its place. If stdlib covers it, stdlib wins.

## Open Decisions

| Decision | Status | Owner | Blocker |
|---|---|---|---|
| Router (stdlib `net/http` vs. `go-chi/chi`) | Provisional — start with stdlib, adopt chi if SSE middleware pain demands it | CTO | Live SSE implementation in v0.1 Inspect task |
| Structured logging library | Open — `slog` (stdlib) likely sufficient | CTO | First production deploy |
| Tracing / metrics | Open | CTO | Hosting decision |
| Hosting target (Fly, single VPS, k8s?) | Open | CEO | Need product requirements and traffic estimates |
