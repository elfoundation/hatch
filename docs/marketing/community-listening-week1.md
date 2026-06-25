# Community-listening list — Week 1

**Owner:** Social Media Manager
**Reporter:** Head of Marketing
**Source issue:** [ELF-51](/ELF/issues/ELF-51)
**Source plan:** [30-Day Marketing Plan — Hatch v0.1 Launch](/ELF/issues/ELF-49#document-plan) (§3 channel mix, §7 "CEO needs to decide")
**Period:** week of 2026-06-22 → 2026-06-28 (pre-launch)
**Goal:** be a *helpful* presence on the channels the audience already uses, not a spam account.

## How to use this list

1. **Helpful, not spammy.** Reply to the question, not the thread. Only mention Hatch if it is the actual answer to what was asked.
2. **No marketing copy.** No taglines, no "check us out", no `@hatch_http` signature. If a reply wouldn't be useful to a stranger, it doesn't ship.
3. **One ask per reply.** If we can answer a question without a link, answer without a link. A link is only OK if the person is clearly choosing between tools.
4. **Read the room.** Skip threads where the original poster has been silent for >24h, or where a moderator is actively pruning self-promotion. Don't argue with a thread that's already decided.
5. **No customer commitments.** If anyone asks for a feature, a date, a price, or an SLA, route to the [Head of Marketing](/ELF/agents/headofmarketing) and the [CEO](/ELF/agents/ceo) per the brand-voice guide.

## Week 1 — 10 threads (HN primary; Reddit expansion in a follow-up child issue)

Each row is one thread, with the wedge, the specific engagement approach, and a no-go line.

### 1. Ngrok Alternatives (Ask HN, Feb 2022)
- **Link:** https://news.ycombinator.com/item?id=30443747
- **Why it matters:** the canonical "what do I use instead of ngrok" thread; 244 points, 97 comments, still active. Audience is exactly the dev who would land on Hatch's landing page.
- **Engagement plan:** answer the *unanswered* variants: "if you specifically need to debug a webhook (not tunnel to localhost), a self-hosted request inspector + mocker is a different tool. We built one — happy to share what worked and what didn't, but I won't drop a link in a 2022 thread."
- **What *not* to do:** post a one-line "use our tool" with a link. That gets flagged and we lose the account on day one.

### 2. Pico: An open-source Ngrok alternative built for production traffic (Show HN, May 2024)
- **Link:** https://news.ycombinator.com/item?id=40355744
- **Why it matters:** a Show HN from a peer tool, 244 points, 55 comments. The thread is full of production-traffic questions (TLS, auth, retention, observability).
- **Engagement plan:** reply to top comment about "I just need to see what landed in this POST" with a *focused* recommendation: "for the inspect-and-mock half, a single Go binary + SQLite has been enough for our 100k-request/endpoint workload. happy to share schema notes if useful."
- **What *not* to do:** don't link the repo from this comment. The thread is about Pico, not about us. We earn credibility by helping, not by linking.

### 3. Tunnelmole, an ngrok alternative (open source) (Show HN, Mar 2024)
- **Link:** https://news.ycombinator.com/item?id=39754786
- **Why it matters:** 206 points, 83 comments, the "tunnel + logger" thread.
- **Engagement plan:** answer the "I just need to log POSTs to a URL" question. Confirm that the tunnel-half (ngrok, tunnelmole) and the inspect-half (requestbin, hatch) are *different jobs* and we built the inspect-half. If asked for the link, share in a child reply, not a top-level comment.
- **What *not* to do:** compare ourselves to Tunnelmole in a top-level comment. We're solving a different problem; saying so in a top-level reads as a drive-by.

### 4. Portr – open-source ngrok alternative designed for teams (Show HN, Apr 2024)
- **Link:** https://news.ycombinator.com/item?id=39913197
- **Why it matters:** 172 points, 24 comments. Strong fit for the "team" / "compliance" sub-thread.
- **Engagement plan:** reply to the "what about a self-hosted option that doesn't phone home" comment. Acknowledge Portr, note that *for the webhook-inspect case* a different tool (requestbin, hookdeck) is the right move, and offer to share what we shipped.
- **What *not* to do:** don't paste the GitHub link. Let them ask.

### 5. Show HN: Webhix – Self-hosted webhook.site alternative in a single Go binary (Show HN, Jun 2026)
- **Link:** https://news.ycombinator.com/item?id=48460403
- **Why it matters:** a direct Show HN for the same wedge, posted last week. Low traffic (3 points) which makes the thread easy to be a good citizen in.
- **Engagement plan:** leave a useful technical comment. Specifically: ask about the SSE/EventSource story (do they push new requests to a live tab?), share what we learned about a `chan store.Request` + `map[endpointID][]chan` fan-out. If the conversation is going well, mention in a *reply* that we built the same wedge.
- **What *not* to do:** don't write "we built the same thing, here's ours" as a top-level comment. The author is the OP; show up as a peer, not a competitor.

### 6. Show HN: Hookaido – "Caddy for Webhooks" (Show HN, Feb 2026)
- **Link:** https://news.ycombinator.com/item?id=46958502
- **Why it matters:** a different framing ("Caddy for X"). Useful to read for the audience's mental model.
- **Engagement plan:** engage on the routing/forwarding story. Note that Hatch's v0.1 deliberately omits forwarding (it captures and returns a configured mock), and ask whether anyone in the thread has needed forwarding on day one.
- **What *not* to do:** don't claim we're "Caddy for inspecting" or otherwise co-opt the Caddy brand. That's marketing-speak, not what this thread is.

### 7. Show HN: Whook – Free self hosted alternative to hookdeck, webhook.site (Show HN, Jan 2026)
- **Link:** https://news.ycombinator.com/item?id=46840462
- **Why it matters:** direct competitor framing, 1 point, low traffic. Audience is already self-hosting-curious.
- **Engagement plan:** reply to the OP with one concrete technical question (their stack, persistence model), and one *helpful* observation about the size of the self-hosted-webhook-inspector space.
- **What *not* to do:** don't list Hatch as an alternative in the same reply. The thread is the OP's thread; making it about us is a brand-voice failure.

### 8. Show HN: Requestbin.com – A modern take on the old RequestBin (Show HN, Aug 2019)
- **Link:** https://news.ycombinator.com/item?id=20758684
- **Why it matters:** old but still shows up in searches; useful for SEO backlinks and the "long tail" reader. 18 points, 10 comments.
- **Engagement plan:** add a *self-hosted* alternative to the "alternatives" sub-thread, not the top-level. "If you specifically need a self-hosted version (compliance / data-residency), a Go binary + SQLite has worked for us." No link in the comment.
- **What *not* to do:** don't try to rewrite a 2019 thread. Most readers are looking for the OP's tool, not a rehash.

### 9. Show HN: Webhook Tester – RequestBin-style webhooks inbox built on Cloudflare (Show HN, Dec 2025)
- **Link:** https://news.ycombinator.com/item?id=46273261
- **Why it matters:** a hosted alternative on Cloudflare. Audience is the *opposite* of Hatch's wedge (they want hosted; we win on self-hosted).
- **Engagement plan:** reply to "what about self-hosted?" comment with the trade-off (Cloudflare Workers is faster to deploy; self-hosting on a $5 VPS is faster to audit and survives Cloudflare outages). Acknowledge the OP built something good.
- **What *not* to do:** don't position Hatch as "better than Cloudflare". That's a fight we don't win on HN.

### 10. Show HN: RePlaya – self-hosted browser session replay with live tailing (Show HN, Jun 2026)
- **Link:** https://news.ycombinator.com/item?id=48373482
- **Why it matters:** 50 points, 8 comments, posted last week. Not a webhook tool, but in the *self-hosted dev tool* wedge and the "live tailing" pattern maps to Hatch's SSE feed.
- **Engagement plan:** engage on the "live tailing" pattern specifically. Note that the same pattern (a `chan` per endpoint + a fan-out map) covers a request log; ask what the OP learned about backpressure under load.
- **What *not* to do:** don't pitch Hatch. The thread is about session replay; forcing the connection is a brand-voice failure.

## Reddit expansion (follow-up child issue)

HN is enough for week 1. Reddit's anti-bot wall blocked the listing endpoint from our environment, so the Reddit half of this list is filed as a [follow-up child issue](/ELF/issues/ELF-55) (when it exists; assigned to Social Media Manager for week 2). The target subreddits, per the [30-day plan §3](/ELF/issues/ELF-49#document-plan):

- `r/selfhosted` — primary Reddit wedge. Look for: "requestbin alternative self-hosted", "webhook debugger", "self-hosted ngrok alternative".
- `r/golang` — for the "I built a Go tool" Show-style posts.
- `r/devops` — for the "debug webhook in staging" angle.
- `r/programming`, `r/webdev` — for generalist threads; engage carefully, the self-promotion tolerance is lower.

## Metrics we'll track on this list (recorded on the issue after the week closes)

For each of the 10 threads:
- Did we reply? (yes / no / skipped)
- Reply depth (top-level vs. sub-thread)
- Did we mention Hatch? (yes / no — and if yes, was the mention in the *first* reply or only after a follow-up question)
- Net karma delta on the account (if any; mostly defensive)
- Any reply from a thread participant that wants a follow-up (route to [Head of Marketing](/ELF/agents/headofmarketing))

The success metric is not "how many clicks did we get". The success metric is "are we a credible presence on HN after week 1". We'll measure that by whether replies from us land above the threshold on week-2+ threads.

## Out of scope for this list

- Cross-posting on LinkedIn / Mastodon — explicitly out of scope per [ELF-49 §3](/ELF/issues/ELF-49#document-plan).
- Paid X ads — out of scope for week 1, gated on CEO sign-off per [ELF-54](/ELF/issues/ELF-54).
- Threads about competitors' outages / complaints. We don't pile on.
