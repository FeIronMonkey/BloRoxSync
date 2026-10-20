# Review: REVIEW-STAGE Stage 9 — roblox-sync

> **Verdict:** PASS  
> **Reviewer model:** composer-2.5-fast  
> **Implementer model:** Cursor Grok 4.6  
> **Artifact reviewed:** commit `0264fdd`

Independent reviewer: [REVIEW-STAGE Stage 9](262db7a5-252d-4fa7-b44d-e3a871c2522d)

## Summary

Commit `0264fdd` delivers Stage 9 per plan: Lune L4 tests under `studio-plugin/tests` load the same seven behavioral mapping fixtures as `sync-agent/internal/mapfs` (`source-names.json`, `nodeid-examples.json`, `inverse-examples.json`, `folder-vs-init-parent.json`, `invalid-duplicate-lua-luau.json`, `invalid-multiple-init.json`, `invalid-forbidden-stems.json`). GREEN is a pure Luau module at `studio-plugin/src/mapping.luau` (no `HttpService`, no `@lune` in `src`). CI `l4-plugin` installs Lune 0.11.0 and runs `lune run studio-plugin/tests` instead of skipping when Lune is absent. CHECK adds PluginMapping rows to §6.4 Verification, D9, and §12; `references.md` documents Lune for L4.

**Tests:** `tmp\lune\lune.exe run studio-plugin/tests` from repo root — **8 passed, 0 failed**.

Plan RED items (names, nodeId, inverse, init semantics, duplicate lua/luau, forbidden stems) are covered. No `DEVIATION:` in the commit; no Stage 10 session/HttpService code. Matrix claim check (stage-scoped): every §6.4/D9/§12 row that names a `PluginMapping*` test has a matching test that would fail if the cited normative behavior were removed.

## Blockers

*(none)*

## Findings

1. **F1 — §14 A5 (L4) not indexed in Verification**  
   AC A5 requires L1 + L2 + **L4**, but `proj-spec/chapters/14-acceptance-criteria.md#protocol-and-mapping` still lists only L1/L2 mapfs/mapping tests. L4 is implemented and indexed in §6.4, but §14 traceability for A5’s L4 slice is incomplete (non-blocking for stage scope; align in a later CHECK or before FINAL).

2. **F2 — §12 lists one PluginMapping test**  
   Commit CHECK cites §12; Verification names only `PluginMappingNodeIdsMatchSpecExamples_WhenFixtureExists`. The module table already says `studio-plugin/tests` mapping covers §6.4; one row proves L4 runs fixtures. Remaining seven tests are indexed in §6.4 only.

3. **F3 — `PluginMappingFolderVsInitParent_*` weaker than Go L2**  
   `TestMapfsFolderVsInitParent_*` also asserts `server/Bar/Child` and `server/Qux/Child` are `ModuleScript`. Luau test checks `parentNodeId` and Bar/Qux class names but not child `className` (values are in the fixture). Matrix claim for init-vs-folder parent behavior still passes via Bar/Qux and parent links.

4. **F4 — `studioPath` split across two tests**  
   `PluginMappingNodeIdsMatchSpecExamples_*` does not assert fixture `studioPath` columns; `PluginMappingStudioPathWalksFromService_*` spot-checks two paths from `nodeid-examples.json`. Together they cover §6.4 `#nodeid` studio-path examples; not full row-by-row parity with fixture columns.

5. **F5 — §6.4 `#roblox-studio-path-resolution` remains path-only**  
   Pre-existing Verification line without named *T* is acceptable until a row claims a test for that heading (bootstrap/snapshot walking is Stage 10). `PluginMappingStudioPathWalksFromService_*` is indexed under `#nodeid`, not that heading.

6. **Carry-forward**  
   Project-file `Open` validation, unmapped-file ignore, LF/BOM/size upsert rules, and live Studio path walking remain L2-only or later stages (3/10+), consistent with plan Stage 9 scope.

## Checklist

- [x] Plan Stage 9 RED: same Stage 2 fixtures as Go for names, nodeId, inverse, init, duplicate lua/luau, forbidden stems; no Roblox Studio
- [x] GREEN: pure Luau mapping under `studio-plugin/src`; no HttpService / session
- [x] CHECK: Spec-Anchor indexing in §6.4, D9, §12 (§14 A5 L4 index still open — F1)
- [x] No `DEVIATION:` without spec-skill CHECK
- [x] No Stage 10 scope bleed
- [x] L4 tests pass (`lune run studio-plugin/tests`)
- [x] Matrix claim check (stage-touched §6.4 / D9 / §12 rows vs `PluginMapping*` tests)
- [x] One anchor tag per line in tests and stage-touched Verification

**Kick-back target:** *(none)*

**VERDICT: PASS**
