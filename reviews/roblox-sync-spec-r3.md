# Review: REVIEW-SPEC — roblox-sync
> **Verdict:** PASS
> **Reviewer model:** composer-2.5-fast
> **Implementer model:** Cursor Grok 4.6
> **Artifact reviewed:** proj-spec/chapters/*, proj-spec/README.md, proj-spec/open-questions.md (re-review r3 after monorepo layout revision; prior PASS reviews/roblox-sync-spec-r2.md superseded)

Independent reviewer: [REVIEW-SPEC re-review r3](f4f2c21f-c08d-4834-99a1-0d9b1785d06e)

## Blockers (must fix before progression)
- none

## Findings (non-blocking)
- **Monorepo layout (verified):** §3.1, §9 D2, §14 A1, root `README.md` / `AGENTS.md`, and package `README.md` / `AGENTS.md` consistently describe **one git repository** with `sync-protocol`, `sync-agent`, and `studio-plugin` as **packages** (not nested remotes/submodules). No nested `.git` under packages; root `.gitignore` does not exclude package trees; game repo remains separate per §3.1 / §5.
- **Leftover “repo” wording:** §2 G7 (“protocol **repo**”), §6.1 expected tests (“protocol **repo**”), and G8 (“each **repo** `README.md`”) read like multi-repo language; normative layout is **package in this repo** (§3.1). Clarify wording in a future CHECK; not a structural contradiction.
- **§6.3 connect flow vs §6.5 / §11:** Steps 6→8 do not explicitly halt on **Cancel**, call **`POST /v1/bootstrap/apply`** after Disk/Studio wins, or gate step 8 until bootstrap completes. §6.5 lifecycle and §11 example supply the missing behavior; implementers must not treat step 8 as unconditional after step 6.
- **§1 overview:** “After a confirmed bootstrap…” implies bootstrap always runs; **equal** path skips bootstrap UI/writes (§6.5, A11b).
- **§4 stack vs §6.2:** Serve-time “validate `rbx-sync.json` against protocol schema” and L1 JSON Schema tests are not matched by an explicit allowed JSON Schema validation dependency for the agent (§4 only names stdlib HTTP/JSON plus an optional watcher library). Plan stage should pin the validator or narrow validation rules.
- **Verification / Spec-Anchor coverage:** Normative subsections still unlisted in some Verification blocks (e.g. §6.3 construction/HTTP, §6.5 disconnect, §7.3 error-handling, §10.1 health check). All listed `Spec-Anchor` slugs were checked with **github-slugger@2.0.0** (0 mismatches), including §3.1 `#packages`, §5 file-tree headings, §9 D2 `#d2--one-git-repository-three-packages`, and §14 `#disk--studio-g2` / `#studio--disk-g1`.
- **§7.2 vs §7.1 `since` defaults:** Events default to no backlog; `/v1/changes` defaults `since=0` — documented and intentional (§7.2).

## Checklist
- [x] All chapters present for greenfield (incl §14)
- [x] README index links every chapter
- [x] open-questions blocking entries resolved or Blocking: none
- [x] No default/port/path contradictions
- [x] Acceptance criteria testable
- [x] Error behavior specified
- [x] No orphan requirements
- [ ] Verification + Spec-Anchor headings present (listed anchors valid; some normative headings still omitted from Verification indexes)
- [x] Layout language consistent (one git repo, three packages; no nested remotes)
