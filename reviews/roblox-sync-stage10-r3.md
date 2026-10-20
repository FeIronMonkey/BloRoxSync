# Review: REVIEW-STAGE — roblox-sync Stage 10 r3
> **Verdict:** PASS
> **Reviewer model:** Composer 2.5 Fast
> **Implementer model:** Cursor Grok 4.6
> **HEAD reviewed:** `f7f7506`

## Blockers

*(none)*

## Findings

1. **Weak matrix exercise for widget choice labels (unchanged from r2)** — `PluginHostShowsConnectDisconnectAndChoices_WhenWidgetCreated` still asserts `app.visibleTexts`, seeded from static `LABELS` in `Host.start` and refreshed without requiring widget `TextButton` instances. §6.3 Verification cites this test for `#construction--configuration` and `#connect-flow` bootstrap/conflict affordances; host code does call `addButton` when a widget exists, but the index claim is not falsifiable via button removal. Tighten when convenient (e.g. assert fake-widget button children or click wiring).

2. **Conflict actions label-only (known product limit)** — `Keep disk` / `Keep Roblox Studio` in `host.luau` remain no-op handlers; conflict behavior stays on `session.luau` via existing `PluginSessionKeep*` L4 tests. Acceptable for this re-RED; L5 A8/A9 manual tails and full widget wiring remain open per `tmp/tasks.md`.

3. **Default DataModel adapter stub** — `Host.start` uses `stubStudio()` when `studio` is not injected; packaged artifact has no real Studio tree adapter until L5 work. Aligns with stated limits; L5 A7–A9 not claimed complete.

4. **HttpService `pcall` layering** — `RequestAsync` is still wrapped inside `session.luau`; `CreateWebStreamClient` additionally in `host.luau` `openStream`. `PluginHostWrapsHttpServiceInPcall_WhenRequestAsyncThrows` proves connect does not throw; it does not assert post-error status copy from the §6.3 error-handling table beyond not-reachable coverage elsewhere.

5. **IPv6 loopback at host layer** — `PluginSessionAcceptsHost_WhenAddressIsIPv6Loopback` covers `::1`; no dedicated `PluginHost*` widget-path test for `::1`. Host accepts via `isLoopback`; optional symmetry only.

6. **Sibling-require regression test is textual** — `TestPackagepluginHostRequiresSessionAsSibling_WhenPackaged` asserts embedded `host` ModuleScript CDATA contains `WaitForChild("session")` (B1 fix class met). It does not additionally assert a `ModuleScript` named `session` exists in the packaged tree (structural guarantee still implied by `TestPackagepluginWritesRbxmx_WhenSrcHasLuau` + packager listing all `src/*.luau` as children).

7. **B1 remedial commit message** — `f7f7506` documents the r2 B1 fix and CHECK in §6.3 but omits the plan’s usual inline `RED:` / `GREEN:` / `CHECK:` body lines; behavior and spec index are updated; process nit only.

8. **Packaged host still embeds Lune fallback** — CDATA retains `require("./session")` on the `pcall` failure path. Studio runtime should take the `WaitForChild("session")` branch when `host` and `session` are sibling ModuleScripts under the entry `Script`; L4 confirms fallback keeps Lune green (46 passed).

## Checklist

| Item | Result |
|------|--------|
| B1 fix (sibling `require` in Studio) | **Pass** — `studio-plugin/src/host.luau` tries `require(script.Parent:WaitForChild("session"))` first, then Lune `require("./session")`. Packager emits all `src` modules as sibling `ModuleScript` children under the entry `Script`. |
| Stage RED vs plan re-RED bullets | **Pass** — At HEAD (`4a1fb7f` + `f7f7506`): `PluginHost*` cover `CreateToolbar("rbx-sync")`, dock widget, defaults `127.0.0.1`/`29850`, `RequestAsync`/`CreateWebStreamClient`, token query, health-not-reachable, LAN reject; `TestPackagepluginScriptStartsHostWithPlugin_WhenPackaged` wires `plugin` + `HttpService` into `Host.start`; `TestPackagepluginHostRequiresSessionAsSibling_WhenPackaged` guards packaged host CDATA. Original Stage 10 session state machine remains in `PluginSession*` tests. |
| Spec adherence (`host.luau` + packager) | **Pass** — Toolbar name, dock widget, loopback gate, HttpService adapter, and packaged entry match §6.3 construction, HTTP, connect-flow wiring for L4 fakes. Packaged runtime load path no longer blocked by filesystem `require`. Keep-disk/Studio stub handlers and injected-only real studio remain within stated limits. |
| Anchors (D1 sites) | **Pass** — §6.3 Verification names stage tests including new `TestPackagepluginHostRequiresSessionAsSibling_WhenPackaged`; one `Spec-Anchor` per line in new/changed Verification rows and in `host_test.luau` / `packageplugin_test.go` behavioral tests; headings include `#slug`; no multi-tag-on-one-line in stage-touched tests. |
| Matrix claim check (stage-scoped) | **Pass** — Host HTTP/toolbar/widget/stream/not-reachable/LAN rows: named *T* exist and assertions fail if cited behavior is removed (except Finding #1 on choice labels). Go packaging rows for Script entry and host `WaitForChild("session")` exist and exercise cited strings/embedded source. |
| No `DEVIATION:` without CHECK | **Pass** — `f7f7506` and prior re-RED commit `4a1fb7f` have no `DEVIATION:`; no pending deviation CHECK in scope. |
| L4 / L2 run | **Pass** — `tmp\lune\lune.exe run studio-plugin/tests` → **46 passed, 0 failed**. `go test ./internal/packageplugin/ -count=1` from `sync-agent/` → **ok**. |

**VERDICT: PASS**
