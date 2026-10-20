# Review: REVIEW-STAGE — roblox-sync Stage 10 r5
> **Verdict:** PASS
> **Reviewer model:** Composer 2.5 Fast
> **Implementer model:** Cursor Grok 4.6
> **HEAD reviewed:** `f1b2be3`

## Blockers

*(none)*

## Findings

1. **Host status tests assert `app.status`, not `statusLabel.Text`** — `PluginHostShowsBootstrapStatus_WhenConnectUnequal`, `PluginHostShowsConnectedStatus_WhenCompareEqual`, and `PluginHostShowsDisconnectedStatus_WhenCancelAfterUnequal` would still pass if `refreshTexts()` stopped assigning `statusLabel.Text` while `app.status` stayed correct. Under Lune, `tryInstance("TextLabel")` often no-ops, so the widget label path is largely unexercised in L4. Production code does set `statusLabel.Text = app.status` in `refreshTexts()` after Connect and Cancel/Disk/Studio handlers.

2. **Disk / Roblox Studio wins → `connected` on host not covered by new tests** — Session `_enterLive()` sets `STATUS_CONNECTED`; `clickDiskWins` / `clickStudioWins` copy `session.status` and refresh. Existing `PluginSessionDiskWinsOmitsNodes_WhenApplyingBootstrap` asserts `live` but not `status == "connected"`. Gap is minor for this re-RED (hunt focused on unequal prompt, equal live, Cancel disconnect).

3. **`clickConnect` outer `pcall` around `Session.connect`** — Connect normally returns a session table with `status` set rather than throwing; inner HTTP still goes through session `_request` pcalls. Extra host-level guard is harmless; `PluginHostWrapsHttpServiceInPcall_WhenRequestAsyncThrows` unchanged.

4. **`httpFromService` omits `Body` when nil** — GET-style calls no longer pass an empty `Body` key to `RequestAsync`. Sensible ancillary fix in the same commit; not called out in plan RED bullets but does not contradict spec.

5. **Inherited from r4 (unchanged in this commit)** — Keep-disk host test still uses in-memory studio stub; widget construction test still checks static `LABELS` / `visibleTexts` rather than live button children; `hasDuplicateLuaLuau` only on session mock; heartbeat test uses injected callback; sibling-require regression tests remain textual. None are regressions introduced by `f1b2be3`.

## Checklist

| Item | Result |
|------|--------|
| Re-RED plan bullets (L5 connect status) | **Pass** — `session.luau`: `STATUS_BOOTSTRAP` after unequal compare, `STATUS_CONNECTED` in `_enterLive`, `STATUS_DISCONNECTED` on Cancel. `host.luau`: after Connect copies `session.status` + `refreshTexts()`; `clickCancel` / `clickDiskWins` / `clickStudioWins` sync status and refresh; widget buttons route through those handlers. |
| Prior r4 DataModel / Keep / sibling require | **Pass** — not regressed; out of scope for this commit. |
| Spec adherence (§6.3 connect-flow + error-handling status table, §12 module map) | **Pass** — Normative status strings documented in connect-flow steps 6–8 and error-handling table; code constants match prose. |
| Anchors (D1 sites) | **Pass** — Six new behavioral tests each carry one `Spec-Anchor` per line; §6.3 Verification lists name those tests with separate anchor lines (incl. §6.5 `#bootstrap` on Cancel rows). |
| Matrix claim check (stage-scoped) | **Pass** — Named tests exist; assertions fail if bootstrap/connected/disconnected session status or host `app.status` after Connect/Cancel are removed. Weak spots: Findings #1–2 (label text, disk-win connected on host). |
| No `DEVIATION:` without CHECK | **Pass** — Commit message has no `DEVIATION:`. |
| Normative spec↔code drift | **Pass** — Addresses reported bug (Connect/Cancel leaving **disconnected** after handshake); no contradiction of normative chapter text for this slice. |
| Stage scope bleed | **Pass** — Changes confined to session/host, tests, spec CHECK, plan note, `tmp/tasks.md`. |
| L4 tests | **Pass** — `tmp\lune\lune.exe run studio-plugin/tests` → **66 passed, 0 failed** (+6 vs r4). |

## Tests run this review

- `tmp\lune\lune.exe run studio-plugin/tests` → 66 passed, 0 failed

**VERDICT: PASS**
