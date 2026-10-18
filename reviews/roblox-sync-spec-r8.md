# Review: REVIEW-SPEC — roblox-sync
> **Verdict:** PASS
> **Reviewer model:** composer-2.5-fast
> **Implementer model:** Cursor Grok 4.6
> **Artifact reviewed:** proj-spec/README.md, proj-spec/open-questions.md, proj-spec/chapters/* (r8 after r7 FAIL NOT_FOUND + installer/A11 clarifications)

Independent reviewer: [REVIEW-SPEC r8](e055c2d6-8b22-48b5-8acb-b9a699cec9ce)

## Blockers (must fix before progression)
- *(none)*

## Findings (non-blocking)
- **§6.5 conflict UI — “Keep disk”:** Keep Roblox Studio is wired to `POST /v1/apply` with `force: true`; Keep disk only says the plugin “applies rbx-sync’s current source/class/presence” with no normative source (e.g. `GET /v1/snapshot`, fields on the 409 body, or a defined apply batch). Two compliant implementations could diverge; L4 conflict tests would need arbitrary choices.
- **§6.4 duplicate `Foo.luau` + `Foo.lua`:** “refuse apply (plugin)” is not tied to connect flow, HTTP errors, or plugin UI text—only rbx-sync’s `PROJECT_INVALID` at serve is clear.
- **§10.1 checksum failure:** Step 3 requires refusing install on mismatch; exit code is not stated (A24/L2 imply non-zero). Minor tighten for installer tests.
- **§14 Verification anchors:** Spec-Anchor lines cover Packages, G2/G1, Session safety, and Platform only—no anchors for **Protocol and mapping** or **Out of scope checks** subsections (headings exist).
- **§9 Verification:** D3, D4, D7, and D9 have no Spec-Anchor lines (D1, D2, D5, D6, D8, D10 do).
- **§6.4 Verification:** No Spec-Anchor for **What is synced** or **Line endings and encoding** despite normative MUST/limits there.
- **§6.1 / §7.1 unlisted routes:** `NOT_FOUND` ↔ 404 is aligned in §6.1 and §7.1 Routes; §7.1 Error handling expected tests still emphasize invalid JSON/401—adding an explicit unlisted-path golden would match “each error code in at least one fixture.”
- **§5 file tree:** `proj-spec/implementation-plan.md` (workflow artifact tree) is not listed under “This repository (now)”—acceptable if plan is post–REVIEW-SPEC, but index readers may look for it in §5.
- **§4 stack:** Filesystem watcher is “allowed” but not named; acceptable for greenfield spec, plan stage should pin dependency.

**Historical blockers (verified fixed):**
- **r6:** §6.3 Connect flow steps 6–7 require **Disk wins / Roblox Studio wins / Cancel** (and `POST /v1/bootstrap/apply` for the two wins) before step 8 live; step 8 forbids events/poll/listeners until bootstrap completes.
- **r7:** Unlisted paths: §7.1 HTTP **404** + `NOT_FOUND`; §6.1 maps `NOT_FOUND` → **404** (no `BAD_REQUEST`/404 pair).
- **Clarifications:** §10.1 step 5 — reachable plugins folder + failed `plugin install` → installer MUST NOT exit 0. §14 A11 — unequal trees require bootstrap choice; wins call `POST /v1/bootstrap/apply` before live; A17b does not substitute A17.

## Checklist
- [x] All chapters in README exist
- [x] §14 done-ness consistent with §2 G4/G9
- [x] open-questions blocking none or resolved
- [x] Actor names consistent with §1 (rbx-sync / Plugin / Roblox Studio / Project vs Place)
- [x] Spec-Anchor slugs match headings
- [x] Install paths consistent §5/§6.2/§7.3/§10.1/A22-A24
- [x] A17 requires WSL round-trip; A17b does not substitute
- [x] r6 §6.3 unequal-bootstrap blocker remains resolved
- [x] r7 §6.1/§7.1 unlisted-path status/code pair is consistent (NOT_FOUND → 404)
