# Engineer Onboarding

## Before Day 1

- [ ] Access to GitHub org granted
- [ ] Access to Paperclip workspace granted
- [ ] Added to project channels / async standup

## Day 1: Context

- [ ] Read the [company charter](../company/charter.md)
- [ ] Read the [operating model](../company/operating-model.md)
- [ ] Read [ways of working](../company/ways-of-working.md)
- [ ] Read [how we decide](../company/how-we-decide.md)
- [ ] Read [CONTRIBUTING.md](../../CONTRIBUTING.md)
- [ ] Read [tech-stack.md](tech-stack.md)
- [ ] Read [local-dev.md](local-dev.md)
- [ ] Read [hatch-architecture.md](hatch-architecture.md)
- [ ] Read [ADR-0002: Hatch detail stack](adrs/0002-hatch-detail-stack.md)
- [ ] Introduce yourself in the team channel (async written standup format)

## Day 2–3: Environment

- [ ] Install Go 1.25 (or newer) — see the [Go install instructions](https://go.dev/doc/install) if you do not have it
- [ ] Verify `go version` reports 1.25 or newer
- [ ] Clone the repo
- [ ] Run `go mod download` to fetch dependencies
- [ ] Run `go test ./...` — should pass on a fresh clone
- [ ] Run `go run ./cmd/hatch` and `curl http://localhost:8080/healthz` — should return `ok`
- [ ] Run `docker compose up --build` and visit `http://localhost:8080/healthz` — should return `ok`
- [ ] Open your first PR (a README typo fix or doc improvement counts)

## Week 1: First Task

- [ ] Pick up a `good first issue` or grab a task from the backlog with CTO approval
- [ ] Follow the full task lifecycle: branch → PR → review → merge
- [ ] Shadow one code review as a reviewer (even if just observing)

## First 30 Days

- [ ] Ship at least one meaningful change to production (or equivalent if pre-launch)
- [ ] Write or update one piece of documentation
- [ ] Attend (async) one decision review or ADR discussion
- [ ] Provide feedback on the onboarding process itself

## Questions?

Ask the CTO or post in the project channel. Async-first: write it down.
