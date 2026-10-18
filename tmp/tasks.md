# Roblox Sync — Task List

## Commit strategy
Greenfield — spec phase only: one commit per logical spec milestone; spec before code. One git repository; `studio-plugin`, `sync-agent`, and `sync-protocol` are packages (directories), not nested remotes.

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
- [ ] Plan written
- [ ] REVIEW-PLAN PASS (`reviews/roblox-sync-plan.md`)
- [ ] User approved plan gate — YYYY-MM-DD

## Implementation / change stages
- [ ] Stage 0 — scaffolding — REVIEW-STAGE PASS

## Close
- [ ] REVIEW-FINAL run (`reviews/roblox-sync-final.md`)
- [ ] REVIEW-FINAL kick-back — <blocker IDs>; remedial CHECK per FAIL review — pending | done YYYY-MM-DD
- [ ] REVIEW-FINAL PASS (`reviews/roblox-sync-final.md`)
