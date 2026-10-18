# sync-agent

IDE-independent **Windows / WSL** CLI and daemon (`rbx-sync`) for roblox-sync. It owns the user’s project on disk and speaks `sync-protocol` to `studio-plugin`.

This **package** is initialized with scaffolding (`go.mod`, `cmd/rbx-sync`). Implementation follows the system spec (`../proj-spec/`).

Release and CI builds MUST use `CGO_ENABLED=0` (Go 1.23 or newer).

## Existing Roblox filesystem-sync tools — source ban

**YOU MUST NOT READ THEIR CODE.** You may read **descriptions** of how existing sync tools work and **API documentation**. You must **not** read any of their source code, even if stuck.

Full rule: [AGENTS.md](AGENTS.md).
