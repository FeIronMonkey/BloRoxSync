# Review: REVIEW-STAGE Stage 7 — roblox-sync

> **Verdict:** PASS  
> **Reviewer model:** composer-2.5-fast  
> **Implementer model:** Cursor Grok 4.6  
> **Artifact reviewed:** commit `be6b318`

Independent reviewer: [REVIEW-STAGE Stage 7](27e0ae35-4bb5-4334-a3f9-52dfa6d8f64f)

## Summary

Stage 7 replaces the apply stub with a full `POST /v1/apply` engine: in-memory per-`nodeId` revisions and conflicted-node tracking, two-phase validate-then-apply (so oversize and schema failures do not write disk), upsert/delete/rename with revision acks (rename ack uses `newNodeId`), partial-batch 409 `CONFLICT` with `applied` + `conflicts`, first-try `force` → 400 `BAD_REQUEST`, post-conflict `force` applies, Folder with present `source` → `BAD_REQUEST`, 413 `PAYLOAD_TOO_LARGE`, and 400 `NOT_MAPPED`. Snapshot returns live revisions; bootstrap studio-to-disk resets the revision table after writes.

Eleven behavioral tests in `apply_test.go` cover the plan RED list. Serve tests pass. CHECK updates name those tests across §6.1, §6.2, §6.5, §7.1, §8, §11, §13, and §14 (A9/A12). No `DEVIATION:` commits; no watch/SSE/plugin scope bleed.

## Blockers (must fix before progression)

None.

## Findings (non-blocking)

1. **F1 — §14 A9 Verification incomplete for “create”**  
   A9 requires create/update/delete/rename on disk. CHECK added upsert/delete/rename rows under `#roblox-studio--disk-g1` but not `TestServeApplyCreatesNode_WhenBaseRevisionIsZeroAndAbsent` for the create path. The test exists and is indexed in §6.5 and §7.1; only §14’s A9 mapping is incomplete (traceability gap, not missing behavior or tests).

2. **F2 — §6.2 `#apply-from-roblox-studio` index omits live create**  
   Serve apply rows list upsert/delete/rename but not `TestServeApplyCreatesNode_WhenBaseRevisionIsZeroAndAbsent`. Create-on-absent is primarily §6.5 `#live-apply`; optional index alignment only.

3. Carry-forward: echo/self-write and SSE remain Stage 8; plugin Keep disk / conflict widget remain Stage 10.

## Checklist

- [x] Plan Stage 7 RED intent covered by tests
- [x] GREEN implements normative §6.5 live apply / conflict force rules and §7.1 `POST /v1/apply`
- [x] CHECK updates stage-relevant chapter Verification
- [x] No `DEVIATION:` without spec-skill CHECK
- [x] No Stage 8/10 scope bleed
- [x] Serve tests pass

**VERDICT: PASS**
