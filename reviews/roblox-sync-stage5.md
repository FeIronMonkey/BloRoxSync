# Review: REVIEW-STAGE Stage 5 — roblox-sync

> **Verdict:** FAIL  
> **Reviewer model:** composer-2.5-fast  
> **Implementer model:** Cursor Grok 4.6  
> **Artifact reviewed:** commit `3c525ea`, plan Stage 5, `internal/serve`

Independent reviewer: [REVIEW-STAGE Stage 5](6aeec6cd-d618-4bc3-b12b-8115395d1531)

## Blockers (must fix before progression)

**B1 — Snapshot emits `revision: 0`**

`snapshotNodes` hard-codes `"revision": 0`. §6.5 revision initialization: fresh serve assigns revision **1**. §8 requires `revision` on snapshots after handshake.

**Kick-back:** `@spec-driven-tdd` re-RED Stage 5 — initialize revisions to **1** at serve; assert in snapshot tests.

**B2 — Missing token on unlisted `/v1` paths returns 404 instead of 401**

Unmatched patterns return `NOT_FOUND` before auth. §3 / §6.1 require `accessToken` on every `/v1` route except health. Authed unlisted → 404; missing token → 401.

**Kick-back:** `@spec-driven-tdd` re-RED Stage 5 — auth first on non-health `/v1`; add unauthorized unlisted-path test.

## Findings (non-blocking)

F1–F8 as recorded by the independent reviewer (unknown-field/`error` RED, golden diffs, stub IO_ERROR tree, §14 index gaps, error Content-Type, bootstrap path-only rows, apply stub, §6.2 stderr prose).

## Kick-back target

`@spec-driven-tdd` re-RED Stage 5 for B1 and B2, then full REVIEW-STAGE re-run.

**VERDICT: FAIL**
