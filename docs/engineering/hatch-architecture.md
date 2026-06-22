# Hatch Architecture

Living document. The CTO sketches the component map; the engineer owns the details as the code lands.

## One-sentence summary

Hatch is a single Go binary that serves a server-rendered HTML UI and captures, inspects, and mocks HTTP requests against per-endpoint URLs, persisting everything in a single SQLite file.

## High-level component map

```
                        ┌─────────────────────────────────────┐
                        │  Browser (Hatch UI)                │
                        │  GET /e/{id}  +  EventSource /events│
                        └──────────────┬──────────────────────┘
                                       │ HTTP / SSE
                                       ▼
┌──────────────────────────────────────────────────────────────┐
│                     hatch (Go binary)                         │
│                                                              │
│  ┌──────────────┐   ┌────────────────┐   ┌───────────────┐   │
│  │  http.ServeMux│──▶│ handler layer  │──▶│  store layer  │   │
│  │  (stdlib)    │   │ (internal/     │   │  (internal/   │   │
│  │              │   │  handler/)     │   │   store/)     │   │
│  └──────────────┘   └────────────────┘   └───────┬───────┘   │
│         │                                        │           │
│         │             ┌────────────────┐          │           │
│         └────────────▶│  SSE hub       │◀─────────┘           │
│                       │  (broadcast    │                      │
│                       │   new requests)│                      │
│                       └────────┬───────┘                      │
│                                │                              │
│                       ┌────────▼───────┐                      │
│                       │ html/template  │                      │
│                       │ (server-render)│                      │
│                       └────────────────┘                      │
└──────────────────────────────┬───────────────────────────────┘
                               │ modernc.org/sqlite (pure Go)
                               ▼
                        ┌──────────────┐
                        │  hatch.db    │
                        │  (SQLite)    │
                        └──────────────┘
```

## Layer responsibilities

### `cmd/hatch/main.go` — process entrypoint

- Reads configuration (env: `PORT`).
- Wires the `http.ServeMux` to the handler layer.
- Starts `http.ListenAndServe`. Logs to stdout. Crashes on bind failure.

### `internal/handler/` — HTTP layer

- One file per route group: `capture.go`, `inspect.go`, `mock.go`, `sse.go`, `health.go`.
- Handlers take a `store.Repository` (interface) — never a concrete type — so tests can swap in `:memory:` SQLite or a fake.
- Handlers translate HTTP ↔ store calls. They do not contain business logic.
- Server-rendered HTML lives next to the handler that renders it. Templates are `//go:embed`-ed at build time.

### `internal/store/` — persistence layer

- `schema.sql` is the canonical DDL. Applied idempotently on first start by `migrate.go`.
- `models.go` defines the Go structs (`Endpoint`, `Request`).
- `repository.go` defines the `Repository` interface. The HTTP layer depends on this interface, not on the SQLite implementation.
- `sqlite_repo.go` is the concrete implementation. Uses `modernc.org/sqlite` (pure Go, no CGO).
- `db.go` opens the database file (or `:memory:` for tests), configures pragmas, returns a `*sql.DB`.

### SSE hub (in `internal/handler/sse.go`)

- A goroutine per connected browser, holding an `http.Flusher`.
- A `chan store.Request` that capture handlers publish to.
- The hub fans out new requests to all subscribers for the relevant endpoint.
- No external broker. No Redis. No pubsub library. A `chan` and a `map[endpointID][]chan` is the entire implementation for v0.1.

## Request lifecycle

1. **Capture** — a webhook hits `/{endpoint-id}` (any method). The capture handler:
   - Reads method, path, query, headers, body.
   - Calls `store.AppendRequest(ctx, request)`.
   - Publishes the new request to the SSE hub.
   - Looks up the endpoint's mock response and returns it.
2. **Inspect** — a browser hits `GET /e/{endpoint-id}`. The inspect handler:
   - Calls `store.ListRequests(ctx, endpointID, limit=100)`.
   - Renders an HTML page with the request list (newest first).
   - The page includes a small vanilla-JS `EventSource` client that subscribes to `/e/{endpoint-id}/events`.
3. **Live update** — the SSE handler holds the connection open, flushes each new request as a `data:` frame, and the browser appends it to the list.
4. **Mock** — `PUT /e/{endpoint-id}/mock` accepts `{status, headers, body}` and updates the endpoint. Subsequent captures to that endpoint return the configured response.

## Data model

Two tables. That's it for v0.1.

```
endpoints
  id           TEXT PRIMARY KEY      -- URL-safe random ID, e.g. "h7c2k"
  created_at   INTEGER NOT NULL      -- Unix epoch seconds, UTC
  mock_status  INTEGER               -- nullable; 200/201/204/etc.
  mock_headers TEXT                  -- JSON object, nullable
  mock_body    BLOB                  -- nullable

requests
  id           INTEGER PRIMARY KEY AUTOINCREMENT
  endpoint_id  TEXT NOT NULL REFERENCES endpoints(id)
  received_at  INTEGER NOT NULL      -- Unix epoch seconds, UTC
  method       TEXT NOT NULL
  path         TEXT NOT NULL         -- request path within the endpoint
  query        TEXT NOT NULL         -- raw query string
  headers      TEXT NOT NULL         -- JSON object
  body         BLOB                  -- nullable
  remote_addr  TEXT                  -- for debugging
  FOREIGN KEY (endpoint_id) REFERENCES endpoints(id) ON DELETE CASCADE
```

Index on `requests(endpoint_id, received_at DESC)` for the list page query.

## Boundaries

- **No authentication.** v0.1 is single-user, self-hosted. If you can reach the port, you can read and write. See [v0.1-scope.md](../adrs/v0-1-scope.md).
- **No multi-tenancy.** One SQLite file, one process, one operator. Cloud is v0.2+.
- **No external services.** No Redis, no Postgres, no S3. If the binary needs to phone home, the design is wrong.
- **No client-side framework.** The browser gets HTML and a 50-line `events.js`. Anything more is out of scope for v0.1.

## Performance budget

v0.1 is sized for a single developer self-hosting on a $5 VPS:

- **Cold start:** under 100 ms (Go binary + SQLite open).
- **Capture latency:** under 5 ms p99 on the happy path (no auth, no remote calls).
- **SSE fan-out:** one goroutine per connected browser, no message broker.
- **Database size:** comfortable to 100k requests per endpoint. Retention policy is out of scope for v0.1.

If a real workload breaks these, we measure first, then add complexity.

## Future seams (do not build now)

These are deliberate extension points, not planned features. The v0.1 implementation should not block them but also should not build them.

- **Pluggable storage** — the `Repository` interface is the seam. A Postgres or S3-backed implementation would slot in without handler changes.
- **Pluggable auth** — a middleware in front of the mux. The v0.1 mux is exposed without auth, which is the right v0.1 default.
- **Replay / forwarding** — the `Request` struct has the full request. Replay is "read from store, write back out." Forwarding is the same plus a destination.
- **Multi-tenant cloud** — partitioning by `endpoint_id` is already first-class. A `tenant_id` column is a v0.2 change.

When the v0.1 design cannot accommodate one of these without a rewrite, we have under-designed. When the v0.1 design has built one of these speculatively, we have over-designed.
