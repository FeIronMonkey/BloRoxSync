# Agent instructions

This directory is the **sync-protocol** package: `protocolVersion`, JSON Schema, and mapping/HTTP fixtures shared by `sync-agent` and `studio-plugin`. It is not a runtime process. System spec: `../proj-spec/`.

## Existing Roblox filesystem-sync tools — source ban

**YOU MUST NOT READ THEIR CODE.** Open-source tools already sync a local tree with Roblox Studio. It is acceptable to read **descriptions** of how they work and to review **API docs** (including official Roblox Creator Hub documentation). It is **not** acceptable, under any circumstances, to read any of their code — not even if you cannot figure out how to do something.

This ban includes source trees, raw file views, diffs, generated sites that paste implementation source, vendored copies, and copying ports, HTTP paths, mapping filenames, or snippets from those projects. Mapping rules for *this* protocol are specified in `proj-spec` and encoded as fixtures here — do not import another tool’s project format.

Use **`../references.md`** as the research allowlist and keep that file updated when you consult a new allowed source. If a task seems to require reading those tools’ code, **stop** and design from official Roblox APIs and the system spec.
