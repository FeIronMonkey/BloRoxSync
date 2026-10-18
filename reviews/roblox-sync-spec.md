# Review: REVIEW-SPEC — roblox-sync
> **Verdict:** FAIL
> **Reviewer model:** composer-2.5-fast
> **Implementer model:** Cursor Grok 4.6
> **Artifact reviewed:** proj-spec/chapters/*, proj-spec/README.md, proj-spec/open-questions.md, README.md, AGENTS.md, references.md, nested repo README/AGENTS

Independent reviewer: [REVIEW-SPEC roblox-sync](3f0f5e3d-73f1-484b-87d3-9e6c2654c2cb)

## Blockers (must fix before progression)
- **§7.1 internal contradiction on `POST /v1/bootstrap/apply` success body:** The routes table lists success as `200 { "applied": [...] }` for all bootstrap/apply calls, but the `disk-to-studio` row says the success body is the **disk snapshot** the plugin must apply (same shape as `GET /v1/snapshot` / `{ "nodes": [...] }`, not `{ "applied": [...] }`). Goldens, L3 tests, and the plugin cannot implement one normative shape without picking a side.
- **Revision state undefined when compare returns `equal: true`:** §6.3 connect flow skips bootstrap UI when trees are equal; §6.5 assigns `revision` only “after bootstrap.” There is no rule for initializing agent/plugin revision tables on the equal path before live `POST /v1/apply` / SSE (e.g. all nodes at `1`, or snapshot + implicit sync). Live apply and A12 conflict tests are not fully specifiable.
- **Bind-host contradiction:** §3 **Processes** table says the agent “binds `127.0.0.1` only,” while §3 **Trust boundary**, §6.2, and §7.3 `--bind` explicitly allow **`::1`** and reject other hosts. Acceptance A14 (“bind host not loopback is rejected”) does not clarify whether `::1` is supported end-to-end (including plugin default host `127.0.0.1` only in §6.3).

## Findings (non-blocking)
- **§5 file tree gaps vs §12:** `sync-protocol` has no documented `tests/` (or equivalent) path for L1 schema/golden validation; `sync-agent` lists `internal/serve`, `watch`, etc. but not HTTP/router layout expected by L3.
- **§6.5 revision wording:** “revision **1** (or increment from 0)” after bootstrap is ambiguous.
- **§7.1 routes vs §7.2 `since` defaults:** Events default “current (no backlog)”; poll `/v1/changes` defaults `since=0`.
- **§14 / equal-bootstrap gap:** A11 covers unequal trees only.
- **§14 A17:** T2 “fail closed with a clear error” not tied to a normative message.
- **Verification / Spec-Anchor coverage:** several error-handling headings unanchored.

## Checklist
- [x] All chapters present for greenfield (incl §14)
- [x] README index links every chapter
- [x] open-questions blocking entries resolved or Blocking: none
- [ ] No default/port/path contradictions
- [ ] Acceptance criteria testable
- [ ] Error behavior specified
- [x] No orphan requirements
- [ ] Verification + Spec-Anchor headings present
