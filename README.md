# roblox-sync

System of tools that **bidirectionally** sync Luau between a local filesystem (any IDE: Cursor, VS Code, Neovim, …) and an open Roblox Studio place.

This git repository holds the **system specification** and three implementation **packages**:

| Package | Role |
|---------|------|
| [`sync-protocol/`](sync-protocol/) | Versioned wire contract (JSON schemas, mapping fixtures) |
| [`sync-agent/`](sync-agent/) | IDE-independent CLI/daemon for Windows and WSL |
| [`studio-plugin/`](studio-plugin/) | Roblox Studio plugin (Luau) |

Normative behavior: [`proj-spec/README.md`](proj-spec/README.md). Research allowlist: [`references.md`](references.md). License: [MIT](LICENSE) (covers this repository, including `sync-protocol`, `sync-agent`, and `studio-plugin`).

## Existing Roblox filesystem-sync tools — source ban

**YOU MUST NOT READ THEIR CODE.** Open-source tools already solve this problem. You may read **descriptions** of how they work and **API documentation** (especially official Roblox docs). You must **not** read any of their source code, even if stuck.

Full rule: [AGENTS.md](AGENTS.md). Allowed sources: [references.md](references.md).

## Why three packages (and not an IDE extension)

Roblox Studio plugins cannot see the host filesystem. A normal OS process cannot see Studio’s DataModel. Official HttpService plugin docs describe a **localhost companion program** as the supported way to bridge that gap. The companion is a CLI/daemon so Cursor, VS Code, and other editors work without a proprietary extension. An editor-extension package is explicitly out of scope for v1.

Plugin, rbx-sync, and protocol stay separate packages because Luau and Go cannot share a build, and the wire contract must not live inside either runtime. They share **this** git repository ([§3.1](proj-spec/chapters/03.1-repository-layout.md)).

## Status

Implementation is in progress through one-command installers (`scripts/install.ps1`, `scripts/install.sh`). Until this repository publishes GitHub Release assets, set `RBX_SYNC_RELEASE_BASE` to an HTTPS directory that contains `rbx-sync-windows-amd64.exe`, `rbx-sync-linux-amd64`, `rbx-sync.rbxmx`, and `SHA256SUMS`. Dry-run: `powershell -File scripts/install.ps1 -DryRun` or `sh scripts/install.sh --dry-run`.

## Install (user)

| Environment | Command |
|-------------|---------|
| Windows (PowerShell) | `powershell -ExecutionPolicy Bypass -File scripts/install.ps1` (or `irm <release-install.ps1-url> \| iex` once the remote exists) |
| Linux or WSL | `sh scripts/install.sh` (or `curl -fsSL <release-install.sh-url> \| sh` once the remote exists) |

Scripts refuse to replace files when a SHA-256 does not match `SHA256SUMS`. WSL uses the **linux** rbx-sync asset.

Release layout (from `sync-agent/`): `go run ./cmd/package-release -out dist` writes `rbx-sync-windows-amd64.exe`, `rbx-sync-linux-amd64`, `rbx-sync.rbxmx`, and `SHA256SUMS`.
