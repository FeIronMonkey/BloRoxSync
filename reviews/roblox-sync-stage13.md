# Review: REVIEW-STAGE — roblox-sync Stage 13
> **Verdict:** PASS
> **Reviewer model:** Composer 2.5 Fast
> **Implementer model:** Cursor Grok 4.6
> **Artifact reviewed:** commit a315a7a, Stage 13 plan, `scripts/install.ps1`, `scripts/install.sh`, `sync-agent/internal/installscripts`, touched chapters

## Blockers

*(none)*

## Findings

1. **Windows “staged plugin” message not asserted (RED coverage asymmetry).** Plan Stage 13 RED requires: when the Roblox plugins folder is unreachable, print the staged `.rbxmx` path and exit **0** after binary install. `TestInstallShHonorsReleaseBase_WhenEnvSet` checks `containsPath(out, plug)` for that message; `TestInstallPs1HonorsReleaseBase_WhenEnvSet` only verifies files on disk and exit 0, not `"staged plugin:"` in output. `install.ps1` implements the behavior; a regression omitting the print would still pass the Windows honors test.

2. **`TestInstallPs1DryRunPrintsWindowsDest_WhenDryRun` substring `"new"`.** The test requires output to contain `"new"`, satisfied by “A **new** terminal is required” rather than an explicit check for that phrase.

3. **G9 / D10 Verification index is dry-run–only.** §2 `#goals` (G9) and §9 D10 name only the two dry-run tests. Plugin-status and checksum coverage live under §10.1 and §14 `#platform`, consistent with A22/A23 L2 dry-run + manual tails.

4. **Post-install PATH messaging on Windows.** After a successful install, `install.ps1` may add the dest dir to the user PATH and still always prints the new-terminal note. Duplicate hint is UX-only.

5. **`requireSh` excludes WSL `bash.exe`.** On Windows, Git Bash is required. Linux CI and Windows+Git Bash are the intended L2 environments.

6. **§10.1 health-check `protocol=1` after install (out of Stage 13 ACs).** Stage 13 plan ACs are A22–A24 only. Stub binary supports `version` but no installer integration test yet — Stage 14 / L5, not a Stage 13 blocker.

## Checklist (RED vs tests, spec adherence, anchors, matrix claim check, no DEVIATION, L2 run)

| Item | Result |
|------|--------|
| **Stage RED vs tests** | **Pass** — Dry-run dest tables and no writes; checksum refuse + non-zero; `RBX_SYNC_RELEASE_BASE` honors tests; plugin-status exit 7; unreachable plugins exit 0 (staged path asserted on sh). WSL uses linux asset. |
| **Spec adherence** | **Pass** — Matches §10.1 destinations, checksum gate, dry-run, `RBX_SYNC_RELEASE_BASE`, plugin install when reachable. README documents env base and dry-run commands. |
| **Anchors (D1 sites)** | **Pass** — New tests use one `Spec-Anchor` per line. CHECK names tests under §10.1, §14, §5, §6.2, §2 G9, §9 D10. |
| **Matrix claim check (stage-scoped)** | **Pass** — Named `TestInstall*` rows exist and would fail if dest/asset/checksum/plugin-status behavior were inverted. Finding 1: ps1 honors test does not fail if staged-plugin print is removed. |
| **No DEVIATION: without CHECK** | **Pass** — Commit `a315a7a` has no `DEVIATION:` body. |
| **L2 run** | **Pass** — `go test ./internal/installscripts -count=1` → ok (~10s). |
