# Contributing to El Foundation

## One Task = One Branch = One Owner

Every issue gets its own branch. Branch names follow this pattern:

```
owner-identifier/short-description
```

Examples:
- `cto/ELF-15-hatch-stack-migration`
- `engineer/hatch-bootstrap`
- `engineer/hatch-storage`

## No Direct Commits to `main`

All changes go through a pull request. No exceptions.

## Pull Request Process

**Important:** Only the CEO or CTO can merge pull requests. See [PR Submission Guidelines](docs/engineering/pr-submission-guidelines.md) for full details.

1. **Open a PR** from your branch to `main`.
2. **Fill out the PR template** (risk, rollback, verification).
3. **Request review** from the relevant owner:
   - Code changes → another engineer or CTO
   - UX-facing changes → UXDesigner
   - Security-sensitive changes → SecurityEngineer
4. **Address feedback** or escalate disagreements in writing.
5. **Wait for merge** — only CEO or CTO can merge your PR.
6. **Tag for merge** — comment on your PR: `@AlexChen or @JordanPatel - PR approved, ready for merge`

## Commit Messages

Commit messages explain **why**, not what. The diff shows what changed; the message explains the reasoning.

Good:
```
Add rotating refresh tokens

Using a rotating refresh token strategy prevents replay attacks
and gives us a clean theft-detection signal. See ADR-0003.
```

Bad:
```
Update auth.ts
```

## Code Style

- **Go, idiomatic.** `gofmt` clean, `go vet ./...` clean. Prefer the standard library over new dependencies. Reach for a third-party package only when stdlib genuinely does not cover the need.
- **Pure logic in `internal/`, I/O in adapters.** Business logic does not call `http.*` or `database/sql` directly. Storage and HTTP are replaced with interfaces in tests.
- **Server-rendered by default.** Reach for client JS or a SPA only when the component genuinely needs state, effects, or live updates. The v0.1 web UI is HTML templates plus a small vanilla-JS SSE client.
- **Keep packages small.** Prefer small, focused packages over clever abstractions. One file per route group in `internal/handler/`.
- **No comments unless the code is genuinely non-obvious** or there is a real `// FIXME`. Let the code explain itself; let the commit message explain the *why*.
- **No defensive error handling around things that should not fail.** Let it panic or return the error. Wrap at the boundary, not at every call site.

## Definition of Done

A task is not done until **all** of the following are true:

1. Code is written and reviewed.
2. Tests pass. `go test ./...` is green. CI is green.
3. Documentation is updated (`docs/engineering/` or `docs/adrs/` as appropriate).
4. No secrets in plain text.
5. User-facing changes are validated.
6. Rollback path is known.
7. Handoff is clean — follow-up work is captured in a new issue.

## Security

- Never commit secrets, credentials, or customer data.
- Security-sensitive changes (auth, crypto, secrets, permissions) require SecurityEngineer review before merging.
- Report vulnerabilities to the CTO immediately.

## Questions?

Open an issue or ask in the project channel. Async-first: write it down.
