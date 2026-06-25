# Operating Model

## Decision Rights

We run a **advice-and-decide** model, not consensus.

- **The owner decides.** Every initiative has a single named owner. That owner gathers input, makes the call, and owns the outcome. No design-by-committee.
- **The owner writes the decision record.** Every significant decision gets a one-paragraph rationale committed to the relevant issue or document.
- **Escalation is a feature, not a failure.** If an owner is stuck, blocked, or faces a one-way door they are uncomfortable opening alone, they escalate to their manager with a clear recommendation, not a question.

### Authority Matrix

| Decision type | Owner | Escalation path |
|---|---|---|
| Product feature priority | Product lead | CEO |
| Technical architecture | CTO | CEO |
| Design system changes | UXDesigner | CTO |
| Marketing message / channel | CMO | CEO |
| Hiring / firing | Department lead | CEO |
| Budget > $500/mo or one-time > $2,000 | CEO | Board |
| Security policy, auth, secrets | SecurityEngineer | CTO → CEO |
| Skill / agent governance | CEO | Board |

## Work Ownership

- **One task = one owner = one branch.** No shared ownership. If a task is too big for one person, split it.
- **Tasks are tracked in a structured workflow.** Backlog → assigned → in progress → review → done.
- **Context is durable.** Every task includes: objective, acceptance criteria, current blocker (if any), and next action. When an owner changes, the context transfers in writing.

## Review and Shipping

- **Nothing ships without review.** Code needs PR review. Design needs design review. Copy needs copy review. The owner names the reviewer, and the reviewer is accountable for the quality of their review.
- **Review is a gate, not a discussion.** Reviewers approve, request specific changes, or escalate. Open-ended "what do you think?" loops are not reviews.
- **Ship on green.** If CI passes and review is approved, the owner merges. No additional sign-off layers.

## Disagreement and Escalation

1. **First, disagree in writing.** Both parties write down their position and the evidence behind it. Oral disagreements are too cheap.
2. **If unresolved in 24 hours, escalate.** The owner escalates to the nearest common manager with both positions documented.
3. **The manager decides within 48 hours.** The manager picks a path, writes the rationale, and both parties commit.
4. **No re-litigation for 30 days.** Once decided, the issue is closed. Revisit only with new evidence.

## Communication

- **Async-first.** Write it down before you say it out loud. Meetings are for decisions that cannot be made async, not for status updates.
- **Default public.** Work in public channels and shared documents unless there is a specific confidentiality reason.
- **No surprises.** If something is going wrong, the affected people hear it from us before they hear it from data.
