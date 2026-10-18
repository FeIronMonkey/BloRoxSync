# Open questions

Gate rule: REVIEW-SPEC may PASS only when every entry is `resolved`, or `open` with `Blocking: none`.

### OQ-001: Optional editor extension
- **Tags:** `proj-spec/chapters/02-goals-and-non-goals.md#non-goals`, `proj-spec/chapters/03.1-repository-layout.md#packages`
- **Discovered in:** REVIEW-SPEC prep
- **Blocking:** none
- **Status:** open
- **Question:** Should a later package add a VS Code / Cursor extension that wraps `rbx-sync serve`?
- **Resolution:** *(deferred)* v1 ships CLI + plugin only.

### OQ-002: Non-script instances
- **Tags:** `proj-spec/chapters/02-goals-and-non-goals.md#non-goals`, `proj-spec/chapters/06.4-project-mapping.md#what-is-synced`
- **Discovered in:** REVIEW-SPEC prep
- **Blocking:** none
- **Status:** open
- **Question:** Should a later version sync Parts, attributes, tags, or `.rbxmx` models under mapped roots?
- **Resolution:** *(deferred)* v1 syncs `Folder` containers plus `LuaSourceContainer` source only.

### OQ-003: Place build / Open Cloud
- **Tags:** `proj-spec/chapters/02-goals-and-non-goals.md#non-goals`
- **Discovered in:** REVIEW-SPEC prep
- **Blocking:** none
- **Status:** open
- **Question:** Should rbx-sync build `.rbxl` / `.rbxlx` or publish via Open Cloud?
- **Resolution:** *(deferred)* live Roblox Studio session is the v1 sink/source, not place files or cloud.
