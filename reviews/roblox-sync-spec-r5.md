# Review: REVIEW-SPEC — roblox-sync
> **Verdict:** PASS
> **Reviewer model:** composer-2.5-fast
> **Implementer model:** Cursor Grok 4.6
> **Artifact reviewed:** proj-spec/README.md, proj-spec/open-questions.md, proj-spec/chapters/* (r5 after r4 FAIL)

Independent reviewer: [REVIEW-SPEC r5](3c353952-d125-4821-81bb-954ec4d26903)

## Blockers

*(none)*

The r4 blocker is **resolved**. §2 still states WSL round-trip as part of v1 success, but §14 now makes that obligation explicit and non-substitutable:

- **A17** requires topology **T2** with WSL mirrored networking (or the equivalent documented in references.md): successful Studio `GET /v1/health`, successful handshake, and a **bidirectional mapped-script round-trip**. A disconnected plugin does **not** satisfy A17.
- **A17b** is separate: health-failure UX only (“agent not reachable”, no sync). It does **not** satisfy WSL done-ness.
- §14 done clause requires **A1–A24 (including A11b and A17b)** — there is no path to “§14 complete” without A17’s WSL round-trip.

## Findings

1. **§6.3 connect flow — unequal bootstrap path:** Steps 6–8 jump from Disk/Studio/Cancel to the equal snapshot path without numbered bootstrap HTTP on the unequal branch (recoverable from §6.5 / §11).
2. **Topology T3 vs A17:** T3 has no dedicated AC row; T2 is the required WSL acceptance topology.
3. **`install.sh` exit code when plugin leg fails:** Exit 0 when plugins folder is unreachable; unspecified when agent install succeeds but `plugin install` exits 1.
4. **§13 snapshot/bootstrap latency:** 30 s MUST has no A-id.
5. **§7.3 `--file` default on Linux/WSL:** Staged `.rbxmx` is not next to the binary; fallback (2) is required.

## Checklist

- [x] All chapters in README exist
- [x] §14 done-ness consistent with §2 G4/G9
- [x] open-questions blocking none or resolved
- [x] Install paths consistent §5/§6.2/§7.3/§10.1/A22-A24
- [x] A17 requires WSL round-trip; A17b does not substitute
