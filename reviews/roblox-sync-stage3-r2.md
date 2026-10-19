# Review: REVIEW-STAGE Stage 3 r2 — roblox-sync

> **Verdict:** PASS  
> **Reviewer model:** composer-2.5-fast  
> **Implementer model:** Cursor Grok 4.6  
> **Artifacts reviewed:** commits `ebcba4b` + `383b1c2` (FAIL record) + `1e13d40` (remediation); full diff `ebcba4b^..HEAD`  
> **Supersedes:** [REVIEW-STAGE Stage 3 FAIL](roblox-sync-stage3.md) (B1 only)

Independent reviewer: [REVIEW-STAGE Stage 3 r2](94accb54-4d40-4181-a0d2-c996260ec3e1)

## Summary

Full Stage 3 re-review after remedial commit `1e13d40`. `sync-agent/internal/mapfs` implements Open/validate, scan, classify, inverse paths, Upsert/Delete/Rename with L2 coverage against Stage 2 mapping fixtures. **24** top-level tests in `./internal/mapfs/` (plus 5 table subtests under invalid-project open) all pass (`go test ./internal/mapfs/ -count=1`).

Prior **B1** is resolved: `TestMapfsUpsertIsAtomic_WhenWriteSucceeds` now hooks production seam `commitAtomic` and fails unless Upsert creates a temp in the destination directory and commits via rename; `TestMapfsUpsertPreservesDest_WhenAtomicCommitFails` proves the destination is unchanged when commit fails. Production `atomicWrite` uses `os.CreateTemp(..., ".rbx-sync-*")` in `filepath.Dir(path)` and `commitAtomic` → `replaceRename`. CHECK adds the new test to §6.2 Verification. No `DEVIATION:` commits; no normative spec↔code drift in stage scope; HTTP/CLI/watch/plugin not in scope.

## Blockers (must fix before progression)

None.

## Findings (non-blocking)

**F1 — §10.2 `#paths` coverage is thin** (unchanged from r1)

- `TestMapfsOpenAcceptsProjectDir_WhenUsingOSSeparators` exercises host-OS `filepath.Join` via Upsert; it does not implement §10.2 “path join on **both** separators in Go.” Backslash-in-`disk` rejection remains covered via invalid-project fixtures.

**F2 — §14 A10 index partial** (unchanged from r1)

- A10 row cites only `TestMapfsUpsertWritesLF_WhenSourceHasCRLF`; BOM strip/omit tests are indexed under §6.4 `#line-endings-and-encoding` but not under §14 `#roblox-studio--disk-g1`.

**F3 — §6.4 Verification heading for oversize** (unchanged from r1)

- Oversize test also cited under `#line-endings-and-encoding`; §13 `#limits` is on the test method — redundant, not wrong.

**F4 — Extra L2 behavior** (unchanged from r1)

- `TestMapfsUpsertRejectsFolderSource_WhenClassIsFolder` has method anchors but no dedicated §6.2 Verification row; acceptable for an unclaimed row.

**F5 — Atomic test seam (informational)**

- `var commitAtomic = replaceRename` is a deliberate test hook. Atomic proof depends on Upsert calling `commitAtomic` after `CreateTemp`; bypassing the var (e.g. inlining `replaceRename`) would fail the atomic tests — acceptable tradeoff for observable rename behavior.

## Plan / spec alignment (table of RED items vs evidence)

| Plan Stage 3 RED item | Evidence | Status |
|----------------------|----------|--------|
| file↔node (fixtures) | `TestMapfsNodesMatchSpecExamples_WhenFixtureExists`, `TestMapfsFolderVsInitParent_…` | OK |
| inverse | `TestMapfsInversePrefersLuau_WhenCreatingFiles`, inverse upsert tests | OK |
| LF write | `TestMapfsUpsertWritesLF_WhenSourceHasCRLF` | OK |
| BOM strip on read | `TestMapfsReadStripsBOM_WhenFileHasBOM` | OK |
| CRLF→LF on write | LF test + `normalizeSource` / upsert path | OK |
| 1 MiB + 1 → `PAYLOAD_TOO_LARGE`, no partial write | `TestMapfsUpsertRejectsOversize_WhenSourceExceedsOneMiB` | OK |
| atomic write (temp + rename) | `atomicWrite` + **`TestMapfsUpsertIsAtomic_WhenWriteSucceeds`**, **`TestMapfsUpsertPreservesDest_WhenAtomicCommitFails`** | OK |
| upsert/delete/rename | Upsert + delete/rename tests | OK |
| same-root rename vs copy+delete | `TestMapfsRenameMovesOnSameRoot_…`, `TestMapfsRenameCopyDeletesAcrossRoots_…` | OK |
| empty mapped dirs removed except disk roots | `TestMapfsDeleteRemovesEmptyMappedDirs_WhenNotDiskRoot` | OK |
| invalid `rbx-sync.json` → `PROJECT_INVALID` | `TestMapfsOpenRejectsInvalidProject_…`, scan-time invalid tree tests | OK |
| GREEN: implement `internal/mapfs` | `mapfs.go`, `scan.go`, `mapping.go`, `apply.go` | OK |
| CHECK: Spec-Anchors + Verification | Chapter updates; one tag per line on new tests | OK |
| AC A5 (L2), A10 | Fixture mapping + LF test; §14 rows (A10 index thin — F2) | OK |

## Traceability audit

**B1 remediation (matrix claim check):**

| Test *T* | Claimed *H* | Exists | Exercises *H*? |
|----------|-------------|--------|----------------|
| `TestMapfsUpsertIsAtomic_WhenWriteSucceeds` | §6.2 `#apply-from-roblox-studio` | Yes | **Yes** — hook requires temp `Stat` OK, `Dir(src)==Dir(dest)`, dest absent pre-rename, `sawTemp` via `commitAtomic`; direct `os.WriteFile` to dest would never call hook → fail |
| `TestMapfsUpsertPreservesDest_WhenAtomicCommitFails` | §6.2 `#apply-from-roblox-studio` | Yes | **Yes** — forced commit error; dest remains `old\n`; no stray entries |

**Other stage-scoped named *T* in CHECK-touched Verification (spot-check):**

| Test *T* | Claimed *H* | Exists | Exercises *H*? |
|----------|-------------|--------|----------------|
| `TestMapfsUpsertWritesLF_WhenSourceHasCRLF` | §6.2 `#apply-from-roblox-studio`, §6.4 `#line-endings-and-encoding` | Yes | Yes |
| `TestMapfsUpsertRejectsOversize_WhenSourceExceedsOneMiB` | §13 `#limits`, §6.1 `#error-codes` | Yes | Yes |
| `TestMapfsOpenRejectsInvalidProject_WhenFixturePackExists` | §6.4 `#project-file`, §6.1 `#error-codes` | Yes | Yes |
| `TestMapfsOpenAcceptsProjectDir_WhenUsingOSSeparators` | §10.2 `#paths` | Yes | Partial (F1) |
| Remaining indexed `TestMapfs*` in §6.4 / §6.2 | Respective headings | Yes | Spot-check OK |

**D1 placement:** New behavioral tests in `mapfs_test.go` and `apply_atomic_test.go` carry `Spec-Anchor` lines before `func`; one tag per line; all include `#heading-slug`. §6.2 Verification names atomic tests with matching Spec-Anchors (no path-only gap for named *T*).

**Spec-code drift:** None identified for normative mapfs apply/scan/mapping behavior in §6.2 / §6.4 / §6.1 error paths exercised this stage.

**DEVIATION / plan deferral:** None; plan deferral not used to waive normative behavior.

**Tests run:** `"C:\Program Files\Go\bin\go.exe" test ./internal/mapfs/ -count=1` — **PASS**.

## Checklist

- [x] Stage 3 scope: mapfs + CHECK chapters; remedial commit on prior FAIL record
- [x] RED/GREEN/CHECK in stage commits (`ebcba4b`, `1e13d40`)
- [x] No `DEVIATION:` without spec-skill CHECK
- [x] Plan RED items implemented and evidenced in tests
- [x] **B1 atomic write + matrix claim for `TestMapfsUpsertIsAtomic_WhenWriteSucceeds`** (fixed in `1e13d40`)
- [x] **Preserve dest on failed atomic commit** (`TestMapfsUpsertPreservesDest_WhenAtomicCommitFails`)
- [x] L2 tests consume Stage 2 mapping fixtures where intended
- [x] Spec-Anchors on new behavioral tests; one tag per line
- [x] CHECK Verification updated for stage-touched named tests
- [x] No stage scope bleed (HTTP, CLI, watch, plugin path-resolution deferred to Stage 10)
- [x] `go test ./internal/mapfs/` passes
- [x] A5 (L2) mapping conformance evidenced
- [x] A10 LF behavior evidenced (§14 BOM index gap — F2 only)
- [x] **REVIEW-STAGE PASS criteria met**

**Kick-back target (prior FAIL):** `@spec-driven-tdd` re-RED Stage 3 for B1 — **addressed** in `1e13d40`. Authorized next phase: Stage 4 RED per `proj-spec/implementation-plan.md`.
