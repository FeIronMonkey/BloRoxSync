# roblox-sync

System of tools that **bidirectionally** sync Luau between a local filesystem (any IDE: Cursor, VS Code, Neovim, …) and an open Roblox Studio place.

This git repository holds the **system specification** and three implementation **packages**:

| Package | Role |
|---------|------|
| [`sync-protocol/`](sync-protocol/) | Versioned wire contract (JSON schemas, mapping fixtures) |
| [`sync-agent/`](sync-agent/) | IDE-independent CLI/daemon for Windows and WSL |
| [`studio-plugin/`](studio-plugin/) | Roblox Studio plugin (Luau) |

Normative behavior: [`proj-spec/README.md`](proj-spec/README.md). Research allowlist: [`references.md`](references.md).

## Existing Roblox filesystem-sync tools — source ban

**YOU MUST NOT READ THEIR CODE.** Open-source tools already solve this problem. You may read **descriptions** of how they work and **API documentation** (especially official Roblox docs). You must **not** read any of their source code, even if stuck.

Full rule: [AGENTS.md](AGENTS.md). Allowed sources: [references.md](references.md).

## Why three packages (and not an IDE extension)

Roblox Studio plugins cannot see the host filesystem. A normal OS process cannot see Studio’s DataModel. Official HttpService plugin docs describe a **localhost companion program** as the supported way to bridge that gap. The companion is a CLI/daemon so Cursor, VS Code, and other editors work without a proprietary extension. An editor-extension package is explicitly out of scope for v1.

Plugin, rbx-sync, and protocol stay separate packages because Luau and Go cannot share a build, and the wire contract must not live inside either runtime. They share **this** git repository ([§3.1](proj-spec/chapters/03.1-repository-layout.md)).

## Status

Stage 0 scaffolding is in place (package trees, `sync-agent` Go module, CI hooks for L1–L4). Packages are not a runnable product yet.
