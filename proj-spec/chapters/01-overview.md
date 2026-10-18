# 1. Overview

> **Spec chapter:** What roblox-sync is and which problem it solves.
> **Parent:** [README.md](../README.md)
> **Index:** [README.md](../README.md)

roblox-sync is a **two-process** system:

1. **sync-agent** — package that builds the `rbx-sync` OS process on Windows or Linux (including WSL). That process owns the user’s **project** on disk (watch files, read/write UTF-8 Luau).
2. **studio-plugin** — a Roblox Studio plugin that owns the open **place**’s DataModel (create/move/delete `Folder` / `LuaSourceContainer`, read/write source).

They speak **sync-protocol** over HTTP on `127.0.0.1`. Neither side can do the other’s job: Roblox Studio plugins have no host filesystem API; a desktop process has no DataModel API. Official plugin HttpService documentation describes exactly this companion-process pattern ([references.md](../../references.md)).

A user edits Luau in any editor. rbx-sync notices file changes and the plugin applies them in Roblox Studio. The user edits a script (or adds/renames/deletes a synced script) in Roblox Studio, and rbx-sync writes the corresponding files. After a confirmed bootstrap, both directions stay live in one session.

This git repository specifies the system and holds the implementation packages described in [§3.1](03.1-repository-layout.md).

## Actors

These names are used throughout the spec. They are **not** an AI agent, Visual Studio, VS Code, or a Roblox place file.

| Actor | Meaning |
|-------|---------|
| User | Developer on a Windows PC, optionally using WSL for the project or for rbx-sync |
| rbx-sync | The `rbx-sync` OS process built from package `sync-agent`. A localhost companion for the plugin, **not** an AI agent. |
| Plugin | Roblox Studio plugin from package `studio-plugin` |
| Roblox Studio | Roblox Studio (the Roblox IDE) with a place open. **Not** Visual Studio or VS Code. |

Package directory names stay `sync-agent` and `studio-plugin`. Wire JSON keeps existing keys (`studio` instance paths, `origin` value `studio`, `agentVersion`, `disk-to-studio` / `studio-to-disk`). Those identifiers are not actor names.

## Trees

| Tree | Meaning |
|------|---------|
| Project | The user’s **local source tree** containing `rbx-sync.json` and mapped Luau files. Typically a git checkout edited in any IDE. rbx-sync’s `--project` path points here. |
| Place | The open Roblox experience in Roblox Studio (the DataModel). Saving it produces `.rbxl` / `.rbxlx`. v1 syncs mapped Luau **live** against this DataModel; it does **not** treat the place file as the sync tree ([§2](02-goals-and-non-goals.md) N3). |
| This repository | The `roblox-sync` git repo: spec plus implementation packages. Distinct from the user’s project. |

The project is the copy on the user’s machine for editing. It is not “the saved game” from Roblox Studio. After bootstrap, mapped Luau in the project and the corresponding instances in the place stay in sync; they remain two trees.

## End-to-end loop

1. User runs `rbx-sync serve` in the project.
2. rbx-sync binds `127.0.0.1` (default port **29850**), prints an **access token**.
3. User installs the plugin (once), opens the place, clicks **Connect**, pastes the token.
4. Plugin and rbx-sync handshake (`protocolVersion` must match).
5. If mapped trees differ, the plugin requires an explicit **Disk wins**, **Roblox Studio wins**, or **Cancel** choice ([§6.5](06.5-sync-session.md)).
6. Live bidirectional apply continues until disconnect.

## Verification

- Spec-Anchor: proj-spec/chapters/01-overview.md#actors
- Spec-Anchor: proj-spec/chapters/01-overview.md#trees
- Spec-Anchor: proj-spec/chapters/01-overview.md#end-to-end-loop

## Related chapters

- [2. Goals and non-goals](02-goals-and-non-goals.md)
- [3. Architecture](03-architecture.md)
- [3.1 Repository layout](03.1-repository-layout.md)
