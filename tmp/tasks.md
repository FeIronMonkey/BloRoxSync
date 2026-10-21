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
- [x] Stage 13 re-RED (L5 A23) — commit `061c100` / REVIEW-STAGE PASS (`reviews/roblox-sync-stage13-r2.md`)
- [x] Stage 14 — cross-compile, out of scope, acceptance — RED / GREEN / CHECK / commit `4ad5f08` / REVIEW-STAGE PASS (`reviews/roblox-sync-stage14.md`)
- [x] Stage 10 re-RED (L5 A7–A9/A16/A17) — packaged plugin Studio host — RED/GREEN/CHECK `4a1fb7f` / REVIEW-STAGE FAIL (`reviews/roblox-sync-stage10-r2.md` B1) / sibling require `f7f7506` / REVIEW-STAGE PASS (`reviews/roblox-sync-stage10-r3.md`)
- [x] Stage 10 re-RED (L5 A7 DataModel adapter + Keep handlers) — RED/GREEN/CHECK `0e6ccda` / REVIEW-STAGE PASS (`reviews/roblox-sync-stage10-r4.md`)
- [x] Stage 10 re-RED (L5 Connect/Cancel status) — RED/GREEN/CHECK `f1b2be3` / REVIEW-STAGE PASS (`reviews/roblox-sync-stage10-r5.md`)
- [x] Stage 10 re-RED (status label contrast) — RED/GREEN/CHECK `b5f1404` / REVIEW-STAGE PASS (`reviews/roblox-sync-stage10-r6.md`)
- [x] Stage 10 re-RED (L5 A7 nested studio Folder on disk-to-studio) — RED/GREEN/CHECK `1630723` / REVIEW-STAGE PASS (`reviews/roblox-sync-stage10-r7.md`)
- [x] Stage 10 re-RED (L5 A7 live disk events → DataModel) — RED/GREEN/CHECK `2964233` / REVIEW-STAGE PASS (`reviews/roblox-sync-stage10-r8.md`)
- [x] Stage 10 re-RED (L5 A9 studio Name → apply rename without property signal) — RED/GREEN/CHECK `712988d` / REVIEW-STAGE PASS (`reviews/roblox-sync-stage10-r9.md`)

## L5 manual checklist (Stage 14 — required for v1 / REVIEW-FINAL)

- [x] A7 — mapped disk create/update/delete/rename appears in Roblox Studio (live session) — 2026-10-20 `tmp/l5-folder` only (`ServerScriptService.RbxSyncL5`): live create `Live`, update `Hello` (`live update 2`), rename `Live`→`Pong`, delete `Pong`. Nested-root bootstrap + live-event re-RED r7/r8. Do **not** Disk wins against whole `ServerScriptService` / `powercubes.rbxl`.
- [x] A8 — open-tab ScriptEditorService path — 2026-10-20 open `Hello` tab then disk write `print("A8 open tab")` appeared in the open editor (not only closed `Source`)
- [ ] A9 — mapped Studio create/update/delete/rename appears on disk — 2026-10-20 create+update PASS (`StudioA9.server.luau`); **rename FAIL**: Studio Explorer `RbxSyncL5.StudioA9Renamed`, snapshot/disk still `server/StudioA9`. `op: rename` is a move, not a second upsert. Stage 10 re-RED (Name observe without property signal). Stay live on `tmp/l5-folder` only. Do **not** Disk wins against whole `ServerScriptService` / `powercubes.rbxl`.
- [ ] A16 — T1 (Windows rbx-sync) bidirectional mapped-script round-trip — blocked on A9
- [ ] A17 — T2 WSL mirrored (or documented equivalent): health + handshake + bidirectional sync — Windows `GET /v1/health` to WSL `:29852` returned 200; round-trip blocked on A9 / plugin UI
- [x] A18 — real `plugin install` into Windows Plugins folder — 2026-10-19 `rbx-sync plugin install --file` wrote `%LOCALAPPDATA%\Roblox\Plugins\rbx-sync.rbxmx`
- [x] A22 — real Windows one-command download/install — 2026-10-19 `scripts/install.ps1` with `RBX_SYNC_RELEASE_BASE=http://127.0.0.1:8765` wrote `%LOCALAPPDATA%\rbx-sync\` and user PATH
- [x] A23 — real Linux/WSL one-command download/install — 2026-10-19 after LF fix: `sh scripts/install.sh` downloaded from WSL `127.0.0.1:8766`, wrote `~/.local/bin/rbx-sync`, `version` printed `protocol=1`
- [x] §13 — snapshot/bootstrap HTTP within 30 s for ≤ 2000 mapped nodes totaling ≤ 20 MiB — 2026-10-19 2000 nodes / 19.08 MiB: snapshot 18.6 s, compare 16.8–18.5 s, disk-to-studio apply 18.5 s

## Close
- [x] REVIEW-FINAL run (`reviews/roblox-sync-final.md`) — VERDICT FAIL
- [ ] REVIEW-FINAL kick-back — B1 (L5 manual checklist), B2 (Close incomplete); see `reviews/roblox-sync-final.md` — pending
- [ ] REVIEW-FINAL PASS (`reviews/roblox-sync-final[-rN].md`)
