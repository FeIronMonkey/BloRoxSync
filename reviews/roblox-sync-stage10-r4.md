# Review: REVIEW-STAGE — roblox-sync Stage 10 r4
> **Verdict:** PASS
> **Reviewer model:** Composer 2.5 Fast
> **Implementer model:** Cursor Grok 4.6
> **HEAD reviewed:** `0e6ccda`

## Blockers

*(none)*

## Findings

1. **Keep-disk host test still uses session stub, not `FakeDataModel`** — `PluginHostKeepDiskMatchesSnapshot_WhenClicked` goes through `startHost()`’s in-memory `newStudio()` stub. It validates widget → `resolveConflict(..., "keep-disk")` and snapshot source on the stub’s `mappedNodes()` table. §6.5 `#conflict-ui` observable DataModel write for Keep disk on the packaged path is not asserted here (session mock test `PluginSessionKeepDiskMatchesSnapshot_WhenConflict` + `writeSource` on real adapter via `PluginStudioApplyNodesWritesEditor_WhenTabOpen` cover pieces separately). Optional follow-up: one host-path test with `FakeDataModel` after Keep disk.

2. **Keep-disk `resolveConflict` mutates iteration copies on real adapter** — In `session.luau`, the loop over `self.studio.mappedNodes()` assigns `m.source = n.source` on returned node tables. On `studio.luau`, `mappedNodes()` builds fresh tables each call, so that assignment does not update instances; `writeSource` is the real write path. Harmless but redundant; stub tests could theoretically pass if `writeSource` regressed while the in-memory assign still updated the stub’s shared `mapped` array.

3. **Widget choice labels (unchanged from r3)** — `PluginHostShowsConnectDisconnectAndChoices_WhenWidgetCreated` still checks `app.visibleTexts` seeded from static `LABELS`, not widget `TextButton` children. §6.3 Verification cites this test for construction/connect affordances; host code does wire `addButton` when a widget exists, but the index claim is not falsifiable via button removal.

4. **`hasDuplicateLuaLuau` only on session mock** — Real `studio.luau` does not implement optional `hasDuplicateLuaLuau`; `_flushMutations` skips the check when absent. Duplicate refuse remains covered by `PluginSessionRefusesApply_WhenDuplicateLuaLuau` on the mock. Real-tree duplicate detection is still a product gap for L5, not introduced as normative drift in this commit.

5. **HttpService pcall layering (unchanged from r3)** — Host `httpFromService` calls `RequestAsync` without a local `pcall`; `Session:_request` wraps `http.request` in `pcall`, satisfying §6.3 “wrap HttpService calls.” `PluginHostWrapsHttpServiceInPcall_WhenRequestAsyncThrows` proves connect does not throw through the stack.

6. **Heartbeat test uses injected `heartbeatConnect`** — `PluginHostTicksSession_WhenHeartbeatFires` does not assert production `RunService.Heartbeat` binding; it proves the host callback flushes a debounced apply after 50 ms. Acceptable for L4; L5 still owns real Studio timing.

7. **Sibling-require regression tests remain textual** — `TestPackagepluginHostRequiresStudioAsSibling_WhenPackaged` (and session twin) assert embedded host CDATA strings, not a parsed rbxmx child list. Same class as r3 Finding #6; packager listing tests still imply all `src/*.luau` modules ship.

## Checklist

| Item | Result |
|------|--------|
| Re-RED plan bullets (L5 A7 DataModel) | **Pass** — `studio-plugin/src/studio.luau` DataModel + ScriptEditorService adapter; `host.luau` `defaultStudio()` from `game` when not injected; sibling `studio` require; Disk bootstrap calls `studio.applyNodes`; Keep disk / Keep Roblox Studio widget handlers call `session:resolveConflict`; Heartbeat (or inject) calls `session:tick`. |
| Prior r3 B1 (sibling `session`) | **Pass** — unchanged; still `siblingRequire("session", ...)`. |
| Spec adherence (§6.3, §6.4, §6.5, §5, §12) | **Pass** — CHECK updates match shipped behavior; new §6.3 normative line on packaged DataModel adapter reflected in code. |
| Anchors (D1 sites) | **Pass** — New/changed §6.3/§6.4/§6.5/§12 Verification rows name tests with one `Spec-Anchor` per line; new behavioral tests in `studio_test.luau`, `host_test.luau`, and Go `TestPackagepluginHostRequiresStudioAsSibling_WhenPackaged` carry anchors where required; no multi-tag-on-one-line in stage-touched tests. |
| Matrix claim check (stage-scoped) | **Pass** — Named stage tests exist; assertions fail if cited behavior is removed (disk bootstrap on `FakeDataModel`, adapter CRUD/listeners, Keep Studio force apply, heartbeat flush → `/v1/apply`, packaged `siblingRequire("studio"`). Weak spots: Findings #1, #3. |
| No `DEVIATION:` without CHECK | **Pass** — `0e6ccda` message has no `DEVIATION:`; no pending deviation CHECK in scope. |
| Normative spec↔code drift | **Pass** — No production behavior contradicting normative chapter text for stage scope; plan deferral not used to waive MUST items. |
| L4 / L2 run | **Pass** — `tmp\lune\lune.exe run studio-plugin/tests` → **60 passed, 0 failed**; `go test ./internal/packageplugin/ -count=1` → **ok**. |

## Tests run this review

- `tmp\lune\lune.exe run studio-plugin/tests` → 60 passed, 0 failed
- `go test ./internal/packageplugin/ -count=1` (from `sync-agent/`) → ok

**VERDICT: PASS**
