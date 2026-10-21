# Review: REVIEW-STAGE — roblox-sync Stage 13 re-RED
> **Verdict:** PASS
> **Reviewer model:** Composer 2.5 Fast
> **Implementer model:** Cursor Grok 4.6
> **Artifact reviewed:** commit 061c100, Stage 13 plan, prior PASS `reviews/roblox-sync-stage13.md`, installer tests/scripts, §10.1/§14 Verification deltas

## Blockers (must fix before progression)
- *(none)*

## Findings (non-blocking)
- **Normative LF not in §10.1 MUST list.** `#one-command-user-install` items 1–7 do not state that `scripts/install.sh` MUST be LF-only in the repo (or that git MUST enforce `eol=lf`). CHECK added Verification rows and §14 annotation only. The matrix claim for `#one-command-user-install` rests on implicit “POSIX `sh`” semantics, not an explicit MUST under *H*; a follow-up CHECK bullet in §10.1 (and optionally A23 table text) would close traceability without changing behavior.
- **GREEN (`.gitattributes`) is not test-guarded.** `TestInstallShHasUnixLineEndings_WhenCheckedIn` reads the working-tree script and rejects `\r`; it does not assert `.gitattributes` exists or that `scripts/*.sh text eol=lf` is present. CRLF could return if attributes were removed while the file stayed LF.
- **`install.sh` git mode 100755 → 100644.** Commit `061c100` only changes mode; content unchanged. Spec/README use `sh scripts/install.sh` / `curl … | sh`, not `./install.sh`, so this is not a Stage 13 blocker; direct execution without `sh` on Unix clones may need `chmod +x`.
- **§12-testing module blurb unchanged.** `sync-agent/internal/installscripts` row still says “dry-run + checksum” only; optional doc nit for this re-RED.
- **Carried from prior Stage 13 PASS (`a315a7a`), unchanged by `061c100`.** `TestInstallPs1HonorsReleaseBase_WhenEnvSet` does not assert staged-plugin output when plugins dir is unreachable (asymmetry vs `install.sh` honors test); Windows dry-run still keys on substring `"new"` for PATH messaging.

## Kick-back target
*(N/A — PASS)*

## Checklist
- [x] Stage RED vs tests — Re-RED intent: `TestInstallShHasUnixLineEndings_WhenCheckedIn` fails if `scripts/install.sh` contains `\r`; GREEN `.gitattributes` `scripts/*.sh text eol=lf`. Original Stage 13 installer behavior not modified in script body; prior RED coverage (dry-run, checksum, release base, plugin exit codes) remains in package tests.
- [x] Spec adherence — Verification CHECK aligns test name with A23 WSL/`sh` rationale; no `DEVIATION:` in commit body.
- [x] Anchors (D1 sites) — New test: two `Spec-Anchor` lines, one tag per line, both include `#heading-slug`. §10.1 and §14 Verification rows name `TestInstallShHasUnixLineEndings_WhenCheckedIn` with `Spec-Anchor` to `#one-command-user-install` / `#platform`.
- [x] Matrix claim check (stage-scoped) — Test *T* exists; would fail if CRLF bytes present (positive exercise of stated re-RED intent). Borderline vs strict “named under *H*” criterion: see Finding on normative LF prose.
- [x] No DEVIATION: without CHECK — Commit message has no `DEVIATION:` line.
- [x] L2 test run — `cd c:\repos\roblox-sync\sync-agent; & "C:\Program Files\Go\bin\go.exe" test ./internal/installscripts -count=1` → **ok** (~11.7s). Working tree `scripts/install.sh`: **0** `0x0D` bytes; `git ls-files -s scripts/install.sh` → mode **100644**.
