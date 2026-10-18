# 3. Architecture

> **Spec chapter:** Runtime processes, trust boundary, and data flow.
> **Parent:** [README.md](../README.md)
> **Index:** [README.md](../README.md)

## Processes

```
┌─────────────────────────────┐         HTTP JSON on 127.0.0.1:29850
│  Project on disk            │         accessToken query param
│  rbx-sync.json + src/**     │
└────────────┬────────────────┘
             │ read/write/watch
             ▼
┌─────────────────────────────┐         ┌──────────────────────────┐
│  sync-agent (rbx-sync)      │◄───────►│  studio-plugin (Luau)    │
│  Windows, WSL, or Linux     │  /v1/*  │  inside Roblox Studio    │
└─────────────────────────────┘         └────────────┬─────────────┘
                                                     │ DataModel
                                                     ▼
                                        Script / LocalScript / ModuleScript
                                        + Folder under mapped roots
```

| Process | Package | Language | Privileges |
|---------|---------|----------|------------|
| rbx-sync | `sync-agent` | Go | User filesystem; binds loopback only (`127.0.0.1` default; `::1` allowed) |
| Plugin | `studio-plugin` | Luau | Plugin security: HttpService, ScriptEditorService, instance tree |
| (none) | `sync-protocol` | JSON Schema + fixtures | Not a runtime process |

This git repository holds the spec and the three packages. Runtime processes are only rbx-sync and the plugin. Actor names: [§1](01-overview.md#actors).

## Trust boundary

rbx-sync MUST treat the plugin as an untrusted HTTP client that happens to be local:

- Bind loopback only: **`127.0.0.1`** (default) or **`::1`**. Never `0.0.0.0` or any non-loopback address. The plugin default host is IPv4 `127.0.0.1`; if rbx-sync is bound to `::1` only, the user MUST set the plugin host to `::1`.
- Require `accessToken` on every `/v1` route except `GET /v1/health`.
- Health MUST NOT return the token, project path, or mapped tree.
- Token is generated per `serve` process (not reused across process restarts).

Roblox Studio’s plugin HTTP permission prompt may appear for marketplace installs; **local plugins bypass** that prompt ([references.md](../../references.md)). The token is still required.

## Direction of changes

| Origin | Detected by | Applied by |
|--------|-------------|------------|
| Disk | rbx-sync file watch | Plugin mutates DataModel / ScriptEditorService |
| Roblox Studio | Plugin instance and source listeners | rbx-sync writes files |

Echo suppression: each applied batch carries `origin` (`disk` \| `studio`). The receiving side MUST NOT bounce the same revision back ([§6.5](06.5-sync-session.md)).

## Transport

| Channel | Role |
|---------|------|
| `GET/POST /v1/*` JSON | Handshake, snapshot, apply, bootstrap, poll |
| `GET /v1/events` SSE (preferred) | Disk-originated live events |
| `GET /v1/changes?since=` | Poll fallback if the stream client fails |

The plugin SHOULD open `HttpService:CreateWebStreamClient` against `/v1/events` ([references.md](../../references.md)). If that fails, it MUST poll `/v1/changes` at 200 ms.

Default listen address: `127.0.0.1:29850`. Port MUST be ≥ 1024 and MUST NOT be 1194.

## Why three packages and not an IDE extension

- **Plugin and rbx-sync ship on different runtimes** (Luau inside Roblox Studio vs a Go binary). They remain separate **packages** with their own builds; they share this git repository ([§3.1](03.1-repository-layout.md)).
- **Protocol as a third package** keeps mapping fixtures and schemas free of either runtime, so neither implementation is source-of-truth for the contract.
- **No IDE extension in v1:** G3 requires rbx-sync to be useful from a terminal. An extension would duplicate `serve` and couple the system to one editor.

## Verification

- Spec-Anchor: proj-spec/chapters/03-architecture.md#processes
- Spec-Anchor: proj-spec/chapters/03-architecture.md#trust-boundary
- Spec-Anchor: proj-spec/chapters/03-architecture.md#transport

## Related chapters

- [3.1 Repository layout](03.1-repository-layout.md)
- [6.5 Sync session](06.5-sync-session.md)
- [7.1 HTTP API](07.1-http-api.md)
