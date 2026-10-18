# Agent instructions

This directory is the **sync-agent** package (`rbx-sync`), an IDE-independent CLI/daemon that watches the local filesystem and serves `sync-protocol` to the Roblox Studio plugin. System spec: `../proj-spec/`.

## Existing Roblox filesystem-sync tools — source ban

**YOU MUST NOT READ THEIR CODE.** Open-source tools already sync a local tree with Roblox Studio. It is acceptable to read **descriptions** of how they work and to review **API docs** (including official Roblox Creator Hub documentation). It is **not** acceptable, under any circumstances, to read any of their code — not even if you cannot figure out how to do something.

This ban includes source trees, raw file views, diffs, generated sites that paste implementation source, vendored copies, and copying ports, HTTP paths, or snippets from those projects.

Use **`../references.md`** as the research allowlist and keep that file updated when you consult a new allowed source. If a task seems to require reading those tools’ code, **stop** and design from official Roblox APIs and the system spec.
