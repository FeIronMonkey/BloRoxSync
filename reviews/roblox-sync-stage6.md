# Review: REVIEW-STAGE Stage 6 — roblox-sync

> **Verdict:** FAIL  
> **Reviewer model:** composer-2.5-fast  
> **Implementer model:** Cursor Grok 4.6  
> **Artifact reviewed:** commit `670922e` (“Stage 6: Serve bootstrap compare and apply HTTP”; parent `ebe6cd5`)

Independent reviewer: [REVIEW-STAGE Stage 6](2c0a2986-f899-4bb3-9e85-bf23d5e206bb)

## Blockers (must fix before progression)

1. **Matrix claim check — §9 D3 overclaimed (traceability)**  
   `proj-spec/chapters/09-design-decisions.md` Verification names `TestServeBootstrapStudioToDiskWritesMappedOnly_WhenNodesSubmitted` with `Spec-Anchor: …#d3--scripts-and-folders-only-in-v1`. The same test carries that anchor in `sync-agent/internal/serve/bootstrap_test.go`.  
   **D3** is the product-scope decision (“Scripts and Folders only in v1”); nothing in that test would fail if v1 allowed arbitrary instance classes at bootstrap. The test exercises **mapped-only writes**, **omitted mapped node deletion**, and **A20 disk preservation** (`README.md`, `notes.txt`), not rejection or scope limits for non-script/non-folder classes.  
   Per `#spec-coverage-audit` matrix claim check, this is a **stale/overclaimed** index row → **Blocker**.  
   **Fix (either):** remove the D3 row and the D3 `Spec-Anchor` from that test; **or** add a RED test that asserts bootstrap (compare/apply) rejects or ignores out-of-scope classes per normative HTTP/mapping rules, and index it honestly under D3.

## Findings (non-blocking)

1. **Revision table deferred (expected until Stage 7)**  
   `snapshotNodes()` always emits `"revision": 1`. That matches plan RED and tests on a **fresh serve** but does not yet implement §6.5 behavior after a prior live session increments revisions. Acceptable for Stage 6 scope.

2. **RED gaps — invalid apply payloads**  
   Covered: compare without `nodes`; `disk-to-studio` with `nodes`; `studio-to-disk` without `nodes`. Not covered: unknown `direction`; `compare` with `"nodes": null`. Low risk.

3. **A11b L3 is intentionally partial**  
   Compare-equal asserts no disk write. Full A11b remains L4 (Stage 10).

4. **A20 test scope**  
   Validates unmapped **disk** files. Unmapped Roblox Studio instances remain plugin-side.

5. **Stage scope**  
   No live apply, SSE, watch, or plugin work. No `DEVIATION:` in commit message.

## Checklist

- [ ] Matrix claim check — remove or honestly cover §9 D3
- [x] Plan Stage 6 RED intent evidenced
- [x] `go test ./internal/serve/` passes
- [x] CHECK chapter updates (except D3 overclaim)
- [x] No `DEVIATION:` without spec-skill CHECK
- [x] Out of scope (live apply, SSE, plugin Cancel) not implemented

**VERDICT:** FAIL

**Kick-back target:** `@spec-driven-tdd` CHECK for Stage 6 traceability — fix D3 index row and/or test anchor (or add a real D3-exercising test); no full re-RED unless behavior gaps are found while fixing the index.
