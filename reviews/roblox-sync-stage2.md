# Review: REVIEW-STAGE Stage 2 — roblox-sync
> **Verdict:** PASS  
> **Reviewer model:** composer-2.5-fast  
> **Implementer model:** Cursor Grok 4.6  
> **Artifact reviewed:** commit b48f005 vs plan Stage 2  

Independent reviewer: [REVIEW-STAGE Stage 2](e3ed7c5e-2bcd-4b11-aaec-97bdd02873cf)

## Summary

Commit `b48f005` delivers the Stage 2 scope: `sync-protocol/fixtures/v1/mapping/` (14 JSON cases) and `sync-protocol/tests/mapping_test.go` (16 tests). CHECK updates name Stage 2 tests in §2, §6.4, §9, §11, §12, and §14 Verification. L1 suite passes: `go test ./tests -count=1` in `sync-protocol` → **ok**. Plan AC **A5 (L1 fixture pack)** and goal **G7** are satisfied at the fixture/conformance-table layer; runtime mapping is explicitly deferred to Stages 3 and 9.

## Blockers (must fix before progression)

None.

## Findings (non-blocking)

1. **Fixture-only vs runtime enforcement** — Tests such as `TestMappingProjectRequiresUniqueStudio_WhenInvalidFixtureExists`, `TestMappingProjectRequiresNonOverlappingDisk_WhenInvalidFixtureExists`, and the source-name invalid cases assert `expectedError`, `reason`, and that the fixture encodes the violation; they do **not** call `mustReject` on `project.schema.json` (unlike dot-dot and backslash). That matches Stage 2’s “encode conformance tables” intent and the plan note that runtime conformance is Stages 3/9, but test names sound like live enforcement. Stage 3 RED should wire these fixtures to actual `PROJECT_INVALID` behavior.

2. **`reason` assertion inconsistency** — `invalid-duplicate-root-id.json` includes `"reason": "duplicateRootId"`, but `TestMappingProjectRequiresUniqueRootIds_WhenInvalidFixtureExists` does not assert `reason` (sibling invalid-project tests do). Harmless for Stage 2; easy consistency fix in a later stage if desired.

3. **§14 A5 index is partial by design** — §14 Verification lists four representative mapping tests for A5 (L1); the full Stage 2 set is indexed under §6.4. Acceptable because A5 spans L1+L2+L4 and the plan limits this stage to the fixture pack.

4. **Roblox Studio path resolution** — §6.4 keeps path-only Verification for `#roblox-studio-path-resolution` (no named *T*). Per review instructions, not treated as a Stage 2 failure; Stage 10 owns plugin create/skip behavior.

## Plan / spec alignment

| Plan Stage 2 RED item | Evidence |
|----------------------|----------|
| Project-file rules (ids, `/`, no `..`, unique studio, non-overlapping disk) | `valid-project-*.json` + tests; invalid fixtures + tests; schema reject for `..` / `\` |
| Source-name table | `source-names.json` + `TestMappingSourceNamesEncodeClassTable_WhenFixtureExists` |
| nodeId examples §6.4 + §11 | `nodeid-examples.json` + nodeId/studioPath tests; `valid-project-example` / `valid-project-demo` |
| Inverse mapping | `inverse-examples.json` + `TestMappingInversePrefersLuau_WhenCreatingFiles` |
| Duplicate `Foo.luau`+`Foo.lua` → `PROJECT_INVALID` | `invalid-duplicate-lua-luau.json` + test |
| At most one `init.*` | `invalid-multiple-init.json` + test |
| Forbidden stems | `invalid-forbidden-stems.json` + test |
| Folder vs init parent | `folder-vs-init-parent.json` + test |
| GREEN: pack validates against project schema | `TestMappingValidProjectsValidate_WhenFixturePackExists` uses compiled schema |
| CHECK: Verification names tests | Chapter diffs in commit |

No scope bleed into mapfs, plugin, or HTTP. No `DEVIATION:` in commit message. No normative spec↔code drift in production code (no mapping runtime in this stage).

## Matrix claim check (stage-scoped)

For each named *T* in stage-touched Verification (§6.4, §11, §02 G7, §09 D9, §12 layers, §14 A5 subset): named tests exist and exercise claimed headings. No stale, missing, or overclaimed index rows for Stage 2 scope. Spec-Anchor lines in tests are one tag per line. `#roblox-studio-path-resolution` remains path-only without a coverage claim on *T* — allowed.

## Tests run

```text
cd sync-protocol && go test ./tests -count=1
ok  	roblox-sync/sync-protocol/tests
```

## Checklist

- [x] Stage diff matches plan Stage 2 (fixtures + L1 tests + CHECK Verification updates)
- [x] RED intent covered (conformance fixture pack, not runtime mapfs)
- [x] GREEN: valid projects validate against `project.schema.json`
- [x] L1 tests pass
- [x] No forbidden sync-tool source read (review scope respected)
- [x] Matrix claim check on stage-touched named *T* — pass
- [x] No blocker anchor violations (D1 placement, one tag per line)
- [x] No `DEVIATION:` without CHECK
- [x] No normative production-code drift for this stage
- [x] A5 L1 fixture pack and G7 shared fixtures — met
- [x] Plugin path-resolution left path-only — not failed per instruction
- [x] Stage scope bleed — none material (`tmp/tasks.md` scratch update only)

## Authorized next

After this PASS: **Stage 3 RED** (`sync-agent/internal/mapfs` vs Stage 2 fixtures) per `proj-spec/implementation-plan.md`.
