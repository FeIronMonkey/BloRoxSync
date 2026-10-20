# References

This file is the **research allowlist** for roblox-sync. When specifying or implementing behavior, prefer these sources. If you add a source, record it here with *why* it is useful so later sessions do not re-discover (or accidentally open) the wrong material.

See also [AGENTS.md](AGENTS.md) for the hard ban on reading existing sync-tool **source code**.

## How to use this file

| Situation | Consult |
|-----------|---------|
| Plugin talking to a process on the same PC | [HttpService — plugins / localhost](#httpservice--in-experience-http-requests) |
| Live stream of disk events into Roblox Studio | [HttpService class — CreateWebStreamClient](#httpservice-class) |
| Reading or writing Luau while a script tab is open | [ScriptEditorService](#scripteditorservice) |
| Toolbar, connect UI, local plugin object | [Plugin class](#plugin-class) |
| Script vs LocalScript vs ModuleScript; `Source` | [LuaSourceContainer](#luasourcecontainer) and [Script](#script-class) |
| Where to install a local `.rbxmx` / `.lua` plugin on Windows | [Local plugins folder](#local-plugins-folder) |
| rbx-sync in WSL, Roblox Studio on Windows, `localhost` | [WSL networking](#wsl-networking) |
| Why a companion process is required at all | [Roblox external tools](#roblox-external-tools--the-syncing-problem) |

## Forbidden (do not add as implementation references)

Do **not** open, clone, vendor, or cite as an implementation guide the **source trees** of existing Roblox filesystem-sync tools, including Rojo, Argon, and forks. That includes GitHub/GitLab file views, raw blobs, diffs, generated site pages that paste source (for example DeepWiki code entity pages), and vendored copies.

High-level **product** pages that describe user-facing workflow without pasting implementation source may be listed under [Existing tools — descriptions only](#existing-tools--descriptions-only). Those entries are for product-shape awareness, not for copying protocols, ports, file names, or code.

---

## Official Roblox documentation

### HttpService — in-experience HTTP requests

- **URL:** https://create.roblox.com/docs/cloud-services/http-service
- **Use for:** The **plugin** subsection: Studio plugins may call `HttpService` without the experience “Allow HTTP Requests” setting; the first request to a domain may prompt the user; plugins may talk to `localhost` and `127.0.0.1` so a separately distributed local program can reach the host filesystem. Security warning: that companion is not part of the plugin package.
- **Do not over-apply:** The page’s **Limitations** list (HTTPS-only, restricted headers, Open Cloud secrets) is written for in-experience / Open Cloud calls. Plugin-to-localhost HTTP is documented in the plugin subsection and is the basis for this system’s transport. Still design the protocol to work with `Content-Type` plus a query-string token so it remains robust if header policy is tight.

### HttpService class

- **URL:** https://create.roblox.com/docs/reference/engine/classes/HttpService
- **Use for:** `RequestAsync` / `GetAsync` / `PostAsync` / `JSONEncode` / `JSONDecode`; `CreateWebStreamClient` for Studio-side streaming (SSE, chunked transfer, WebSockets) including the documented localhost streaming pattern.

### Plugin HTTP permissions (announcement)

- **URL:** https://devforum.roblox.com/t/introducing-plugin-http-permissions/493269
- **Use for:** Per-domain permission prompts; preferences persist in Plugin Management; **local plugins bypass HTTP permission prompts**; marketplace plugins ignore the experience HTTP setting.

### HttpService port restrictions (announcement)

- **URL:** https://devforum.roblox.com/t/port-restrictions-for-httpservice/1500073
- **Use for:** Game-server `HttpService` blocks ports below 1024 except 80/443, and blocks 1194. The announcement states Studio plugins on the local computer are not the same as game-server requests. This project still uses an unprivileged default port (**29850**) and never binds reserved ports.

### ScriptEditorService

- **URL:** https://create.roblox.com/docs/reference/engine/classes/ScriptEditorService
- **Use for:** Plugin-security API to find, open, and edit `ScriptDocument`s. If a script is open in the editor, prefer this service over writing `Source` so the open tab stays consistent. `FindScriptDocument` returns nil when the tab is closed.

### Plugin class

- **URL:** https://create.roblox.com/docs/reference/engine/classes/Plugin
- **Use for:** `CreateToolbar` / toolbar buttons, plugin UI, `OpenScript` (prefer `ScriptEditorService:OpenScriptDocumentAsync` when integrating with the editor).

### LuaSourceContainer

- **URL:** https://create.roblox.com/docs/reference/engine/classes/LuaSourceContainer
- **Use for:** Abstract base of `Script`, `LocalScript`, and `ModuleScript` — the only instance classes this system syncs as source files in v1.

### Script class

- **URL:** https://create.roblox.com/docs/reference/engine/classes/Script
- **Use for:** `Source` is a protected string with **PluginOrOpenCloud** capability. Official guidance: when a user has the script open, use `ScriptEditorService` (`UpdateSourceAsync` / `GetEditorSource`) rather than assigning `Source` directly.

### Roblox external tools — the syncing problem

- **URL:** https://create.roblox.com/docs/projects/external-tools
- **Use for:** Roblox’s own framing: instances as files, getting files back into Studio, conflicts, and the need to listen for file changes. Confirms the **local server + Studio plugin** companion pattern at a product level. Treat any named third-party tool on that page as **description-only** — do not follow links into that tool’s source.

### Local plugins folder

- **Windows path:** `%LOCALAPPDATA%\Roblox\Plugins` (typically `C:\Users\<user>\AppData\Local\Roblox\Plugins`)
- **Use for:** Installing the Studio plugin as a local `.lua` / `.rbxm` / `.rbxmx` file. Studio’s Plugins → Open Plugins Folder is the user-facing way to confirm the path. Fully restart Studio after replacing a local plugin.
- **Forum confirmations (path only, not implementation):** https://devforum.roblox.com/t/how-to-import-a-plugin-from-your-computer-to-roblox-studio/3234934 and https://devforum.roblox.com/t/where-is-the-local-plugins-folder-located/1369813

---

## Platform

### WSL networking

- **URL:** https://learn.microsoft.com/en-us/windows/wsl/networking
- **Use for:** WSL2 default **NAT** vs **mirrored** mode (`networkingMode=mirrored` in `.wslconfig`, Windows 11 22H2+). Mirrored mode: Windows and WSL can use `127.0.0.1` to reach each other (IPv6 `::1` not supported for that path). NAT mode: Windows can often reach a WSL listening server via `localhost`; WSL reaching a Windows-only listener may need the Windows host IP from `ip route`. Do not bind rbx-sync to `0.0.0.0` just to paper over NAT — prefer mirrored mode or running rbx-sync on Windows next to Roblox Studio.

---

## Existing tools — descriptions only

These pages describe *that* filesystem↔Studio sync products exist and how users operate them. They are **not** a protocol or architecture specification for this repo.

| Tool | Allowed page types | Use for |
|------|--------------------|---------|
| Rojo | Roblox “external tools” overview; project README **prose** if it does not include source files | Existence of the plugin + local server workflow; reminder that two-way sync is a hard product problem |
| Argon | https://argon.wiki/ (home, installation, common usage) | Existence of CLI + Studio plugin (+ optional editor extension); confirmation that an IDE extension is **optional** if a CLI exists |

**Do not** open their GitHub source paths, plugin Luau trees, HTTP route tables copied from source, or generated code wikis.

---

## Toolchain (tests)

### Lune

- **URL:** https://lune-org.github.io/docs/
- **Use for:** Running L4 `studio-plugin/tests` as standalone Luau (modules, `fs`, `serde` JSON, `process`) without Roblox Studio. Plan pin: [implementation-plan.md](proj-spec/implementation-plan.md) Stage 9 / §12 L4.
- **Do not over-apply:** `@lune/*` libraries are test-harness only. Production plugin scripts under `studio-plugin/src` stay engine/plugin APIs plus pure Luau.

### Lune command-line / modules / stdio APIs

- **CLI:** https://lune-org.github.io/docs/getting-started/2-command-line-usage/
- **Modules (`require`, `@self`, `init.luau`):** https://lune-org.github.io/docs/the-book/7-modules/
- **fs:** https://lune-org.github.io/docs/api-reference/fs/
- **serde:** https://lune-org.github.io/docs/api-reference/serde/
- **process:** https://lune-org.github.io/docs/api-reference/process/
- **Install:** https://lune-org.github.io/docs/getting-started/1-installation/ (CI uses GitHub Release binaries; local install may use Rokit or the same zips)

## Recording new references

When you consult a new document:

1. Confirm it is not forbidden source.
2. Add a subsection: URL, **Use for**, and any **Do not over-apply** caveat.
3. If it changes normative spec text, update the relevant `proj-spec/chapters/` file in the same change.
