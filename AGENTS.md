# Agent instructions

This repository is the **system specification and implementation workspace** for roblox-sync (plugin + rbx-sync process + protocol). Implementation packages in this repo: `studio-plugin`, `sync-agent`, `sync-protocol`.

## Existing Roblox filesystem-sync tools — source ban

**YOU MUST NOT READ THEIR CODE.** Open-source tools already sync a local tree with Roblox Studio. It is acceptable to read **descriptions** of how they work and to review **API docs** (including official Roblox Creator Hub documentation). It is **not** acceptable, under any circumstances, to read any of their code — not even if you cannot figure out how to do something.

This ban includes, without limitation:

- GitHub, GitLab, or other host **source trees**, raw file views, blobs, and diffs
- Plugin Luau, CLI/daemon implementation, protocol codecs, and tests from those projects
- Generated sites that paste implementation source (for example code-entity wikis)
- Vendored, cloned, or submodule copies of those codebases
- Copying ports, HTTP paths, file-format names, or snippets from those repositories

**You MAY read:**

- Official Roblox Creator Hub documentation and engine class pages
- High-level product descriptions, README overviews, and API documentation that do not include implementation source
- Entries listed in [references.md](references.md)

If a task seems to require reading those tools’ code, **stop** and design from official Roblox APIs, [references.md](references.md), and `proj-spec/`.

Allowed research sources and why to use them: **[references.md](references.md)** (keep that file updated when you consult a new allowed source).

## Spec-driven process

This project uses spec-driven development. Do not implement behavior until `proj-spec/` chapters, REVIEW-SPEC, user spec-gate approval, implementation plan, REVIEW-PLAN, and user plan-gate approval are complete. See `proj-spec/README.md`.
