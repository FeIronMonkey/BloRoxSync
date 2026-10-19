# Review: REVIEW-STAGE Stage 1 — roblox-sync

> **Verdict:** PASS  
> **Reviewer model:** composer-2.5-fast  
> **Implementer model:** Cursor Grok 4.6  
> **Artifact reviewed:** commits `ed2eb86` + `4e5c462` (B1 CHECK) vs plan Stage 1; prior FAIL `reviews/roblox-sync-stage1.md`

Independent reviewer: [REVIEW-STAGE Stage 1 r2](b9b97098-50a1-4fd8-af34-3a554a9c2a43)

Re-review after prior FAIL (B1 only). Independent reviewer pass; READ-ONLY.

## Blockers

**None.**

### B1 (prior) — Resolved

**Was:** §6.1 Verification named `TestSchemaExistsAsDraft202012_WhenV1SchemaFilesPresent` with `Spec-Anchor: …#protocolversion` while that test only checked draft 2020-12 and schema compile.

**Fix (`4e5c462`):** Removed the §6.1 Verification row and the matching `// Spec-Anchor: …#protocolversion` line on `TestSchemaExistsAsDraft202012_WhenV1SchemaFilesPresent`. `#protocolversion` remains indexed only on `TestProtocolRecordsVersion1_WhenProtocolVersionFileExists` (Verification + test anchors).

**Confirm:** `TestProtocolRecordsVersion1_WhenProtocolVersionFileExists` reads `protocolVersion.txt` and asserts trimmed content `"1"` — satisfies the matrix exercise heuristic for `#protocolversion`. `TestSchemaExistsAsDraft202012_*` is indexed under §4 `#chosen-stack` and §5 `#sync-protocol-target-after-implementation` only, matching behavior.

---

## Findings

Non-blocking; carried from prior review unless noted.

### F1 — RED breadth: `Change` ops beyond `upsert`

`TestChangeShapeValidates_WhenApplyOpPresent` validates upsert golden and rejects folder-with-source; no passing golden for `delete` or `rename` branches in `http.schema.json`. **Severity:** Finding (upsert + compile path covered; not normative production drift).

### F2 — Test name vs behavior: `TestRevisionIsJsonNumber_WhenNodeAndEventGoldensExist`

Asserts `json.Number` on `node-script.json` only; does not decode `event-upsert.json`. Event revision typing is indirectly guarded by schema + `TestEventShapeValidates_WhenDiskOriginChange`. **Severity:** Finding (naming/clarity).

### F3 — §14 A6 table vs Verification wording

AC table still lists verify column **L3** for A6; §14 Verification correctly notes L1 partial via error goldens. Consistent with plan (handshake/HTTP rejection in Stage 5). **Severity:** Finding (doc clarity).

### F4 — `sync-protocol/go.mod` marks `jsonschema/v6` as `// indirect`

Direct import in `schema_test.go`; `go mod tidy` would typically mark it direct. **Severity:** Finding (cosmetic).

---

## Checklist

| Check | Result |
|-------|--------|
| Stage scope: schemas + L1 goldens; no rbx-sync HTTP | Pass |
| Plan RED: `protocolVersion.txt` == 1 | Pass |
| Plan RED: three schemas, JSON Schema 2020-12 | Pass |
| Plan RED: every §6.1 error code in ≥1 HTTP golden + status | Pass |
| Plan RED: unlisted path 404 + `NOT_FOUND` | Pass |
| Plan RED: Node / Event / Root / project shapes validate | Pass |
| Plan RED: Change shapes validate | Partial (upsert only; F1) |
| Plan RED: camelCase keys; revisions as JSON numbers | Pass |
| AC A4 partial (version 1) | Pass |
| AC A6 partial (L1 envelope goldens) | Pass |
| CHECK: Verification names Stage 1 tests in touched chapters | Pass |
| No `DEVIATION:` without spec-skill CHECK | Pass |
| Scope bleed (sync-agent HTTP, plugin, etc.) | Pass |
| Tests green | Pass |
| D1 anchors on new behavioral tests (one tag per line) | Pass |
| Matrix claim check (stage-scoped, named *T* rows) | **Pass** (B1 fixed) |
| Normative spec↔code drift in production behavior | N/A (fixtures/schemas only) |

---

## Evidence summary

- **PASS** after B1 CHECK (`4e5c462`). Prior FAIL: `reviews/roblox-sync-stage1.md`.
- Tests: `ok roblox-sync/sync-protocol/tests` (11 tests).
- Non-blocking F1–F4 remain for later stages / tidy; they do not block Stage 2.

VERDICT: PASS
