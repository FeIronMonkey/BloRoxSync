# Review: REVIEW-SPEC — roblox-sync
> **Verdict:** FAIL
> **Reviewer model:** composer-2.5-fast
> **Implementer model:** Cursor Grok 4.6
> **Artifact reviewed:** proj-spec/README.md, proj-spec/open-questions.md, proj-spec/chapters/* (r6 after actor terminology revision)

Independent reviewer: [REVIEW-SPEC r6](f9353fab-cafa-40b0-abba-e30dea5b08bd)

## Blockers (must fix before progression)

- **§6.3 Connect flow contradicts bootstrap/session requirements (unequal trees).** Steps 5–6 run `POST /v1/bootstrap/compare` and, if unequal, show **Disk wins / Roblox Studio wins / Cancel**. Step 7 applies only when `equal: true` (`GET /v1/snapshot`). Steps 8–9 then **always** open the live stream and start Roblox Studio listeners. There is no numbered step for: **Cancel → stop (no live, no writes)**; **Disk wins → `POST /v1/bootstrap/apply` `{ "direction": "disk-to-studio" }` then apply returned nodes and store revisions**; **Roblox Studio wins → `POST /v1/bootstrap/apply` `{ "direction": "studio-to-disk", "nodes": [...] }` then store revisions**. That directly conflicts with §6.5 session lifecycle (bootstrap before live; Cancel does not enter live), §7.1 bootstrap/apply contracts, §11 example step 4, and §14 **A11** (unequal trees require explicit bootstrap; Cancel writes nothing). An implementer following §6.3 literally would enter live sync without bootstrap after an unequal compare—unacceptable for v1.

## Findings (non-blocking)

- **§6.3 vs §6.5 / §7.1 (Roblox Studio wins detail):** Even outside the numbered-list bug, §6.3 never states that the Roblox Studio snapshot built at compare must be resent as `nodes` on `studio-to-disk` bootstrap/apply (§6.5 revision table and §7.1 do). Recoverable only by cross-reading; connect flow should name it.
- **§10.1 one-command exit codes when plugin leg runs:** Step 5 mandates exit **0** when the plugins folder is **not** reachable (binary-only success). When the folder **is** reachable and `rbx-sync plugin install` exits **1** (§7.3 / §6.2), the installer script exit code is **unspecified**—weak for **G9** “install plugin in the same command when reachable.”
- **§13 snapshot/bootstrap 30 s MUST** for ≤2000 nodes has no §14 acceptance ID; done-ness is not tied to that requirement (only debounce/size partial coverage elsewhere).
- **Topology T3:** §10.2 documents T3; §14 **A17** requires T2 mirrored (or references.md equivalent) only—intentional but T3 has no acceptance row.
- **§6.3 connect flow (equal path):** Steps 6–8 ordering is still awkward (unequal branch has no bootstrap HTTP between 6 and 8 even after a fix must insert steps); §11 remains the best end-to-end unequal walkthrough.
- **§5 file tree:** `proj-spec/implementation-plan.md` and post–Stage 0 package contents are not listed under “this repository (now)” (expected later; minor index gap).
- **§9 section order:** D10 appears before D9 in the file (cosmetic only).
- **§06.3 / §06.4 Verification coverage:** Connect flow, Roblox Studio change detection, and HTTP subsections lack Spec-Anchor lines (not required for every heading, but connect flow is high-risk after the blocker fix). *(Connect flow already has a Spec-Anchor line; change-detection / HTTP still lack them.)*
- **Prior r5 PASS superseded:** Actor rename is largely clean (no “the agent” / “game repository” actor usage found; **Roblox Studio** / **rbx-sync** / **Plugin** / **Project** vs **Place** read consistently). Wire keys `agentVersion`, `studio`, `origin: studio` match §1 disclaimer.

## Checklist

- [x] All chapters in README exist
- [x] §14 done-ness consistent with §2 G4/G9
- [x] open-questions blocking none or resolved
- [x] Actor names consistent with §1 (rbx-sync / Plugin / Roblox Studio / Project vs Place)
- [x] Spec-Anchor slugs match headings after rename (sampled: §5 `#installed-files-user-machine-after-101` ↔ “after §10.1”; §14 `#disk--roblox-studio-g2` / `#roblox-studio--disk-g1` use GitHub-style arrow/em-dash slugs; no systematic slug drift found)
- [x] Install paths consistent §5/§6.2/§7.3/§10.1/A22-A24
- [x] A17 requires WSL round-trip; A17b does not substitute
