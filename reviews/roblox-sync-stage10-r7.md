# Review: REVIEW-STAGE — roblox-sync Stage 10 r7
> **Verdict:** PASS
> **Reviewer model:** Composer 2.5 Fast
> **Implementer model:** Cursor Grok 4.6
> **HEAD reviewed:** `1630723`

## Blockers (must fix before progression)

- *(none)*

## Findings (non-blocking)

1. **`resolveRoot` false not asserted on the nested path under test** — The three new tests prove Folder/Script creation when `studio` is `ServerScriptService.RbxSyncL5`, but none assert `adapter.resolveRoot("ServerScriptService.RbxSyncL5") == false` before apply. `PluginStudioSkipsRoot_WhenStudioPathMissing` still guards lookup-only behavior on `ReplicatedStorage.Shared` (creating intermediates in `resolveRoot` would fail that test).
2. **§12 layers index is partial** — `12-testing.md` adds only `PluginHostAppliesDiskBootstrapToNestedRoot_WhenStudioPathMissing`, not the two `studio_test` nested-root rows. That matches the chapter’s sampling pattern (many studio tests are omitted); not a traceability gap because §6.3/§6.4 Verification indexes are complete for all three tests.
3. **Session live path with nested `studio` root** — Normative live create is covered at the adapter via `PluginStudioCreateInstanceCreatesNestedRootFolder_WhenStudioPathMissing`; integration is covered by the host bootstrap test. There is no separate `session_test` live-upsert with a nested `studio` root (similar to existing `PluginSessionCreatesFolderAncestors_WhenUpsertNewNodeId` for nodeId segments only).
4. **Pre-existing `ensureChild` edge case** — If a name on the mapped `studio` path exists with a non-`Folder`, non-mapped class (e.g. `Part`), `ensureChild` returns it without replacement and Folder creation can stall. Not introduced by this commit; L5 fixture uses a clean tree.
5. **`tmp/tasks.md` hygiene** — Stage 10 nested-root line still unchecked pending this review; update after PASS is recorded.

## Checklist

| Item | Result |
|------|--------|
| Re-RED intent (L5 A7 nested root) | **Pass** — Plan RED/GREEN name missing `Folder` segments on nested `studio` paths; GREEN adds `ensureStudioPath` and keeps `resolveRoot` lookup-only. |
| `ensureStudioPath` vs `resolveRoot` | **Pass** — `ensureNode` → `ensureStudioPath` + `ensureChild(..., "Folder")` for dot-path segments; `adapter.resolveRoot` still uses `resolveStudioPath` only; `rootInstFor` / `findInstance` / snapshot enumeration unchanged (lookup-only). |
| Spec adherence (§6.4) | **Pass** — CHECK prose matches code: disk-to-studio creates missing `studio` path segments as `Folder`; `resolveRoot` false must not skip that create; studio-to-disk still skip-only via session `resolveRoot`. |
| Anchors (D1 placement) | **Pass** — New tests: one `Spec-Anchor` per line, headings present (`#roblox-studio-path-resolution`, `#creating-and-deleting-instances`, `#connect-flow` on host test). §6.3/§6.4 Verification rows added with one tag per line. |
| Matrix claim check (stage-scoped) | **Pass** — All three named tests exist. Reverting `ensureNode` to `rootInstFor(root)` (pre-`1630723` behavior) leaves `RbxSyncL5` nil and fails folder/Script assertions (including host negative `Hello` not under SSS). `PluginStudioSkipsRoot_WhenStudioPathMissing` would fail if `resolveRoot` created intermediates. |
| No `DEVIATION:` without CHECK | **Pass** — Commit has CHECK only; chapter/plan updates reflect shipped behavior. |
| Normative spec↔code drift | **Pass** — None for this slice. |
| Stage scope bleed | **Pass** — `studio.luau`, three tests, spec/plan/tasks only; `newHttpService` `sessionRoots` / `bootstrapNodes` opts are test seams for the new host case. |
| Prior r6 (status contrast) | **Pass** — Not regressed; no host contrast changes in this commit. |
| L4 tests | **Pass** — Full plugin suite at HEAD. |

## Tests run this review

- `tmp\lune\lune.exe run studio-plugin/tests` (repo root) → **70 passed, 0 failed**

**VERDICT: PASS**
