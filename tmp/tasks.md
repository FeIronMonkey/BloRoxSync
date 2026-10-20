# Roblox Sync — Task List

## Commit strategy
Greenfield — implementation: one commit per plan stage; RED / GREEN / CHECK in the message. Spec milestone already committed (`5f42a6d`). Plan commit after REVIEW-PLAN PASS, before Stage 0 TDD.

Durable plan: `proj-spec/implementation-plan.md`

## Spec phase (greenfield only)
- [x] Chapters written
- [x] README index
- [x] `references.md` (research allowlist)
- [x] Package directories created (`studio-plugin`, `sync-agent`, `sync-protocol`) — tracked in this repo
- [x] Actor terminology revision (Project / rbx-sync / Roblox Studio) — 2026-10-18; r5 PASS superseded
- [x] REVIEW-SPEC r6 FAIL (`reviews/roblox-sync-spec-r6.md`) — §6.3 connect flow skipped unequal bootstrap apply
- [x] REVIEW-SPEC r7 FAIL (`reviews/roblox-sync-spec-r7.md`) — §7.1 404 used `BAD_REQUEST` vs §6.1 400
- [x] REVIEW-SPEC PASS (`reviews/roblox-sync-spec-r8.md`; r7 FAIL `reviews/roblox-sync-spec-r7.md`; r6 FAIL `reviews/roblox-sync-spec-r6.md`; r5 PASS `reviews/roblox-sync-spec-r5.md` superseded by actor terminology revision)
- [x] User approved spec gate — 2026-10-18

## Plan phase
- [x] Plan written: `proj-spec/implementation-plan.md`
- [x] REVIEW-PLAN FAIL (`reviews/roblox-sync-plan.md`) — Stage 0 A2/A3 RED incomplete; no L3 `IO_ERROR`
- [x] REVIEW-PLAN PASS (`reviews/roblox-sync-plan-r2.md`; r1 FAIL `reviews/roblox-sync-plan.md`)
- [x] User approved plan gate — 2026-10-18

## Implementation / change stages
- [x] Stage 0 — scaffolding — RED / GREEN / CHECK / commit / REVIEW-STAGE PASS (`reviews/roblox-sync-stage0.md`, commit `075e523`)
- [x] Stage 1 — protocol schemas — RED / GREEN / CHECK / commit `ed2eb86` / REVIEW-STAGE FAIL (`reviews/roblox-sync-stage1.md` B1) / CHECK `4e5c462` / REVIEW-STAGE PASS (`reviews/roblox-sync-stage1-r2.md`)
- [x] Stage 2 — mapping fixtures — RED / GREEN / CHECK / commit `b48f005` / REVIEW-STAGE PASS (`reviews/roblox-sync-stage2.md`)
- [x] Stage 3 — Go mapfs — RED / GREEN / CHECK `ebcba4b` / REVIEW-STAGE FAIL (`reviews/roblox-sync-stage3.md` B1) / CHECK `1e13d40` / REVIEW-STAGE PASS (`reviews/roblox-sync-stage3-r2.md`)
- [x] Stage 4 — CLI, bind, token — RED / GREEN / CHECK / commit `587a999` / REVIEW-STAGE PASS (`reviews/roblox-sync-stage4.md`)
- [x] Stage 5 — HTTP core — RED / GREEN / CHECK / commit `3c525ea` / REVIEW-STAGE FAIL (`reviews/roblox-sync-stage5.md` B1/B2) / CHECK `c085a24` / REVIEW-STAGE PASS (`reviews/roblox-sync-stage5-r2.md`)
- [x] Stage 6 — bootstrap — RED / GREEN / CHECK `670922e` / REVIEW-STAGE FAIL (`reviews/roblox-sync-stage6.md` B1) / CHECK `a10e44c` / REVIEW-STAGE PASS (`reviews/roblox-sync-stage6-r2.md`)
- [x] Stage 7 — live apply and conflicts — RED / GREEN / CHECK / commit `be6b318` / REVIEW-STAGE PASS (`reviews/roblox-sync-stage7.md`)
- [x] Stage 8 — watch, SSE, poll — RED / GREEN / CHECK `c8b7580` / REVIEW-STAGE FAIL (`reviews/roblox-sync-stage8.md` B1) / CHECK `5d4f9fb` / REVIEW-STAGE PASS (`reviews/roblox-sync-stage8-r2.md`)
- [x] Stage 9 — plugin mapping (Luau) — RED / GREEN / CHECK / commit `0264fdd` / REVIEW-STAGE PASS (`reviews/roblox-sync-stage9.md`)
- [x] Stage 10 — plugin session — RED / GREEN / CHECK / commit `1b6f287` / REVIEW-STAGE PASS (`reviews/roblox-sync-stage10.md`)
- [x] Stage 11 — plugin install helper — RED / GREEN / CHECK / commit `9f199af` / REVIEW-STAGE PASS (`reviews/roblox-sync-stage11.md`)
- [x] Stage 12 — plugin packaging — RED / GREEN / CHECK / commit `7ca9c59` / REVIEW-STAGE PASS (`reviews/roblox-sync-stage12.md`)
- [x] Stage 13 — one-command installers — RED / GREEN / CHECK / commit `a315a7a` / REVIEW-STAGE PASS (`reviews/roblox-sync-stage13.md`)
- [x] Stage 14 — cross-compile, out of scope, acceptance — RED / GREEN / CHECK / commit `4ad5f08` / REVIEW-STAGE PASS (`reviews/roblox-sync-stage14.md`)

## L5 manual checklist (Stage 14 — required for v1 / REVIEW-FINAL)

- [ ] A7 — mapped disk create/update/delete/rename appears in Roblox Studio (live session)
- [ ] A8 — open-tab ScriptEditorService path
- [ ] A9 — mapped Studio create/update/delete/rename appears on disk
- [ ] A16 — T1 (Windows rbx-sync) bidirectional mapped-script round-trip
- [ ] A17 — T2 WSL mirrored (or documented equivalent): health + handshake + bidirectional sync
- [ ] A18 — real `plugin install` into Windows Plugins folder
- [ ] A22 — real Windows one-command download/install
- [ ] A23 — real Linux/WSL one-command download/install
- [ ] §13 — snapshot/bootstrap HTTP within 30 s for ≤ 2000 mapped nodes totaling ≤ 20 MiB

## Close
- [x] REVIEW-FINAL run (`reviews/roblox-sync-final.md`) — VERDICT FAIL
- [ ] REVIEW-FINAL kick-back — B1 (L5 manual checklist), B2 (Close incomplete); see `reviews/roblox-sync-final.md` — pending
- [ ] REVIEW-FINAL PASS (`reviews/roblox-sync-final[-rN].md`)
