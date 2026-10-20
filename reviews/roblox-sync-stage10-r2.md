# Review: REVIEW-STAGE — roblox-sync Stage 10 r2
> **Verdict:** FAIL
> **Reviewer model:** Composer 2.5 Fast
> **Implementer model:** Cursor Grok 4.6
> **HEAD reviewed:** `4a1fb7f`

## Blockers (must fix before progression)

### B1 — Packaged host cannot `require` session inside Roblox Studio

`studio-plugin/src/host.luau` loads the session module with Lune-style `require("./session")`. The packager embeds that source verbatim into the `host` `ModuleScript` in the generated `.rbxmx`, while the entry `Script` correctly uses `require(script:WaitForChild("host"))`.

Roblox Studio `require` expects a `ModuleScript` instance (or equivalent child reference), not a filesystem-relative string. At runtime in Studio, the packaged plugin is therefore expected to error when the entry `Script` starts `Host.start`, before toolbar/widget/HttpService behavior can run. That contradicts the Stage 10 re-RED intent (L5 A7 kick-back): packaged `.rbxmx` must start the Studio host with the `plugin` global, and §6.3 construction/configuration for a working dockable widget.

**Expected fix class:** `require(script.Parent:WaitForChild("session"))` (or the project’s established sibling-module pattern), with L4 still passing under Lune and a spot-check that packaged CDATA matches.

## Findings (non-blocking)

1. **Weak matrix exercise for widget choice labels** — `PluginHostShowsConnectDisconnectAndChoices_WhenWidgetCreated` asserts `app.visibleTexts`, which is initialized from the static `LABELS` table in `Host.start` and would still pass if widget `TextButton` creation were removed. §6.3 Verification cites this test for `#construction--configuration` and `#connect-flow` widget/bootstrap/conflict affordances. Code does call `addButton` when a widget exists; tighten the test (e.g. assert button instances or click handlers under the fake widget) so the index claim is falsifiable.

2. **Conflict actions label-only (known product limit)** — `Keep disk` / `Keep Roblox Studio` handlers in `host.luau` are empty; session conflict resolution remains covered by existing `PluginSessionKeep*` L4 tests. Acceptable for this re-RED scope; L5 A8/A9 manual tails and wiring remain open per `tmp/tasks.md`.

3. **Default DataModel adapter stub** — `Host.start` uses `stubStudio()` when no `studio` dep is injected; no real Studio tree adapter in the packaged artifact. Aligns with stated limits; L5 A7–A9 not claimed complete.

4. **HttpService `pcall` layering** — `RequestAsync` is invoked inside `session.luau`’s `pcall` around `http.request`; `CreateWebStreamClient` is additionally wrapped in `host.luau` `openStream`. `PluginHostWrapsHttpServiceInPcall_WhenRequestAsyncThrows` proves connect does not throw; it does not assert post-error status text from the §6.3 error-handling table (connect-flow “not reachable” is covered by a separate host test).

5. **IPv6 loopback at host layer** — `PluginSessionAcceptsHost_WhenAddressIsIPv6Loopback` covers `::1`; no dedicated `PluginHost*` test for `::1` in the widget path. Host code accepts `::1` via `isLoopback`; optional symmetry only.

6. **Packaging regression test is textual** — `TestPackagepluginScriptStartsHostWithPlugin_WhenPackaged` substring-checks Script CDATA; structural packaging remains guarded by `TestPackagepluginWritesRbxmx_WhenSrcHasLuau`. Sufficient for stage scope; does not catch B1-style require mistakes without reading embedded module bodies.

7. **Stage 12 packager touch in re-RED** — `packageplugin` entry change is intentional for FINAL B1 host wiring, not scope bleed; prior Stage 12 PASS artifact still valid for its RED, with runtime UI deferred to this re-RED.

## Checklist

| Item | Result |
|------|--------|
| Stage RED vs tests (re-RED bullets) | **Partial** — Plan re-RED items map to named `PluginHost*` tests and `TestPackagepluginScriptStartsHostWithPlugin_WhenPackaged`: toolbar `rbx-sync`, dock widget, defaults `127.0.0.1`/`29850`, `RequestAsync`/`CreateWebStreamClient`, token query, health-not-reachable, LAN reject, Script→`Host.start` with `plugin` + `HttpService`. **Packaged runtime load path fails B1.** |
| Spec adherence (`host.luau` + packager entry) | **Fail (B1)** — Toolbar/widget/HTTP adapter and loopback gate match §6.3 construction, HTTP, and connect-flow wiring for L4 fakes. Packaged sibling `require` breaks Studio runtime despite normative packaged-plugin host. Keep-disk/Studio buttons and stub studio are within stated limits. |
| Anchors (D1 sites) | **Pass** — §6.3 / §07.2 / §12 CHECK rows name stage tests; one `Spec-Anchor` per line in new Verification rows and `host_test.luau` / `packageplugin_test.go` behavioral tests; all include `#heading-slug`; no multi-tag-on-one-line in stage-touched tests. |
| Matrix claim check (stage-scoped) | **Partial** — Host HTTP/toolbar/widget/stream/not-reachable/LAN rows: named *T* exist and assertions fail if cited behavior is removed (except Finding #1 on choice labels). Go packaging row for `TestPackagepluginScriptStartsHostWithPlugin_WhenPackaged` exercises Script entry strings, not module `require` graph (B1). |
| No `DEVIATION:` without CHECK | **Pass** — Commit `4a1fb7f` body has no `DEVIATION:`; no pending deviation in scope. |
| L4 / L2 run | **Pass** — Re-run: `c:\repos\roblox-sync\tmp\lune\lune.exe run studio-plugin/tests` → **46 passed, 0 failed**. `C:\Program Files\Go\bin\go.exe test ./internal/packageplugin/ -count=1` from `sync-agent/` → **ok**. |

**VERDICT: FAIL**
