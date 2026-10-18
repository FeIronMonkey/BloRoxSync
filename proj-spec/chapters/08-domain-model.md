# 8. Domain model

> **Spec chapter:** Shared JSON types.
> **Parent:** [README.md](../README.md)
> **Index:** [README.md](../README.md)

Types are normative for `sync-protocol` schemas.

## Node

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `nodeId` | string | yes | [§6.4](06.4-project-mapping.md#nodeid) |
| `className` | string | yes | `Folder` \| `Script` \| `LocalScript` \| `ModuleScript` |
| `source` | string | if not Folder | UTF-8, LF newlines |
| `revision` | number | on snapshots after handshake | unsigned integer `< 2^53` |

## Change

See [§7.1 POST /v1/apply](07.1-http-api.md#post-v1apply). Additional field `force` boolean, default false.

## Event

See [§7.2](07.2-event-stream.md#event-object).

## Root (handshake)

| Field | Type | Notes |
|-------|------|-------|
| `id` | string | root id |
| `studio` | string | Roblox Studio instance path from the project file (`roots[].studio`) |

Disk paths are **not** sent to the plugin.

## Project file

See [§6.4](06.4-project-mapping.md#project-file).

## Verification

- Spec-Anchor: proj-spec/chapters/08-domain-model.md#node
- Spec-Anchor: proj-spec/chapters/08-domain-model.md#change
- Spec-Anchor: proj-spec/chapters/08-domain-model.md#event
- Spec-Anchor: proj-spec/chapters/08-domain-model.md#root-handshake
- Spec-Anchor: proj-spec/chapters/08-domain-model.md#project-file

## Related chapters

- [6.1 sync-protocol](06.1-sync-protocol.md)
- [6.4 Project mapping](06.4-project-mapping.md)
