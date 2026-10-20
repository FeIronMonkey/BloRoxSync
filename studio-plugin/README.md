# studio-plugin

Roblox Studio plugin for **roblox-sync**. It connects to local rbx-sync over localhost HTTP and applies bidirectional Luau sync against the open place.

Implementation follows the system spec (`../proj-spec/`). Mapping tables live in `src/mapping.luau` (pure Luau, no HttpService). Connect/session logic lives in `src/session.luau` (HTTP and DataModel injected for L4). The Studio host (`src/host.luau`) creates the `rbx-sync` toolbar, dockable widget, and HttpService adapter; the packaged `.rbxmx` Script calls `Host.start` with the `plugin` global. L4 unit tests run under [Lune](https://lune-org.github.io/docs) from the repository root:

```
lune run studio-plugin/tests
```

Those tests load the same `sync-protocol/fixtures/v1/mapping/` pack as rbx-sync. They do not start Roblox Studio.

Package a local-plugin `.rbxmx` from `src/` (run from `sync-agent/`):

```
go run ./cmd/package-plugin -src ../studio-plugin/src -out rbx-sync.rbxmx
```

Then `rbx-sync plugin install --file rbx-sync.rbxmx`.

## Existing Roblox filesystem-sync tools — source ban

**YOU MUST NOT READ THEIR CODE.** You may read **descriptions** of how existing sync tools work and **API documentation**. You must **not** read any of their source code, even if stuck.

Full rule: [AGENTS.md](AGENTS.md).
