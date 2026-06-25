# Role Definition: First Software Engineer

## Context

El Foundation has established its charter, operating model, and engineering baseline. We are pre-product but have a clear technical direction. The first engineer will work directly with the CTO to build the initial product and set the engineering culture for every hire that follows.

The engineering stack is **Go + SQLite + stdlib `net/http` + server-rendered HTML + SSE**, per [ADR-0001](https://github.com/elfoundation/hatch/blob/main/docs/adrs/0001-stack.md) and [ADR-0002](adrs/0002-hatch-detail-stack.md). The product is **Hatch**, a self-hostable HTTP request inspector + mocker that ships as a single static binary.

## What They Will Own

- **Product implementation.** Write the first lines of production code. Own features end-to-end from task assignment to merge.
- **Technical foundation.** Help solidify the stack, tooling, and conventions. The code you write sets the standard.
- **Quality bar.** Write tests, review PRs, and catch regressions before they reach users.
- **Documentation.** If it is not written down, it does not exist. Document APIs, runbooks, and decisions as you go.
- **Production reliability.** Once we have users, own on-call rotation with the CTO and ensure systems stay healthy.

## Technical Skills Required

### Must-Have

- **Go** — comfortable with the language, idiomatic style, the `net/http` package, and the standard library. A `go test ./...` / `go vet ./...` workflow is the daily loop.
- **HTTP fundamentals** — methods, status codes, headers, content negotiation, request/response lifecycle. SSE and chunked transfer are a plus.
- **SQL and relational schema design** — schema design, query optimization, migration discipline. SQLite-specific knowledge is a plus but not required.
- **Git and GitHub** — branching, rebasing, PR discipline, code review.
- **Server-rendered HTML** — at least one templating system (`html/template`, Jinja, ERB, Handlebars). Comfort with progressive enhancement.
- **Testing mindset** — writes tests as a default, not an afterthought. Knows when to reach for `httptest`, table tests, and in-memory fixtures.

### Nice-to-Have

- **`net/http` internals** — middleware, `http.Handler` composition, the `ServeMux` method-based routing added in Go 1.22.
- **Docker / multi-stage builds** — comfortable debugging a `Dockerfile`, reading layer output, and shipping a `scratch`-based image.
- **SQLite** — pragmas, indexes, JSON columns, the `modernc.org/sqlite` driver.
- **Linux / single-binary services** — has run a Go service in production behind a reverse proxy (Caddy, nginx, Envoy).
- **Devtools / API design taste** — has built or used webhook inspectors, request bin tools, or mocking tools. Empathy for the user is the job.
- **TypeScript or another typed language** — transferable skill; nice but not a v0.1 requirement.
- **Mongolian language or market context** — our early users are likely in Mongolia.

### Explicitly Deferred (not required for v0.1)

- Frontend frameworks (React, Next.js, Vue) — Hatch v0.1 is server-rendered. We may adopt one later; we are not starting there.
- TypeScript / Node.js — the foundation is Go. TypeScript remains in the toolbox for tooling only.
- Monorepo tooling (Turborepo, Nx) — single Go module, single binary, no need.
- Tailwind / utility CSS — hand-written CSS is enough for v0.1 surfaces.

## Attributes We Value

- **Slope over intercept.** We care more about how fast you learn than what you already know. Go fluency is learnable; engineering judgment is not.
- **Writes things down.** Async-first communication. Clear documentation. Decision records.
- **Disagrees and commits.** Healthy dissent, then full commitment once a decision is made.
- **Protects focus.** Ruthless prioritization. Says no to multitasking.
- **Pulls for bad news.** Surfaces problems early. Does not hide blockers.
- **Boring by default.** Reaches for the standard library before adding a dependency.

## First 30-Day Priorities

| Week | Focus | Deliverable |
|---|---|---|
| 1 | Onboard and ship the bootstrap | Merged `engineer/hatch-bootstrap` PR with `cmd/hatch` HTTP server, `Dockerfile`, CI green |
| 2 | Ship Task 2 (SQLite storage layer) | Merged PR with `internal/storage/...`, migration runner, `:memory:` tests, schema in PR description |
| 3 | First user-facing task (Capture, Inspect, or Mock) | At least one merged feature PR that exercises the full stack |
| 4 | Document and refine | Updated [`local-dev.md`](local-dev.md) / [`hatch-architecture.md`](hatch-architecture.md) with real friction, first own-authored ADR |

## Reporting

- **Reports to:** CTO
- **Peers:** None yet — you are the first IC
- **Growth path:** Senior Engineer → Staff Engineer → Engineering Lead as the team scales

## Compensation & Logistics

- **Decision owner:** CEO (pending approval)
- **Budget:** To be confirmed by CEO
- **Location:** Remote / async-first
- **Start date:** As soon as approved and hired

## Recommendation

**Hire a mid-level engineer with Go and HTTP-services experience.** They should have shipped a Go service to production and be comfortable with `net/http`, SQL, and Docker. A senior engineer would be ideal but may be overkill for our current stage and budget. A junior engineer would require too much hands-on guidance from the CTO, slowing both product velocity and hiring velocity.

If a strong TypeScript engineer is the only available candidate, that's acceptable — the stack is small and the ramp is short — but Go experience is preferred.

**Suggested title:** Software Engineer  
**Suggested level:** Mid-level (2–5 years shipping production code)  
**Priority:** High — we cannot build product without an engineer.
