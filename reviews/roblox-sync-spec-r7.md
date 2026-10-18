# Review: REVIEW-SPEC — roblox-sync
> **Verdict:** FAIL
> **Reviewer model:** composer-2.5-fast
> **Implementer model:** Cursor Grok 4.6
> **Artifact reviewed:** proj-spec/README.md, proj-spec/open-questions.md, proj-spec/chapters/* (r7 after r6 FAIL §6.3 connect-flow fix)

Independent reviewer: [REVIEW-SPEC r7](f2b9caca-27c4-41db-a271-b40745000c1d)

## Blockers (must fix before progression)

- **§7.1 vs §6.1 — HTTP status for `BAD_REQUEST` on unknown routes.** [§7.1 Routes](proj-spec/chapters/07.1-http-api.md) requires unlisted paths to return HTTP **404** with envelope `error.code` **`BAD_REQUEST`**. [§6.1 Error codes](proj-spec/chapters/06.1-sync-protocol.md) normatively maps **`BAD_REQUEST` → HTTP 400** (body/schema failures). The same §7.1 chapter also states under Error handling that “Statuses follow §6.1,” which contradicts the 404 rule in Routes. L3 httptest goldens and both runtimes cannot satisfy the table in §6.1 and the unlisted-path rule in §7.1 without an explicit carve-out or a dedicated code (e.g. not-found) added to §6.1.

## Findings (non-blocking)

- **r6 §6.3 unequal-bootstrap fix is present.** Connect flow now gates live on step 8; step 7 defines **Cancel** (no `bootstrap/apply`, no live), **Disk wins** (`POST /v1/bootstrap/apply` `disk-to-studio`), and **Roblox Studio wins** (same step-5 snapshot in `nodes`, `studio-to-disk`), consistent with §6.5, §7.1, §11 example step 4, and A11 intent.
- **§10.1 one-command exit when `plugin install` fails:** Step 5 only mandates exit **0** when the plugins folder is **not** reachable. When the folder **is** reachable and `rbx-sync plugin install` exits **1** (§7.3 / §6.2), the installer script’s exit code is still unspecified — weak for verifying G9 “install plugin in the same command when reachable.”
- **§14 A11 wording:** Criterion text emphasizes Cancel writes nothing; Disk / Roblox Studio wins outcomes (must call `bootstrap/apply` before live) are implied via L3 + plugin L4 and §6.3, not stated in the A11 row itself.
- **§13 snapshot/bootstrap 30 s MUST** (≤2000 nodes, ≤20 MiB) has no §14 acceptance ID; done-ness is not tied to that requirement.
- **Topology T3** is documented in §10.2; §14 A17 covers T2 (mirrored / references.md equivalent) only — T3 has no acceptance row (may be intentional).
- **§5 file tree:** `implementation-plan.md` and post–Stage 0 package contents remain under “target after implementation” only (expected pre-plan; minor index gap).
- **§9 section order:** D10 appears before D9 (cosmetic).
- **Actor terminology:** No leftover “the agent” / “game repository” actor usage; rbx-sync / Plugin / Roblox Studio / Project vs Place read consistently with §1.
- **Spec-Anchor slugs (sampled):** Headings align with Verification lines (e.g. §6.3 `#connect-flow`, §6.4 `#inverse-mapping-roblox-studio--disk`, §14 `#disk--roblox-studio-g2` / `#roblox-studio--disk-g1`, §5 `#installed-files-user-machine-after-101`).

## Checklist

- [x] All chapters in README exist
- [x] §14 done-ness consistent with §2 G4/G9
- [x] open-questions blocking none or resolved
- [x] Actor names consistent with §1 (rbx-sync / Plugin / Roblox Studio / Project vs Place)
- [x] Spec-Anchor slugs match headings after rename
- [x] Install paths consistent §5/§6.2/§7.3/§10.1/A22-A24
- [x] A17 requires WSL round-trip; A17b does not substitute
- [x] r6 §6.3 unequal-bootstrap blocker resolved (Cancel stops; Disk wins and Roblox Studio wins call bootstrap/apply before live)
