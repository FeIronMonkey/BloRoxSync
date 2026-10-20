# sync-agent

IDE-independent **Windows / WSL** CLI and daemon (`rbx-sync`) for roblox-sync. It owns the user’s project on disk and speaks `sync-protocol` to `studio-plugin`.

This **package** contains `cmd/rbx-sync`, `cmd/package-plugin` (`studio-plugin/src` → `rbx-sync.rbxmx`), and `internal/cli`. `rbx-sync version` prints `agent=0.1.0 protocol=1` (binary semver in `internal/cli.AgentVersion`; protocol from `sync-protocol`). Implementation follows the system spec (`../proj-spec/`).

Release and CI builds MUST use `CGO_ENABLED=0` (Go 1.23 or newer).

## Existing Roblox filesystem-sync tools — source ban

**YOU MUST NOT READ THEIR CODE.** You may read **descriptions** of how existing sync tools work and **API documentation**. You must **not** read any of their source code, even if stuck.

Full rule: [AGENTS.md](AGENTS.md).
