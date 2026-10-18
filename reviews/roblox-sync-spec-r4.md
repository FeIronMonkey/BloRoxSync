# Review: REVIEW-SPEC — roblox-sync
> **Verdict:** FAIL
> **Reviewer model:** composer-2.5-fast
> **Implementer model:** Cursor Grok 4.6
> **Artifact reviewed:** proj-spec/README.md, proj-spec/open-questions.md, proj-spec/chapters/*

Independent reviewer: [REVIEW-SPEC r4](e16c4abc-6fc1-41f8-9b84-d5bbab6bfa39)

## Blockers (must fix before progression)
- **§2 success definition vs A17 (WSL done-ness):** [§2](../proj-spec/chapters/02-goals-and-non-goals.md) states v1 is done when §14 is met **and** a Windows user can **round-trip mapped scripts in both directions with the agent on WSL** (mirrored networking or documented equivalent). [A17](../proj-spec/chapters/14-acceptance-criteria.md) allows **manual pass** when T2 yields only **“agent not reachable”** (health failure) with **no successful handshake and no sync**. §14 also says v1 is done when A1–A24 hold. A reviewer can mark §14 complete without any WSL round-trip, which directly contradicts §2/G4’s WSL success story. Tighten A17 (require successful health + session + bidirectional sync on a documented T2 setup) or narrow §2’s WSL requirement to match the disjunct.

## Findings (non-blocking)
- **Install artifact staging path:** [§10.1](../proj-spec/chapters/10.1-install.md) requires downloading `rbx-sync.rbxmx` but does not normatively say where it is placed before `rbx-sync plugin install`. [§7.3](../proj-spec/chapters/07.3-cli.md) `--file` default (“next to the binary or documented share path”) is undefined for the one-command layout after agent install to `%LOCALAPPDATA%\rbx-sync\` or `$HOME/.local/bin/`.
- **§5 file tree gap:** [§5](../proj-spec/chapters/05-file-tree.md) lists `scripts/install.ps1` and `install.sh` under “target after implementation” only; it does not document post-install locations for the plugin artifact (only the Studio Plugins destination). Not wrong for greenfield, but implementers lack a single normative on-disk layout for release files beside the agent binary.
- **§13 limits without §14 hook:** [§13](../proj-spec/chapters/13-performance.md) requires snapshot/bootstrap HTTP to complete within 30 s for ≤2000 nodes / ≤20 MiB; no A-id or L2/L3 test row in §12/§14 encodes that requirement (orphan normative limit).
- **WSL plugin path placeholders:** [§6.2](../proj-spec/chapters/06.2-sync-agent.md) uses `/mnt/c/Users/<user>/…`; [§10.1](../proj-spec/chapters/10.1-install.md) uses `<windowsUser>`. No spec algorithm for resolving the correct Windows profile from WSL (risk for G9 plugin leg on T2/T3).
- **Partial CONFLICT UX:** [§7.1](../proj-spec/chapters/07.1-http-api.md) allows 409 `CONFLICT` with partial `applied` and `conflicts`; [§6.3](../proj-spec/chapters/06.3-studio-plugin.md) conflict handling does not describe behavior for mixed-success batches.
- **Index chapters:** [§6](../proj-spec/chapters/06-components.md), [§7](../proj-spec/chapters/07-external-api.md), and [§10](../proj-spec/chapters/10-operations.md) have no Verification subsections (leaf behavioral chapters do).
- **G9 acceptance mapping:** G9 is covered indirectly via A22–A23 and §10.1; there is no explicit “G9 satisfied” A-row tying Windows **and** Linux/WSL one-command flows including plugin leg when reachable.

## Checklist
- [x] All chapters in README exist on disk
- [x] §14 exists and is draft-complete enough to know done-ness
- [x] open-questions blocking entries resolved or Blocking: none
- [ ] No contradictions on ports, paths, install, OS, language
- [x] Verification subsections present on behavioral chapters
- [x] G9/D10/§10.1/A22-A24 internally consistent

## Notes
Ports (29850, loopback-only, no 1194), Go + `linux/amd64`/`windows/amd64` cross-compile (§4), N5/N9 (no macOS agent, no store/msi v1 requirement), and one-command install scripts as post-implementation targets (§5, §10.1, D8/D10) are aligned. OQ-001–003 are open with **Blocking: none**. Primary fail is **done-ness logic**, not wire/protocol shape.
