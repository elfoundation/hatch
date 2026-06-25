# Ways of Working

## Version Control Discipline

- **One task = one branch = one owner.**
  Every issue gets its own branch. Branch names: `owner-identifier/short-description` (e.g., `ceo/ELF-3-hiring-plan`).
- **No direct commits to `main`.** All changes go through a pull request.
- **Rebase before merge.** Keep history linear and readable. Squash fixup commits.
- **Commit messages explain why, not what.** The diff shows what changed; the message explains the reasoning.

## Task Lifecycle

```
backlog → todo → in_progress → in_review → done
   ↑___________|    ↑___________|
```

- **backlog:** Valid idea, not scheduled. Anyone can add items.
- **todo:** Scheduled and ready. Has clear acceptance criteria.
- **in_progress:** Checked out by an owner. Active work is happening.
- **in_review:** Owner has handed off for review or approval.
- **done:** Accepted and merged. No further action required.
- **blocked:** Cannot proceed. Must name the blocker and who must act.

## Standups

We do not do live standups. We do **async written standups** in the issue thread or project channel.

### Standup Template

```
**Yesterday:** What did I complete?
**Today:** What am I working on?
**Blocked:** What is in my way and who needs to act?
```

- Posted by 10:00 AM local time, every working day.
- If you are blocked, you must name the owner of the unblock action.

## Definition of Done

A task is not done until **all** of the following are true:

1. **Code is written and reviewed.** At least one approval from a qualified reviewer.
2. **Tests pass.** CI is green. New code has relevant test coverage.
3. **Documentation is updated.** If the change affects APIs, interfaces, or runbooks, the docs are updated in the same PR.
4. **No secrets in plain text.** Credentials, tokens, and keys are never committed.
5. **User-facing changes are validated.** QA or the owner has verified the change in a staging environment.
6. **Rollback path is known.** For infra or deploy changes, the owner can articulate how to revert.
7. **Handoff is clean.** If follow-up work exists, it is captured in a new issue with acceptance criteria and an owner.

## Working Hours and Availability

- **Async-first.** We do not require synchronous availability. Write things down.
- **Response-time expectations:** 4 hours during business hours for blocking questions; 24 hours for non-blocking context.
- **Deep work blocks.** Meetings and interruptions are the exception. Protect 2-4 hour blocks for focused work.
