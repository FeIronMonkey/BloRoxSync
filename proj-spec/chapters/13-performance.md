# 13. Performance

> **Spec chapter:** Bound latencies and payload sizes for v1.
> **Parent:** [README.md](../README.md)
> **Index:** [README.md](../README.md)

## Limits

These are **requirements**, not micro-optimizations.

| Limit | Value |
|-------|-------|
| Max source per node | 1 MiB |
| Disk watch debounce | 50 ms |
| Roblox Studio mutation debounce | 50 ms |
| Poll interval (fallback) | 200 ms |
| SSE heartbeat | ≤ 15 s |
| Snapshot / bootstrap HTTP | Complete within 30 s for ≤ 2000 mapped nodes totaling ≤ 20 MiB |
| Bind | Single listener; one live plugin session |

rbx-sync MUST NOT spin a busy loop on watch. Apply batches SHOULD include all coalesced events for a debounce window.

Exceeding max source: `PAYLOAD_TOO_LARGE`, no partial write of that node.

## Verification

- `TestMapfsUpsertRejectsOversize_WhenSourceExceedsOneMiB`
  Spec-Anchor: proj-spec/chapters/13-performance.md#limits
- `TestServeApplyPayloadTooLarge_WhenSourceExceedsOneMiB`
  Spec-Anchor: proj-spec/chapters/13-performance.md#limits

Expected tests: 1 MiB + 1 byte apply rejected (mapfs + HTTP); debounce coalesces two writes into one event (rbx-sync).

## Related chapters

- [6.5 Sync session](06.5-sync-session.md)
