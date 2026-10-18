# 11. Examples

> **Spec chapter:** Concrete walkthroughs of mapping and session.
> **Parent:** [README.md](../README.md)
> **Index:** [README.md](../README.md)

## Example project file

```json
{
  "protocolVersion": 1,
  "name": "demo",
  "roots": [
    { "id": "server", "disk": "src/server", "studio": "ServerScriptService" },
    { "id": "shared", "disk": "src/shared", "studio": "ReplicatedStorage.Shared" }
  ]
}
```

## Example mapping

Disk:

```
src/server/Main.server.luau
src/server/Lib/init.luau
src/server/Lib/Util.luau
src/shared/Config.luau
```

Roblox Studio:

- `ServerScriptService.Main` — Script
- `ServerScriptService.Lib` — ModuleScript (`init.luau`)
- `ServerScriptService.Lib.Util` — ModuleScript
- `ReplicatedStorage.Shared` — Folder (created if missing)
- `ReplicatedStorage.Shared.Config` — ModuleScript

nodeIds: `server/Main`, `server/Lib`, `server/Lib/Util`, `shared/Config`.

## Example first connect (trees differ)

1. rbx-sync stderr contains `accessToken=abc…` and `listen=127.0.0.1:29850`.
2. Plugin Connect → `POST /v1/session` 200.
3. `POST /v1/bootstrap/compare` → `{ "equal": false }`.
4. User clicks **Disk wins** → `POST /v1/bootstrap/apply` `{ "direction": "disk-to-studio" }` → plugin creates the instances above.
5. User edits `Util.luau` on disk → SSE `upsert` `server/Lib/Util` → plugin updates source.
6. User edits `Main` in Roblox Studio → `POST /v1/apply` upsert `server/Main` → rbx-sync overwrites `Main.server.luau`.

## Example conflict

Both sides edit `shared/Config` before either apply lands. rbx-sync returns 409 `CONFLICT` with `currentRevision`. Plugin shows Keep disk / Keep Roblox Studio. Keep Roblox Studio resends with `"force": true`.

## Verification

- Spec-Anchor: proj-spec/chapters/11-examples.md#example-mapping
- Spec-Anchor: proj-spec/chapters/11-examples.md#example-first-connect-trees-differ

## Related chapters

- [6.4 Project mapping](06.4-project-mapping.md)
- [6.5 Sync session](06.5-sync-session.md)
