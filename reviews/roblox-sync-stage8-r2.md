# Review: REVIEW-STAGE Stage 8 r2 — roblox-sync

> **Verdict:** PASS  
> **Reviewer model:** composer-2.5-fast  
> **Implementer model:** Cursor Grok 4.6  
> **Artifact reviewed:** commits `c8b7580` + `5d4f9fb` (prior FAIL `reviews/roblox-sync-stage8.md`)

Independent reviewer: [REVIEW-STAGE Stage 8 r2](4b4346b3-2c63-492c-a253-a52113819f99)

## Summary

Full re-review after prior **B1** (bootstrap studio-to-disk disk writes bypassed echo suppression while `StartWatch()` runs at serve start). Remedial commit `5d4f9fb` addresses the normative gap: `applyStudioSnapshot` calls `noteMappedWrite` before each `Tree.Upsert` and `noteMappedDelete` before each omitted-node `Tree.Delete`, then `refreshDiskIndex()`; apply paths use shared `noteMappedWrite` / `noteMappedDelete` helpers that delegate to `noteDiskWrite` → `watcher.NoteWrite`. CHECK indexes `TestServeBootstrapDoesNotEmitDiskEvent_WhenStudioToDisk` under §6.2 `#watch-to-roblox-studio` and §6.5 `#echo-suppression`. `go.mod` now direct-requires `fsnotify` (prior F4).

**B1 verification**

| Criterion | Evidence |
|-----------|----------|
| Bootstrap notes before disk mutations | `sync-agent/internal/serve/bootstrap.go` — `noteMappedWrite` / `noteMappedDelete` immediately before `Upsert` / `Delete` |
| Apply path parity | `sync-agent/internal/serve/apply.go` — same helpers for upsert/delete/rename |
| Watch active in integration test | `newWatchedServer` calls `StartWatch()`; production `cli.runServe` still calls `StartWatch()` before HTTP |
| Test fails on event leak | `TestServeBootstrapDoesNotEmitDiskEvent_WhenStudioToDisk` polls `/v1/changes` after 150 ms and fails if any changes |

**Tests:** `go test ./internal/watch/ ./internal/serve/ -count=1` — **pass**.

Plan Stage 8 RED is covered by named stage tests. No `DEVIATION:` commits; no Stage 9+ plugin/Lune scope bleed.

## Blockers

*(none)*

## Findings

1. **F1 — Watcher errors not logged (§6.2 `#error-handling`)**  
   Watcher errors → log stderr, keep HTTP up. `watch.loop` still discards `fw.Errors`. Non-blocking; unchanged from prior review.

2. **F2 — Disk-originated delete / rename not exercised in Stage 8 serve integration tests**  
   Only upsert-from-disk is integration-tested. Plan RED did not require them.

3. **F3 — Debounce/coalesce indexed under `#construction--configuration`**  
   Matrix claim check passes; reverse lookup from `#watch-to-roblox-studio` omits coalesce tests.

4. **F4 — `fsnotify` indirect require**  
   **Resolved in `5d4f9fb`:** direct require in `sync-agent/go.mod`.

5. **F5 — Heartbeat default vs mechanism split**  
   Default 15 s constant test plus injected-ticker mechanism test. Acceptable for L3.

6. **F6 — Bootstrap echo test is upsert-only**  
   Delete path of studio-to-disk is not integration-tested; `noteMappedDelete` is called. Non-blocking.

7. **Carry-forward**  
   Plugin SSE client, poll loop, Studio-side delete/rename, full A7 manual Studio — Stages 10 / 14.

## Checklist

- [x] Plan Stage 8 RED intent covered by behavioral tests (including bootstrap self-write after remedial)
- [x] GREEN implements normative §6.5 echo for apply **and** studio-to-disk bootstrap while watch is active
- [x] GREEN implements §6.2 watch debounce, §7.1 `/v1/changes`, §7.2 SSE/poll/stream lifecycle, §13 debounce/heartbeat/no busy loop, §4 watcher (`fsnotify` pinned)
- [x] CHECK updates stage-relevant Verification
- [x] No `DEVIATION:` without spec-skill CHECK
- [x] No Stage 9+ plugin scope bleed
- [x] `go test ./internal/watch/ ./internal/serve/ -count=1` pass
- [x] Matrix claim check (stage-scoped)
- [x] New behavioral tests use one anchor tag per line with `#heading-slug`
- [x] Normative §6.5 self-write ignore met with serve-time watch (B1 remediated)

**Kick-back target:** *(none)*

**VERDICT: PASS**
