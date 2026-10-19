# Review: REVIEW-STAGE Stage 3 — roblox-sync

> **Verdict:** FAIL  
> **Reviewer model:** composer-2.5-fast  
> **Implementer model:** Cursor Grok 4.6  
> **Artifact reviewed:** commit `ebcba4b` vs plan Stage 3  

Independent reviewer: [REVIEW-STAGE Stage 3](5259f4fd-ba19-4131-a4af-a9073a78c201)

## Summary

Stage 3 delivers `sync-agent/internal/mapfs` (Open/validate, scan, classify, inverse paths, Upsert/Delete/Rename) with broad L2 coverage against Stage 2 mapping fixtures. All **22** tests pass (`go test ./internal/mapfs/ -count=1`). CHECK updated Verification rows in §6.1, §6.2, §6.4, §9, §10.2, §12, §13, and §14; commit message includes RED/GREEN/CHECK; no `DEVIATION:` commits; no stage scope bleed (HTTP/CLI/watch/plugin).

**Failure driver:** The indexed test `TestMapfsUpsertIsAtomic_WhenWriteSucceeds` is claimed against `#apply-from-roblox-studio` (MUST: temp file + rename) but does **not** satisfy the matrix claim-check heuristic—it would still pass under a direct, non-atomic write and does not reliably detect leftover temps (prefix `.rbx-sync-*`, not `tmp`).

## Blockers (must fix before progression)

**B1 — Matrix claim check: atomic write (`TestMapfsUpsertIsAtomic_WhenWriteSucceeds`)**

- **Index claim:** `proj-spec/chapters/06.2-sync-agent.md` Verification names `TestMapfsUpsertIsAtomic_WhenWriteSucceeds` with `Spec-Anchor: …#apply-from-roblox-studio`, where normative text requires atomic writes via **temp file + rename in the destination directory**.
- **Evidence:** Production uses `atomicWrite` → `os.CreateTemp(..., ".rbx-sync-*")` + `replaceRename` (`apply.go`), which is aligned with spec.
- **Test gap:** The test only asserts (1) final file length/content and (2) no directory entry whose name contains substring `"tmp"`. A direct write to `Foo.luau` would pass. Leftover temps named `.rbx-sync-*` would **not** match the `"tmp"` substring check, so the test would not catch incomplete atomic cleanup either.
- **Plan RED:** Explicitly lists “atomic write (temp + rename)”.
- **Kick-back:** `@spec-driven-tdd` re-RED for Stage 3 — strengthen atomic-write test (e.g. assert no `.rbx-sync-*` entries, inject/hook failure path, or observe rename behavior) and refresh §6.2 Verification if the test name changes.

## Findings (non-blocking)

**F1 — §10.2 `#paths` coverage is thin**

- `TestMapfsOpenAcceptsProjectDir_WhenUsingOSSeparators` verifies Upsert + `filepath.Join` on the host OS; it does not match §10.2 “Expected tests: path join on **both** separators in Go.” Backslash-in-`disk` rejection is covered elsewhere (`invalid-backslash-disk` via `TestMapfsOpenRejectsInvalidProject_…`).

**F2 — §14 A10 index partial**

- A10 (“LF written on disk; UTF-8”) lists only `TestMapfsUpsertWritesLF_WhenSourceHasCRLF`. BOM strip/omit tests exist and are indexed under §6.4 `#line-endings-and-encoding` but not under §14 `#roblox-studio--disk-g1` / A10 row.

**F3 — §6.4 Verification heading for oversize**

- `TestMapfsUpsertRejectsOversize_WhenSourceExceedsOneMiB` is cited under `#line-endings-and-encoding` (where the 1 MiB limit also lives); §13 `#limits` is correctly cited on the test method. Not wrong, slightly redundant indexing.

**F4 — Extra L2 behavior**

- `TestMapfsUpsertRejectsFolderSource_WhenClassIsFolder` has method anchors but no §6.2 Verification row; acceptable (index cites decision on the test; not a D1 gap for an unclaimed row).

## Plan / spec alignment (table of RED items vs evidence)

| Plan Stage 3 RED item | Evidence | Status |
|----------------------|----------|--------|
| file↔node (fixtures) | `TestMapfsNodesMatchSpecExamples_WhenFixtureExists`, `TestMapfsFolderVsInitParent_…` | OK |
| inverse | `TestMapfsInversePrefersLuau_WhenCreatingFiles`, inverse upsert tests | OK |
| LF write | `TestMapfsUpsertWritesLF_WhenSourceHasCRLF` | OK |
| BOM strip on read | `TestMapfsReadStripsBOM_WhenFileHasBOM` | OK |
| CRLF→LF on write | same LF test + `normalizeSource` | OK |
| 1 MiB + 1 → `PAYLOAD_TOO_LARGE`, no partial write | `TestMapfsUpsertRejectsOversize_WhenSourceExceedsOneMiB` | OK |
| atomic write (temp + rename) | `atomicWrite` in code; **`TestMapfsUpsertIsAtomic_WhenWriteSucceeds`** | **FAIL (B1)** |
| upsert/delete/rename | Upsert + delete/rename tests | OK |
| same-root rename vs copy+delete | `TestMapfsRenameMovesOnSameRoot_…`, `TestMapfsRenameCopyDeletesAcrossRoots_…` + `Rename` impl | OK |
| empty mapped dirs removed except disk roots | `TestMapfsDeleteRemovesEmptyMappedDirs_WhenNotDiskRoot` | OK |
| invalid `rbx-sync.json` → `PROJECT_INVALID` | `TestMapfsOpenRejectsInvalidProject_…`, scan-time invalid tree tests | OK |
| GREEN: implement `internal/mapfs` | `mapfs.go`, `scan.go`, `mapping.go`, `apply.go` | OK |
| CHECK: Spec-Anchors + Verification | Chapter updates in commit; one tag per line on new tests | OK |
| AC A5 (L2), A10 | Fixture-driven mapping + LF test; §14 rows added | OK (A10 index thin — F2) |

## Traceability / coverage audit

**Stage-scoped matrix claim check (named *T* in CHECK-touched Verification):**

| Test *T* | Claimed *H* | Exists | Exercises *H*? |
|----------|-------------|--------|----------------|
| `TestMapfsUpsertIsAtomic_WhenWriteSucceeds` | §6.2 `#apply-from-roblox-studio` | Yes | **No (B1)** |
| `TestMapfsUpsertWritesLF_WhenSourceHasCRLF` | §6.2 `#apply-from-roblox-studio`, §6.4 `#line-endings-and-encoding` | Yes | Yes |
| `TestMapfsUpsertRejectsOversize_WhenSourceExceedsOneMiB` | §13 `#limits`, §6.1 `#error-codes` | Yes | Yes |
| `TestMapfsOpenRejectsInvalidProject_WhenFixturePackExists` | §6.4 `#project-file`, §6.1 `#error-codes` | Yes | Yes |
| `TestMapfsOpenAcceptsProjectDir_WhenUsingOSSeparators` | §10.2 `#paths` | Yes | Partial (F1) |
| Other new `TestMapfs*` rows in §6.4 / §6.2 | Respective headings | Yes | Spot-check OK |

**D1 placement (new behavioral tests):** All `TestMapfs*` functions have `Spec-Anchor` lines before `func`; no multiple tags on one line; all anchors include `#heading-slug`.

**Spec-code drift:** No normative mismatch found between mapfs behavior and §6.4 / §6.2 apply bullets (aside from untested atomic-write **proof** in tests).

**DEVIATION / plan deferral:** None.

**Tests run:** `go test ./internal/mapfs/ -count=1` — **PASS** (0.297s).

## Checklist

- [x] Commit is Stage 3 scoped (mapfs + CHECK chapters + tasks); parent follows Stage 2 review record
- [x] RED/GREEN/CHECK reflected in commit message
- [x] No `DEVIATION:` without spec-skill CHECK
- [x] Plan RED items implemented in code (except atomic-write **test** proof)
- [ ] **Atomic write RED + matrix claim for `TestMapfsUpsertIsAtomic_WhenWriteSucceeds` (B1)**
- [x] L2 tests consume Stage 2 mapping fixtures where intended
- [x] Spec-Anchors on new tests; one tag per line; headings present
- [x] CHECK updated Verification for stage-touched tests (no path-only claim where named *T* added)
- [x] No stage scope bleed (HTTP, CLI, watch, plugin)
- [x] `go test ./internal/mapfs/` passes
- [x] A5 (L2) mapping conformance evidenced
- [x] A10 LF behavior evidenced (BOM tests exist; §14 index incomplete — F2)
- [ ] **REVIEW-STAGE PASS criteria met** → blocked by B1
