# Review: REVIEW-STAGE Stage 1 — roblox-sync

> **Verdict:** FAIL  
> **Reviewer model:** composer-2.5-fast  
> **Implementer model:** Cursor Grok 4.6  
> **Artifact reviewed:** commit `ed2eb86` vs `proj-spec/implementation-plan.md` Stage 1 and chapters §6.1 / §8 / §4 JSON

Independent reviewer: [REVIEW-STAGE Stage 1](e0a623e5-c98a-4447-918d-8c6979467db2)

## Blockers

### B1 — Matrix claim check: `TestSchemaExistsAsDraft202012_WhenV1SchemaFilesPresent` ↔ §6.1 `#protocolversion`

**Where:** `proj-spec/chapters/06.1-sync-protocol.md` Verification (CHECK in `ed2eb86`).

**Claim:** The index names `TestSchemaExistsAsDraft202012_WhenV1SchemaFilesPresent` and cites `Spec-Anchor: proj-spec/chapters/06.1-sync-protocol.md#protocolversion`.

**Reality:** That test only asserts the three `schema/v1/*.json` files exist, declare `$schema` = JSON Schema draft 2020-12, and compile. It does not read `protocolVersion.txt` or assert `protocolVersion` = 1. Protocol version is covered by `TestProtocolRecordsVersion1_WhenProtocolVersionFileExists`, which is correctly listed separately.

**Rule:** Per `#spec-coverage-audit` matrix claim check (stage-scoped), a named *T* that cites normative heading *H* must exercise *H* (assertions would fail if *H*’s requirement were removed). This row fails that heuristic for `#protocolversion`.

**Kick-back target:** `@spec-driven-tdd` CHECK-only — fix §6.1 Verification (re-anchor or remove the erroneous row; keep draft-2020-12 coverage under §4 `#chosen-stack`, where the test is already correctly indexed). Re-run full REVIEW-STAGE after commit.

---

## Findings

### F1 — RED breadth: `Change` ops beyond `upsert` not positively validated

Plan Stage 1 RED requires Change shapes from §8 to validate. `TestChangeShapeValidates_WhenApplyOpPresent` validates `change-upsert.json` and rejects folder-with-source, but there is no golden or inline instance for `delete` or `rename` ops (distinct required fields in `http.schema.json`). Schema `allOf` branches for those ops are untested by a passing instance. **Severity:** Finding (schema compile + upsert path covered; not normative code drift).

### F2 — Test name vs behavior: `TestRevisionIsJsonNumber_WhenNodeAndEventGoldensExist`

The test asserts `json.Number` on `node-script.json` only; it does not decode `event-upsert.json` for revision typing. Event revision as a JSON number is indirectly guarded by `TestEventShapeValidates_WhenDiskOriginChange` + integer schema on `events.schema.json`. **Severity:** Finding (naming/clarity).

### F3 — §14 A6 table vs Verification wording

AC table row A6 still lists verify column **L3**; §14 Verification correctly notes L1 partial via `TestErrorGoldensCoverEveryCode_WhenHttpFixturesExist`. Consistent with plan Stage 1 (handshake rejection deferred to Stage 5), but the table row could confuse readers until Stage 5. **Severity:** Finding (doc clarity, not stage scope violation).

### F4 — `sync-protocol/go.mod` marks `jsonschema/v6` as `// indirect`

Direct import in `sync-protocol/tests/schema_test.go`; tidy would typically mark it direct. **Severity:** Finding (cosmetic).

---

## Checklist

| Check | Result |
|-------|--------|
| Stage scope: schemas + L1 goldens; no rbx-sync HTTP | Pass |
| Plan RED: `protocolVersion.txt` == 1 | Pass (`TestProtocolRecordsVersion1_WhenProtocolVersionFileExists`) |
| Plan RED: three schemas, JSON Schema 2020-12 | Pass (`TestSchemaExistsAsDraft202012_WhenV1SchemaFilesPresent`) |
| Plan RED: every §6.1 error code in ≥1 `fixtures/v1/http/` golden + status | Pass (11/11; `TestErrorGoldensCoverEveryCode_WhenHttpFixturesExist`) |
| Plan RED: unlisted path 404 + `NOT_FOUND` | Pass (`error-NOT_FOUND.json`, `TestErrorGoldenUsesNotFound404_WhenUnlistedPath`) |
| Plan RED: Node / Event / Root / project shapes validate | Pass |
| Plan RED: Change shapes validate | Partial (upsert only; F1) |
| Plan RED: camelCase keys; revisions as JSON numbers | Pass (`TestWireKeysAreCamelCase_*`, `TestRevisionIsJsonNumber_*` + schema) |
| AC A4 partial (version 1 in fixtures) | Pass |
| AC A6 partial (L1 envelope goldens) | Pass |
| CHECK: Verification names tests in touched chapters | Pass (04, 05, 06.1, 06.4, 07.1, 07.2, 08, 12, 14) |
| No `DEVIATION:` without spec-skill CHECK | Pass (commit message clean) |
| Scope bleed (sync-agent HTTP, plugin, etc.) | Pass |
| Tests green | Pass (`roblox-sync/sync-protocol/tests`; layout cached ok) |
| D1 anchors on new behavioral tests (one tag per line) | Pass (`schema_test.go`) |
| Matrix claim check (stage-scoped index rows) | **Fail (B1)** |
| Normative spec↔code drift in production behavior | N/A (fixtures/schemas only) |

---

## Evidence summary

**Commit:** `ed2eb86` (parent `8a2525a`). Message documents RED/GREEN/CHECK; no `DEVIATION:`.

**Deliverables (GREEN):**

- `sync-protocol/schema/v1/project.schema.json`, `http.schema.json`, `events.schema.json` — draft 2020-12, `$defs` for node/change/root/error goldens and event object.
- `sync-protocol/fixtures/v1/http/` — 11 `error-*.json` goldens (all §6.1 codes), plus shape exemplars (`node-*`, `change-upsert.json`, `event-upsert.json`, `root-handshake.json`, `project-file.json`).
- `sync-protocol/tests/schema_test.go` — 11 L1 tests with method-level `Spec-Anchor` lines.
- `sync-protocol/go.mod` / `go.sum` — Go 1.23 + `github.com/santhosh-tekuri/jsonschema/v6`.

**CHECK (spec):** Verification sections updated to name Stage 1 tests and cite headings; `proj-spec/chapters/12-testing.md` module row now `sync-protocol/tests` covering §6.1, §8, §4 JSON.

**Tests (reviewer re-run):** `ok roblox-sync/sync-protocol/tests` from `C:\repos\roblox-sync\sync-protocol\tests`.

**Traceability highlight:** §4 `#chosen-stack` and `#json-conventions` rows for `TestSchemaExistsAsDraft202012_*`, `TestWireKeysAreCamelCase_*`, and `TestRevisionIsJsonNumber_*` align with test behavior. §6.1 duplicate listing of `TestSchemaExistsAsDraft202012_*` under `#protocolversion` does not.

**Not in scope (correctly deferred):** §6.1 `#auth`, handshake mismatch (Stage 5), mapping fixture pack (Stage 2), L3 httptest routes.

---

**Summary:** Implementation and L1 tests match Stage 1 plan intent and pass locally. **FAIL** solely because §6.1 Verification overclaims coverage of `#protocolversion` for a schema-draft test — a `#spec-coverage-audit` matrix claim **Blocker** (B1). Remediation is CHECK-class index fix, then full REVIEW-STAGE re-review.
