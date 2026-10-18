# Review: REVIEW-PLAN — roblox-sync
> **Verdict:** FAIL
> **Reviewer model:** composer-2.5-fast
> **Implementer model:** Cursor Grok 4.6
> **Artifact reviewed:** proj-spec/implementation-plan.md vs proj-spec/chapters/* (spec r8)

Independent reviewer: [REVIEW-PLAN](04dc373b-ecf8-4803-bf2c-3f70d390b384)

## Blockers (must fix before progression)

- **Stage 0 RED vs A2 (§14, §3.1):** Stage 0 claims A2 but RED only requires `AGENTS.md` + `README.md` on the **three package directories**. The spec and A2 require the **repository root** plus each package. Without a root-level layout test, Stage 0 GREEN can pass while A2 fails.

- **Stage 0 RED vs A3 (§14):** Stage 0 lists **AC A3** (`references.md` lists allowed research sources) but the RED step never names a failing test or assertion for `references.md` (existence, minimum structure, or allowlist content). A3 is mapped to a stage with no verifiable RED criterion.

- **§6.2 `IO_ERROR` MUST with no L3 stage:** [§6.2 sync-agent](../proj-spec/chapters/06.2-sync-agent.md#error-handling) requires surfacing **`IO_ERROR`** on snapshot/apply when mapped roots are unreadable. Stage 1 can place an **`IO_ERROR`** golden in L1 fixtures, but no stage (5–7 serve tests) assigns an **L3 httptest RED** that forces unreadable roots and expects **500 + `IO_ERROR`**. That leaves a normative runtime MUST without a TDD hook; it is not covered by “every code in a golden” alone.

## Findings (non-blocking)

- **`NOT_MAPPED` (§6.1):** Stage 1 requires every error `code` in at least one L1 golden, but no stage RED names **L3 apply/snapshot** behavior for `nodeId` outside configured roots → **400 `NOT_MAPPED`**.

- **§7.3 bind-in-use:** Exit **1** when the listen address is in use is normative in [§7.3](../proj-spec/chapters/07.3-cli.md#error-handling); Stage 4 RED covers bind host/port validation and token printing but not **address already in use**.

- **§6.4 Roblox Studio path resolution (plugin):** [§6.4](../proj-spec/chapters/06.4-project-mapping.md#roblox-studio-path-resolution) requires skipping unresolvable roots on studio-to-disk snapshot and surfacing a **plugin warning**. Stages 9–10 RED focus on fixtures and connect/conflict flow; this rule has no named L4 RED.

- **§13 plugin-side 50 ms debounce:** [§13](../proj-spec/chapters/13-performance.md#limits) and [§6.3](../proj-spec/chapters/06.3-studio-plugin.md#roblox-studio-change-detection) require **50 ms** batching before `POST /v1/apply`. Stage 8 covers **disk** debounce; Stage 10 does not name plugin mutation debounce in RED.

- **§6.5 plugin echo suppression:** [§6.5](../proj-spec/chapters/06.5-sync-session.md#echo-suppression) **self-apply ignore** on the plugin is not assigned in Stage 10 (only rbx-sync self-write in Stage 8).

- **§10.1 release assets beyond CI cross-build:** Stages 12–14 cover packaging, installer checksum tests, and **windows/amd64 + linux/amd64** builds. The plan does not explicitly stage **release artifact layout** (`SHA256SUMS` generation for the three published assets, tagged release publish) as RED/GREEN; Stage 13 tests checksum refusal with a fake base URL only.

- **Stage order 11 → 12:** `plugin install` (11) precedes packaging (12). Mitigated by testdata `.rbxmx` in Stage 11 and Stage 13 after 12 for user install; contributor path in §10.1 is “build binary → package → plugin install,” which is the reverse of 11/12 (acceptable if documented as intentional test ordering).

- **REVIEW-SPEC r8 carryover — Keep disk:** Stage 10 tests observable parity with `GET /v1/snapshot` after Keep disk; plan correctly forbids inventing a new route. Residual spec ambiguity remains; plan handling is consistent with r8 non-blocking finding.

- **Plan skill vs artifact:** `@spec-driven-plan` expects `tmp/tasks.md` initialized at plan time; the durable plan defers L5 checklist to Stage 14 CHECK only. Not a spec violation but easy to miss before TDD.

- **§13 snapshot/bootstrap 30 s / 2000 nodes:** Normative limit in §13 has no stage RED (likely manual or post-v1 perf); acceptable gap if intentional, but not named as out-of-scope in the plan.

## Checklist

- [x] Every stage has RED, GREEN, CHECK — Stages 0–14 each define all three in **Stage details**; summary table states **RED → GREEN → CHECK → commit → REVIEW-STAGE**.
- [x] Every stage names a test module — Stages 0–14 name modules (Stage 8 combines `internal/watch` + serve events; Stage 12/13/14 use descriptive L2 labels).
- [x] Stage 0 is scaffolding (greenfield) — Stage 0 is **Scaffolding** with layout, `go.mod`, stub `main`, `protocolVersion.txt`, CI hook; matches greenfield Stage 0 semantics.
- [x] Protocol/fixtures before Go mapfs and Luau mapping — Stages **1–2** before **3** and **9**.
- [x] Bind/CLI before HTTP routes that need serve — Stage **4** before **5**.
- [x] Bootstrap/apply/watch cover §6.5 + §7.1/§7.2 — Stages **6–8** (bootstrap, live apply, watch/SSE, poll); HTTP routes in **5–7**.
- [x] Plugin connect flow (equal, disk wins, studio wins, cancel) is a named stage — **Stage 10** RED enumerates all four branches plus live gating.
- [x] Installers + checksum A24 have a stage — **Stage 13** (A24 non-zero on mismatch).
- [x] Cross-compile windows/amd64 + linux/amd64 has a stage — **Stage 14** RED.
- [ ] A1–A24 each mapped to at least one stage — All IDs appear in stage **AC** lines or Stage 14 aggregate **except verification strength issues:** **A2/A3** claimed on Stage 0 but RED incomplete; **A8** manual only via Stage 14 “A7–A9 Studio” (implicit, not named A8).
- [x] REVIEW-STAGE after each stage; REVIEW-FINAL after last — Stated in **Stages** intro and Stage 14 handoff to REVIEW-FINAL.
- [x] Blocking open-questions: none or listed — OQ-001–OQ-003 **Blocking: none**; plan repeats this.
