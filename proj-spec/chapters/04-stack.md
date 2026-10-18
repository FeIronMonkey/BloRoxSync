# 4. Stack

> **Spec chapter:** Languages, toolchains, and third-party constraints.
> **Parent:** [README.md](../README.md)
> **Index:** [README.md](../README.md)

## Chosen stack

| Component | Stack | Why |
|-----------|-------|-----|
| sync-protocol | JSON Schema 2020-12, UTF-8 JSON fixtures | Language-neutral contract; Go and Luau can both validate against examples |
| sync-agent | Go 1.23 or newer, standard library `net/http`, `cgo` disabled | One source tree; **separate static binaries** per OS/arch (`windows/amd64`, `linux/amd64`) with no runtime install |
| studio-plugin | Luau, Roblox Plugin security APIs | Only language Roblox Studio plugins run |

## Cross-compilation (rbx-sync)

v1 MUST ship **one Go module** and **two release binaries** (same `protocolVersion`):

| `GOOS` / `GOARCH` | Artifact name | Runs on |
|-------------------|---------------|---------|
| `windows` / `amd64` | `rbx-sync.exe` | Native Windows |
| `linux` / `amd64` | `rbx-sync` | WSL and native Linux |

Building for Windows does **not** require a Windows compiler when `CGO_ENABLED=0`. Building for Linux does **not** require a Linux host. CI MAY produce both artifacts from a single runner. Users who use [§10.1](10.1-install.md) one-command install never run `go build`.

v1 MUST NOT require `cgo`, MSVC, gcc, or a system libc beyond what a typical `linux/amd64` musl-or-glibc static/mostly-static Go binary already needs. `CGO_ENABLED=0` is required for release builds.

macOS (`darwin`) rbx-sync binaries are out of scope for v1 ([§2](02-goals-and-non-goals.md) N5).

## rbx-sync dependencies

rbx-sync MUST implement HTTP serving and filesystem watch using:

- Go standard library for HTTP JSON (`net/http`, `encoding/json`)
- A maintained filesystem watcher **library** is allowed if it is not taken from a Roblox sync tool

rbx-sync MUST NOT vendor or generate code from Rojo, Argon, or similar projects.

## Plugin dependencies

The plugin MUST use only Roblox engine APIs listed in [references.md](../../references.md) (HttpService, Plugin, ScriptEditorService, instance tree). No extra Roblox Studio packages are required for v1.

## JSON conventions

| Rule | Value |
|------|-------|
| Encoding | UTF-8 |
| Numbers | JSON numbers; revisions are unsigned integers in range `[0, 2^53)` so Luau and JSON agree |
| Keys | camelCase |
| Time | not required on the wire in v1 |
| Line endings on disk | LF (`\n`) |

## Verification

- `TestLayoutHasGoModule_WhenSyncAgentExists`
  Spec-Anchor: proj-spec/chapters/04-stack.md#chosen-stack
- `TestLayoutDocumentsCgoDisabled_WhenSyncAgentReadmeExists`
  Spec-Anchor: proj-spec/chapters/04-stack.md#cross-compilation-rbx-sync
- `TestLayoutBuildsRbxSyncStub_WhenGoModuleExists`
  Spec-Anchor: proj-spec/chapters/04-stack.md#cross-compilation-rbx-sync
- Spec-Anchor: proj-spec/chapters/04-stack.md#json-conventions

## Related chapters

- [5. File tree](05-file-tree.md)
- [9. Design decisions](09-design-decisions.md)
- [10.1 Install](10.1-install.md)
