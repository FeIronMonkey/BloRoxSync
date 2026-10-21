# Review: REVIEW-STAGE — roblox-sync Stage 10 r8

> **Verdict:** PASS
> **Reviewer model:** Composer 2.5 Fast
> **Implementer model:** Cursor Grok 4.6
> **HEAD reviewed:** `2964233`

## Blockers (must fix before progression)

- *(none)*

## Findings (non-blocking)

1. **Live `rename` still not applied on poll/SSE path** — §7.2 event objects include `op: rename`, and CHECK prose requires each poll/SSE entry to be applied like `event: change`. `Session:applyDiskEvent` only handles `upsert` and `delete`; `rename` events routed through `_pollChanges` / `onStreamMessage` are silently ignored. This predates the wiring fix (direct `applyDiskEvent` tests never covered rename). L5 A7 rename remains dependent on studio-side behavior or a future RED row; not a regression in this slice, which targets upsert apply.
2. **No L4 assertion for `Enum.WebStreamClientType.SSE`** — `host.luau` sets `streamType` via `pcall` on `Enum.WebStreamClientType.SSE` (correct for Studio; `nil` under Lune). `PluginHostCallsCreateWebStreamClient_WhenEnteringLive` does not assert `log.streams[1].streamClientType` when a test double provides `Enum`.
3. **§12 Verification sampling** — CHECK updates the host module row for `MessageReceived` disk apply but does not add Verification bullets for the three new tests (consistent with §12’s partial sampling; §6.3 / §6.5 / §7.2 indexes name them).
4. **Session unit mock does not exercise SSE callback** — `session_test` `openStream` ignores the `onMessage` callback when the stream succeeds; SSE apply integration is covered by `PluginHostAppliesDiskUpsert_WhenStreamMessageReceived`.
5. **`PluginSessionPollsSinceCursor_WhenChangesReturned`** — Asserts `since` on the second poll, not a second upsert apply (mock returns the same change batch every time). Matches RED intent (cursor passthrough).
6. **Bundled `reviews/roblox-sync-stage10-r7.md`** — Commit adds the r7 artifact alongside code; not stage scope bleed for this slice.
7. **`tmp/tasks.md`** — Stage 10 live-event line still unchecked pending this review; update after PASS is recorded.

## Checklist

| Item | Result |
|------|--------|
| Re-RED intent (L5 A7 live disk events → DataModel) | **Pass** |
| Named RED tests | **Pass** |
| SSE + poll vs spec (§6.3, §6.5, §7.2) | **Pass** |
| Anchors (D1 placement) | **Pass** |
| Matrix claim check (stage-scoped) | **Pass** |
| No `DEVIATION:` without CHECK | **Pass** |
| Normative spec↔code drift (this slice) | **Pass** for upsert wiring; rename gap is finding #1 |
| Stage scope bleed | **Pass** |
| Prior r7 (nested root) | **Pass** — not regressed |
| L4 tests | **Pass** — 73 passed, 0 failed |

## Tests run this review

- `tmp\lune\lune.exe run studio-plugin/tests` (repo root) → **73 passed, 0 failed**

**VERDICT: PASS**
