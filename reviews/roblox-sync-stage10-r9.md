# Review: REVIEW-STAGE — roblox-sync Stage 10 r9 (re-RED L5 A9)

> **Verdict:** PASS
> **Reviewer model:** Composer 2.5 Fast
> **Implementer model:** Cursor Grok 4.6
> **HEAD reviewed:** `712988d`
> **Parent context:** `f7885a5` (Stage 10 r8 PASS record); live-event code baseline `2964233`

## Blockers (must fix before progression)

- *(none)*

## Findings (non-blocking)

1. **No adapter-only L4 row for `pollMappedMutations`** — Behavior is covered by `PluginHostPostsRenameApply_WhenMappedNameChangesWithoutPropertySignal` (heartbeat → `Session:tick` → debounced `POST /v1/apply`). §6.3 / §14 index rows name the host test; a dedicated `studio_test` poll row would tighten isolation but is not required for this slice’s matrix claims.
2. **Carry-forward from r8: disk `rename` on poll/SSE** — `Session:applyDiskEvent` still handles only `upsert` / `delete`; disk-originated `rename` events remain unapplied on the live path. Out of scope for this re-RED (Studio→disk Explorer rename without Name signal); still relevant for full A7/A9 L5 manual rename symmetry.
3. **RED test omits `baseRevision` assertion** — Spec requires batched apply with `baseRevision` per node; the test asserts `op` / `nodeId` / `newNodeId` only. Snapshot seeds revision `14`; mock apply likely succeeds without asserting revision wiring (consistent with several other host mutation tests).
4. **`pollMappedMutations` `oldId == nil` branch emits `upsert`** — On heartbeat, unbound mapped instances get a one-shot upsert. After connect, `startListeners` pre-binds descendants, so the A9 scenario uses the `rename` branch; the upsert branch is reasonable for drift without `DescendantAdded` but is untested in isolation.
5. **`tmp/tasks.md` bundled L5 manual churn** — Same commit marks A7/A8 manual items `[x]` while adding the A9 re-RED line. Operational scratch-pad updates, not plugin code; A9 manual remains `[ ]` until L5 re-run on `tmp/l5-folder`.
6. **§12 Verification bullets** — Host module table row updated; §12 does not add per-test Verification bullets for the new host test (same partial-sampling pattern as r8).

## Checklist

| Item | Result |
|------|--------|
| Re-RED intent (L5 A9: Explorer rename without Name signal → disk apply) | **Pass** — `SetNameSilent` simulates missing signal; heartbeat path posts `rename` |
| Named RED test | **Pass** — `PluginHostPostsRenameApply_WhenMappedNameChangesWithoutPropertySignal` |
| GREEN vs spec (§6.3 change detection) | **Pass** — normative heartbeat compare + `op: rename`; `Session:tick` calls `pollMappedMutations` when live |
| CHECK (§6.3, §12 host row, §14 A9, plan Stage 10 re-RED) | **Pass** |
| Anchors (D1 placement) | **Pass** — new test has two anchor lines; §6.3 / §14 index rows name test + cite `#roblox-studio-change-detection` and `#roblox-studio--disk-g1` |
| Matrix claim check (stage-scoped) | **Pass** — named *T* exists; assertions fail if rename apply or nodeIds removed |
| No `DEVIATION:` without CHECK | **Pass** |
| Normative spec↔code drift (this slice) | **Pass** |
| Stage scope bleed | **Pass** — code limited to `studio.luau`, `session.luau`, tests, spec CHECK; tasks.md manual notes only |
| Prior r8 (live disk upsert) | **Pass** — not regressed |
| Tests encode behavior not implementation quirks | **Pass** |

## Tests run this review

- `C:\repos\roblox-sync\tmp\lune\lune.exe run studio-plugin/tests` (repo root) → **74 passed, 0 failed**

**VERDICT: PASS**
