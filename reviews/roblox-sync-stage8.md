# Review: REVIEW-STAGE Stage 8 — roblox-sync

> **Verdict:** FAIL  
> **Reviewer model:** composer-2.5-fast  
> **Implementer model:** Cursor Grok 4.6  
> **Artifact reviewed:** commit `c8b7580`

Independent reviewer: [REVIEW-STAGE Stage 8](3b948d05-341a-4e4b-a205-7787668f214b)

## Summary

Stage 8 adds `internal/watch` (50 ms debounce, 500 ms self-write ignore via `NoteWrite`), an in-memory event log with monotonic cursor, `GET /v1/events` (SSE) and `GET /v1/changes` (poll), apply-path echo suppression, session replace / shutdown `event: end`, and wires `StartWatch()` from `rbx-sync serve`. CHECK replaces path-only §7.2 Verification rows with named tests across §4, §6.1–§6.2, §6.5, §7.1–§7.2, §12–§14. **`go test ./internal/watch/ ./internal/serve/ -count=1` passes** (watch ~2 s, serve ~3 s).

Plan RED intent is largely covered by behavioral tests (debounce coalesce, self-write, cursor 0→1→2, SSE content-type / no default backlog / `since` replay, heartbeat, `end` replaced/shutdown, poll non-blocking and `since=0` backlog, origin `disk`, 401 without SSE, idle watcher). No `DEVIATION:` commits; no plugin/Lune scope bleed.

**Normative gap:** With watch enabled at serve start, **bootstrap studio-to-disk writes do not participate in echo suppression**, so §6.5 self-write ignore is not met for all rbx-sync disk mutations. That is spec↔code drift introduced by this stage’s wiring, not waivable as “pre-bootstrap” behavior.

## Blockers (must fix before progression)

1. **B1 — Echo suppression incomplete once watch runs at serve start (§6.5 `#echo-suppression`, §6.2 `#watch-to-roblox-studio`)**  
   `cli.runServe` calls `syncSrv.StartWatch()` before HTTP serves. Live `POST /v1/apply` correctly calls `noteDiskWrite` before `Upsert`/`Delete`/`Rename` (`apply.go`). **`applyStudioSnapshot` in `bootstrap.go` writes via `Tree.Upsert` / `Tree.Delete` with no `noteDiskWrite`**, so Roblox Studio–win bootstrap can emit spurious `origin: disk` events into the log and SSE/poll while the spec requires rbx-sync to ignore paths it wrote within 500 ms. RED covered apply-only self-write (`TestServeApplyDoesNotEmitDiskEvent_WhenSelfWrite`); bootstrap is in scope now that watch is always on. **Kick-back:** `@spec-driven-tdd` re-RED Stage 8 — centralize disk-write notification (bootstrap + apply), add a watched-server bootstrap test proving no disk events for self-writes.

## Findings (non-blocking)

1. **F1 — Watcher errors not logged (§6.2 `#error-handling`)**  
   Normative table: watcher errors → log stderr, keep HTTP up. `watch.loop` discards `fw.Errors` with `_ = err` and does not write to stderr.

2. **F2 — Disk-originated delete / rename not exercised in Stage 8 tests**  
   Plan RED did not list them; `emitDiskPath` implements delete via `diskIndex` but only upsert is tested (`TestServeChangesOriginDisk_WhenFileWritten`). No `op: rename` from filesystem (rename likely surfaces as delete+upsert). §14 A7 L3 is indexed only via upsert — consistent with Stage 7-style partial AC indexing, not a missing RED row.

3. **F3 — Debounce/coalesce index headings (§6.2)**  
   `TestWatchCoalescesWrites_*` and `TestServeChangesCoalescesWrites_*` are indexed under `#construction--configuration` (debounce parameter table) rather than `#watch-to-roblox-studio`. Debounce is normative in the construction table, so matrix claim check passes; reverse lookup from `#watch-to-roblox-studio` omits coalesce tests (traceability only).

4. **F4 — `fsnotify` marked `// indirect` in `go.mod`**  
   Plan GREEN says pin `fsnotify`; version is pinned and `TestWatchGoModRequiresFsnotify_WhenModuleExists` passes. Direct `require` would match intent; not behavioral drift.

5. **F5 — Heartbeat interval proof split across two tests**  
   Production default is asserted by `TestServeEventsHeartbeatIntervalIsFifteenSeconds_WhenDefault`; mechanism by `TestServeEventsSendsHeartbeat_WhenStreamIdle` with injected 50 ms ticker in `newWatchedServer`. Together they satisfy §13 ≤ 15 s; neither alone proves end-to-end default timing under load (acceptable for L3).

6. **Carry-forward**  
   Plugin SSE client, 200 ms poll loop, disk delete/rename in Studio, and full A7 manual Studio remain later stages (plan Stage 10 / 14).

## Checklist

- [x] Plan Stage 8 RED intent covered by tests (apply self-write; bootstrap self-write **not** covered — see B1)
- [ ] GREEN implements normative §6.5 echo for **all** rbx-sync disk writes while watch is active (**B1**)
- [x] GREEN implements §6.2 watch debounce, §7.1 `/v1/changes`, §7.2 SSE/poll/stream lifecycle, §13 debounce/heartbeat/no busy loop (watch idle test), §4 watcher library presence
- [x] CHECK updates stage-relevant Verification (named tests; §7.2 D2 satisfied)
- [x] No `DEVIATION:` without spec-skill CHECK
- [x] No Stage 9+ plugin scope bleed in commit
- [x] `go test ./internal/watch/ ./internal/serve/ -count=1` pass
- [x] Matrix claim check: stage-touched index rows name existing tests that assert normative post-conditions (no stale *T*); coalesce/debounce rows exercise §13 limits
- [x] New behavioral tests carry `Spec-Anchor` with `#heading-slug`; no multiple tags on one line observed
- [ ] Normative chapter bullets fully met (**B1** fails §6.5 self-write ignore with serve-time watch)

**Kick-back target:** `@spec-driven-tdd` re-RED Stage 8 (bootstrap self-write ignore).

**VERDICT: FAIL**
