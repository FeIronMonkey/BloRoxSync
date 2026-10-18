# 2. Goals and non-goals

> **Spec chapter:** In-scope outcomes and explicit exclusions for v1.
> **Parent:** [README.md](../README.md)
> **Index:** [README.md](../README.md)

## Goals

The v1 system MUST:

| ID | Goal |
|----|------|
| G1 | Sync Luau **from Roblox Studio to disk** for every mapped `LuaSourceContainer` (create, update, rename, delete). |
| G2 | Sync Luau **from disk to Roblox Studio** for every mapped source file (create, update, rename, delete). |
| G3 | Work with **any IDE** or none — rbx-sync is a CLI/daemon, not an editor extension. |
| G4 | Run rbx-sync on **Windows natively** and on **Linux amd64** (including **WSL**), with Roblox Studio remaining a Windows process. |
| G5 | Restrict the HTTP listener to **localhost** and require a per-serve **access token**. |
| G6 | Use a versioned **sync-protocol** so plugin and rbx-sync can disagree loudly (`PROTOCOL_MISMATCH`) instead of silently corrupting trees. |
| G7 | Keep mapping rules **normative in the protocol repo** so both implementations can be tested against the same fixtures. |
| G8 | Document and enforce the ban on reading existing sync-tool **source** (`AGENTS.md`, each repo `README.md`, [references.md](../../references.md)). |
| G9 | Provide a **one-command** install of rbx-sync for Windows and for Linux/WSL; install the plugin in the same command when Roblox Studio’s local plugins folder is reachable. |

## Non-goals

v1 MUST NOT claim to:

| ID | Non-goal |
|----|----------|
| N1 | A VS Code, Cursor, or other **editor extension** (optional future repo; [OQ-001](../open-questions.md)). |
| N2 | Sync of **non-script** instances (Parts, GUIs, attributes, tags, models) except `Folder` containers needed to hold scripts ([OQ-002](../open-questions.md)). |
| N3 | Building or importing **`.rbxl` / `.rbxlx` / `.rbxm`** as the primary sync mechanism ([OQ-003](../open-questions.md)). |
| N4 | Roblox **Open Cloud** publish/download. |
| N5 | macOS Roblox Studio, native Linux Roblox Studio, or a **macOS rbx-sync** binary. Roblox Studio is Windows in this project’s target. The Linux rbx-sync binary is for **WSL and native Linux** processes, not for Roblox Studio on Linux. |
| N6 | Multi-user simultaneous Roblox Studio sessions against one rbx-sync process (one plugin session; a new handshake replaces the previous). |
| N7 | Compatibility with other tools’ project files, ports, or wire formats. |
| N8 | Syncing CoreGui, plugin UIs, or the rbx-sync plugin’s own instances. |
| N9 | Package-manager distribution (winget, apt, Homebrew, Microsoft Store) or a signed Windows installer (`.msi` / `.msix`) as a v1 requirement. GitHub Releases (or equivalent) plus the one-command scripts in [§10.1](10.1-install.md) are sufficient. |

## Success definition

v1 is done when [§14](14-acceptance-criteria.md) is met: a Windows user can round-trip mapped scripts in both directions with rbx-sync on Windows **and** with rbx-sync on WSL (mirrored networking or documented equivalent), and G9 one-command install works on Windows and Linux/WSL.

## Verification

- Spec-Anchor: proj-spec/chapters/02-goals-and-non-goals.md#goals
- Spec-Anchor: proj-spec/chapters/02-goals-and-non-goals.md#non-goals

## Related chapters

- [14. Acceptance criteria](14-acceptance-criteria.md)
- [9. Design decisions](09-design-decisions.md)
- [10.1 Install](10.1-install.md)
