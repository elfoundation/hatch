# ADR-0002: Hatch detail stack

- Status: Accepted
- Date: 2026-06-22
- Authors: CTO
- Supersedes: none
- Related: [ADR-0001](https://github.com/elfoundation/hatch/blob/main/docs/adrs/0001-stack.md), [Hatch v0.1 plan (ELF-13)](https://github.com/elfoundation/hatch/issues/13)

## Context

[ADR-0001](https://github.com/elfoundation/hatch/blob/main/docs/adrs/0001-stack.md) fixed the v0.1 stack at the level of "Go + SQLite + stdlib `net/http` + server-rendered HTML + SSE + single Docker image." That ADR deliberately did not pick:

- which HTTP router (stdlib `net/http` with method-based mux vs. a third-party router)
- which SQLite driver
- which templating system (only said "server-rendered HTML")
- which test framework
- which license

This ADR closes those gaps with concrete choices, named alternatives, and the rollback path for each.

## Decision

| Concern | Choice | One-line reason |
|---|---|---|
| HTTP router | `go-chi/chi` (v5) for v0.1 | SSE middleware story is cleaner than hand-rolled on `http.ServeMux`; chi is the lightest router that still gives us real middleware composition. |
| SQLite driver | `modernc.org/sqlite` (pure Go) | Keeps the "no CGO" promise that the static-binary distribution rests on. |
| Templates | stdlib `html/template`, embedded with `//go:embed` | Auto-escaping by default, no template-engine dependency, no runtime parsing of the filesystem. |
| Test framework | stdlib `testing` + `net/http/httptest` | No assertion library unless pain demands one. |
| License | Apache-2.0 | Permissive, explicit patent grant, matches Go ecosystem norms. |
| Config | Environment variables only (`PORT`, `HATCH_DB_PATH`, `HATCH_LISTEN`) | Twelve-factor. No YAML/JSON config files for v0.1. |
| Logging | stdlib `log` to stdout, structured via `slog` (Go 1.21+) | No third-party logging library until we need log shipping. |
| Process model | Single process, one port, in-process SSE hub | v0.1 is a self-hosted single-binary product. No workers, no sidecars. |

### Router: `go-chi/chi`

Chi v5 is the smallest router that still composes well. It sits on top of `http.Handler`, so it does not break the stdlib-first principle — it just gives us `r.Use(middleware.Logger)`, `r.Route("/e/{id}", func(r chi.Router) { ... })`, and a clean SSE middleware (`middleware.Flush`).

**Why not stdlib `net/http` only?** Go 1.22 added method-based routing and path parameters to `http.ServeMux`. For v0.1's surface (~6 routes) it is genuinely enough. We adopt chi for two reasons: (1) the SSE middleware in `chi/middleware` is one line vs. ~20 lines of hand-rolled flush-on-write logic, and (2) the moment we add a second cross-cutting concern (request ID, auth-ready, rate limit) the hand-rolled composition gets ugly fast.

**Reversibility.** Drop-in: chi is an `http.Handler`. Reverting to stdlib `ServeMux` is a one-file change to `cmd/hatch/main.go`. The handler layer is router-agnostic by construction.

### SQLite driver: `modernc.org/sqlite`

Pure Go. Translates SQLite's C source via `ccgo` at build time, so the produced driver is a normal Go package with no `cgo` import.

**Why not `mattn/go-sqlite3`?** It is the most popular SQLite driver, but it requires CGO. CGO forces us to either (a) ship a glibc-linked binary that breaks on musl-based images, or (b) maintain a separate CGO build pipeline. Both are paid complexity for a feature (a C compiler in CI) we do not otherwise need.

**Why not the stdlib `database/sql` only?** The `database/sql` package is the API; it is not a driver. We still need a driver, and we still need migrations, and we still need a `Repository` interface above it. The choice here is the driver.

### Templates: stdlib `html/template`, `//go:embed`-ed

`html/template` auto-escapes context-sensitively (HTML, JS, URL, CSS contexts). It is in the standard library, has no dependencies, and produces no runtime surprises. We embed templates at build time with `//go:embed templates/*.html` so the binary is self-contained and there is no filesystem layout to manage at deploy time.

**Why not `templ` or `jet`?** Both add a build step (`templ generate`) to the dev loop. The v0.1 surface is small enough that a hand-rolled `{{range .Requests}}` is not a liability. We will revisit if template churn becomes a real cost.

### Test framework: stdlib `testing` + `net/http/httptest`

`testing` is in the standard library, plays well with `go test`, and is what every Go engineer already knows. `httptest` gives us `httptest.NewRecorder` for handler unit tests and `httptest.NewServer` for end-to-end smoke tests without a real network socket.

**Why not `testify` or `gocheck`?** No assertion library has yet earned its place. We will adopt one when the boilerplate of `if got != want { t.Errorf(...) }` becomes a real cost, not before.

### License: Apache-2.0

Permissive. Explicit patent grant. Matches what the Go ecosystem uses (Kubernetes, Docker, the Go toolchain itself are permissive-licensed). Compatible with corporate adoption without legal-review friction.

**Why not MIT?** MIT is also acceptable, but the patent grant in Apache-2.0 is a real protection for downstream contributors and is a default in our target ecosystem.

## Consequences

### Positive

- The full v0.1 dependency surface is: `go-chi/chi`, `modernc.org/sqlite`, and `github.com/google/uuid` (if we use it for endpoint IDs). Everything else is stdlib. That is a small, reviewable surface.
- The single-binary distribution story holds: `CGO_ENABLED=0 go build` produces a working static binary, and the Docker build ends in `scratch` with no runtime.
- The license, the test framework, and the template engine are all things an incoming Go engineer already knows. Zero ramp cost on tooling.
- Each choice has a one-day-or-less reversal path. The stack is composable from independent parts; nothing in this ADR forces a bigger rewrite to undo.

### Negative / Risks

- Chi is one more dependency to track. If a future Go stdlib release subsumes the middleware story, we have a small migration.
- `modernc.org/sqlite` is slower than the C version on raw throughput benchmarks (~2–3× on simple SELECTs). For v0.1's workload (a self-hosted single-user product) this is invisible. We will measure before optimizing.
- Stdlib `html/template` has no inheritance, no partials-with-arguments, and no "components." If template complexity grows, we will feel it. The mitigation is to keep the page count low (Capture page, Inspect page, Mock config page, error page).
- Apache-2.0 is more verbose than MIT. Trivial cost; mentioned for completeness.

## Alternatives Considered

### Option A: `net/http` only (no chi)

- Pros: zero router dependency; one fewer thing to upgrade; Go 1.22's `ServeMux` is genuinely good.
- Cons: SSE flush middleware is hand-rolled; auth/cors/rate-limit middleware are also hand-rolled; the cost of NOT having middleware composition grows with every new cross-cutting concern.
- Why rejected: SSE is in v0.1 (Inspect live). The first place we will want a second middleware is auth, even if auth itself ships later. Paying for chi now is cheaper than paying for hand-rolled composition later.

### Option B: `mattn/go-sqlite3`

- Pros: most popular SQLite driver, fast, well-known.
- Cons: requires CGO. Breaks the static-binary promise or forces a second build pipeline.
- Why rejected: the cost of CGO exceeds the value of `mattn`'s ecosystem familiarity.

### Option C: `templ` for templates

- Pros: type-safe templates, IDE autocompletion, components.
- Cons: requires a code-generation step (`templ generate`) on every save. The dev loop slows down. The runtime feature set is ~equivalent to `html/template` for our surface.
- Why rejected: the v0.1 template surface is small enough that hand-written `html/template` is not yet a liability. We can adopt `templ` later if template churn becomes a real cost.

### Option D: MIT license

- Pros: shorter, more familiar to individual contributors.
- Cons: no explicit patent grant.
- Why rejected: Apache-2.0's patent grant is a real downstream protection and matches our target ecosystem. The longer license header is a trivial cost.

## Rollback Plan

| Choice | Rollback |
|---|---|
| `go-chi/chi` | Replace `chi.Router` with `http.ServeMux` in `cmd/hatch/main.go`. Handlers and store are unchanged. ~half a day. |
| `modernc.org/sqlite` | Swap to `mattn/go-sqlite3`. Add CGO to the Dockerfile. Update CI to install `gcc`. ~half a day, plus the Dockerfile/CI change. |
| `html/template` | Adopt `templ` (or `jet`). Adds a code-generation step. ~1 day plus regen. |
| stdlib `testing` | Adopt `testify`. Mechanical import swap. ~2 hours. |
| Apache-2.0 | Relicense with contributor agreement. Not a code change, but a legal process. Not reversible by us alone. |
| Env-only config | Adopt `viper` or `koanf`. ~1 day including flag plumbing. |

The first five rollbacks are independent and can be done in any order. The license change is not ours to make unilaterally.

## Related

- [ADR-0001](https://github.com/elfoundation/hatch/blob/main/docs/adrs/0001-stack.md) — parent decision
- [Hatch v0.1 plan (ELF-13)](https://github.com/elfoundation/hatch/issues/13) — build plan and ordering
- [`docs/engineering/tech-stack.md`](../engineering/tech-stack.md) — current stack summary
- [`docs/engineering/hatch-architecture.md`](../engineering/hatch-architecture.md) — component map
- [`docs/engineering/local-dev.md`](../engineering/local-dev.md) — day-to-day commands
