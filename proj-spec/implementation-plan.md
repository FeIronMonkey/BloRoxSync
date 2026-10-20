# roblox-sync — Implementation Plan

Spec-only phases complete. Adversarial TDD per stage. Do not start Stage 0 RED until REVIEW-PLAN PASS and user plan-gate approval.

## Spec status

- [x] Chapters written (`proj-spec/chapters/`)
- [x] §14 acceptance criteria
- [x] REVIEW-SPEC PASS (`reviews/roblox-sync-spec-r8.md`)
- [x] User approved spec gate — 2026-10-18

Open questions (`proj-spec/open-questions.md`): OQ-001–OQ-003 remain `open` with `Blocking: none`. They do not block any stage.

## Toolchain pins (plan-time)

These pins fill §4 “allowed but unnamed” slots. Changing a pin without changing behavior is CHECK/deviation Path A, not a spec rewrite.

| Concern | Pin | Constraint |
|---------|-----|------------|
| rbx-sync language | Go 1.23+; `CGO_ENABLED=0` | [§4](chapters/04-stack.md) |
| HTTP | Go `net/http`, `encoding/json` | [§4](chapters/04-stack.md) |
| Filesystem watch | `github.com/fsnotify/fsnotify` | Maintained; not taken from a Roblox sync tool ([§4](chapters/04-stack.md#rbx-sync-dependencies)) |
| L1 schema tests | Go test module `sync-protocol/tests` + JSON Schema 2020-12 library (`github.com/santhosh-tekuri/jsonschema/v6` or successor API-compatible) | [§12](chapters/12-testing.md) L1 lives in `sync-protocol` |
| L4 plugin unit | [Lune](https://lune-org.github.io/docs) running `studio-plugin/tests` (pure Luau; mock HTTP; no Roblox Studio) | [§12](chapters/12-testing.md) L4 |
| Plugin artifact | Stage 12 packages `studio-plugin/src` → `rbx-sync.rbxmx`; tests use a fixture `.rbxmx` until then | [§5](chapters/05-file-tree.md), [§10.1](chapters/10.1-install.md) |
| Release host | GitHub Releases for this repo (or `RBX_SYNC_RELEASE_BASE` in tests) | [§10.1](chapters/10.1-install.md) |

CI MUST run L1–L4 ([§12](chapters/12-testing.md)). L5 (A16, A17, parts of A7–A9, A18, A22–A23) is manual and is a Stage 14 checklist, not a waiver.

## Stages

Each stage: **RED → GREEN → CHECK → commit → REVIEW-STAGE** (PASS required before the next stage). No batched REVIEW-STAGE. Final: `@spec-driven-workflow` → REVIEW-FINAL.

| Stage | Spec | Deliverable | Test module |
|-------|------|-------------|-------------|
| 0 | §3.1, §4, §5, G8, A1–A3 | Package skeletons, `go.mod`, `protocolVersion.txt`, CI hook for L1–L4, layout tests | `sync-agent/internal/layout` |
| 1 | §6.1, §8, §4 JSON | Schemas, error-envelope goldens (every `code`), `protocolVersion` | `sync-protocol/tests` (L1 schema) |
| 2 | §6.4, §11, G7 | Mapping fixture pack (conformance tables) | `sync-protocol/tests` (L1 mapping) |
| 3 | §6.4, §6.2 apply, A5, A10 | Go `mapfs` vs L1 fixtures; LF/BOM/1 MiB | `sync-agent/internal/mapfs` (L2) |
| 4 | §7.3, §6.2 config, A14, D5 | CLI version/help/unknown; serve flags; loopback bind; token print | `sync-agent` CLI + bind tests (L2) |
| 5 | §7.1, §6.1, A4, A6, A13, A15 | HTTP health/session/snapshot; 401; 404 `NOT_FOUND`; mismatch | `sync-agent/internal/serve` (L3) |
| 6 | §6.5 bootstrap, §7.1 bootstrap, A11, A11b, A20 | Compare/apply; revision init; unmapped preserved | `sync-agent/internal/serve` bootstrap (L3) |
| 7 | §6.5 live, A12, §13 size | Apply ops; 409 conflict; `force` rules; `PAYLOAD_TOO_LARGE` | `sync-agent/internal/serve` apply (L3) |
| 8 | §6.2 watch, §7.2, §13 debounce, echo | Watch, SSE, poll, self-write ignore, heartbeat, replace | `sync-agent/internal/watch` + serve events (L2/L3) |
| 9 | §6.4 Luau, A5 L4 | Pure Luau mapping vs L1 fixtures | `studio-plugin/tests` mapping (L4) |
| 10 | §6.3, §6.5 plugin, A11 L4, A17b | Connect state machine; conflict UI; source read/write; LAN reject | `studio-plugin/tests` session (L4) |
| 11 | §6.2 install helper, §7.3 `--file`, A18 | `plugin install` destination resolution | `sync-agent/internal/installplugin` (L2) |
| 12 | §5, §10.1 packaging | Produce `rbx-sync.rbxmx` from `studio-plugin/src` | packaging test (L2) |
| 13 | §10.1, A22–A24, D10 | `scripts/install.ps1` / `install.sh` dry-run + checksum refuse | installer tests (L2) |
| 14 | §4 cross-compile, §10.2, §14, A16–A17, A19–A21 | Dual binaries in CI; out-of-scope checks; L5 manual checklist | cross-build + out-of-scope tests; §14 manual |

## Stage details

### Stage 0 — Scaffolding (§3.1, §4, §5)

| Step | Action |
|------|--------|
| RED | `sync-agent/internal/layout` tests fail until: **repository root** and each of the three package dirs have `AGENTS.md` + `README.md` repeating the source ban (A2); no nested `.git` in packages (A1); `references.md` exists at repo root and lists allowed research sources (A3); `sync-agent/go.mod` module path exists; `sync-protocol/protocolVersion.txt` is a single line `1`; target dirs from [§5](chapters/05-file-tree.md) exist (`schema/v1`, `fixtures/v1/{mapping,http}`, `cmd/rbx-sync`, `internal/{serve,mapfs,watch,installplugin}`, `studio-plugin/src`, `studio-plugin/tests`, `scripts/`). Test names: `<unit><action>_<condition>` ([§12](chapters/12-testing.md#naming)). |
| GREEN | Add `go.mod` (Go 1.23), stub `cmd/rbx-sync/main.go` that is compilable, `protocolVersion.txt`, empty package trees / `.gitkeep` as needed, CI workflow that invokes L1–L4 commands (may be no-op packages until later stages). `CGO_ENABLED=0` documented in `sync-agent/README.md`. |
| CHECK | Root `README.md` status may note scaffolding; **do not** revise normative chapter behavior. Adding `implementation-plan.md` to §5 is optional CHECK-class documentation of a workflow file, not a behavioral change. |

**ACs:** A1, A2, A3, G8.

### Stage 1 — Protocol schemas (§6.1, §8)

| Step | Action |
|------|--------|
| RED | L1 tests: `protocolVersion.txt` == `1`; `project.schema.json`, `http.schema.json`, `events.schema.json` exist and are JSON Schema 2020-12; every error `code` in [§6.1](chapters/06.1-sync-protocol.md#error-codes) appears in at least one `fixtures/v1/http/` golden (including unlisted-path **404** + `NOT_FOUND`); Node / Change / Event / Root / project shapes from [§8](chapters/08-domain-model.md) validate; camelCase keys; revisions as JSON numbers. |
| GREEN | Author schemas and goldens. No rbx-sync HTTP yet. |
| CHECK | Verification expected-test lists already name schema + goldens; update only if fixture filenames are claimed in a chapter matrix. |

**ACs:** A4 (`protocolVersion` is 1 in fixtures), A6 (envelope goldens). Handshake rejection waits for Stage 5.

### Stage 2 — Mapping fixtures (§6.4, §11)

| Step | Action |
|------|--------|
| RED | L1 mapping fixtures fail until they encode: project-file rules (ids, `/` disk paths, no `..`, unique studio, non-overlapping disk); source-name table (`.luau`/`.lua`, `.server`/`.client`/`init.*`); nodeId examples from [§6.4](chapters/06.4-project-mapping.md#nodeid) and [§11](chapters/11-examples.md); inverse mapping; duplicate `Foo.luau`+`Foo.lua` → `PROJECT_INVALID`; at most one `init.*`; forbidden stems; Folder vs init parent. |
| GREEN | `sync-protocol/fixtures/v1/mapping/` pack. Goldens MUST validate against project schema. |
| CHECK | None unless fixture paths are listed in Verification. |

**ACs:** A5 (fixture pack exists). Runtime conformance is Stages 3 and 9.

### Stage 3 — Go mapfs (§6.4, §6.2 apply)

| Step | Action |
|------|--------|
| RED | `sync-agent/internal/mapfs` tests consume Stage 2 fixtures: file↔node, inverse, LF write, BOM strip, CRLF→LF on write, 1 MiB + 1 byte → `PAYLOAD_TOO_LARGE`, atomic write (temp + rename), upsert/delete/rename (same-root rename vs copy+delete), empty mapped dirs removed except disk roots, invalid `rbx-sync.json` → `PROJECT_INVALID`. |
| GREEN | Implement `internal/mapfs`. |
| CHECK | Spec-Anchor lines on new tests per [§12](chapters/12-testing.md#spec-anchors). |

**ACs:** A5 (L2), A10.

### Stage 4 — CLI, bind, token (§7.3, §6.2, D5)

| Step | Action |
|------|--------|
| RED | `rbx-sync version` stdout `agent=<semver> protocol=<n>`; `help` exit 0; unknown command exit 2; `serve` missing project → exit 1 + `PROJECT_INVALID`; `--bind 0.0.0.0` (and any non-loopback) rejected exit 1; `--bind 127.0.0.1` and `--bind ::1` accepted; `--port 1194` rejected; port `< 1024` rejected; bind when address is already in use → exit 1 and stderr mentions address in use ([§7.3](chapters/07.3-cli.md#error-handling)); stderr contains `protocolVersion=`, `listen=`, `accessToken=` + 64 lowercase hex. Path join accepts Windows and POSIX separators ([§10.2](chapters/10.2-windows-and-wsl.md#paths)). |
| GREEN | `cmd/rbx-sync` + config validation. HTTP handlers may still 501 except what Stage 5 needs; bind tests may use a short-lived listener. |
| CHECK | Spec-Anchor on new CLI/bind tests. |

**ACs:** A14. Token 401 on HTTP waits for Stage 5.

### Stage 5 — HTTP core (§7.1, §6.1)

| Step | Action |
|------|--------|
| RED | `httptest` goldens: `GET /v1/health` 200 `{ok, protocolVersion}` no token, body has neither token nor project path; missing token on `/v1/snapshot` → 401 `UNAUTHORIZED`; session protocol mismatch → 409 `PROTOCOL_MISMATCH`; matching session 200 with `roots` **without** disk paths; `GET /v1/snapshot` 200 nodes; unlisted path → **404** `NOT_FOUND` (not `BAD_REQUEST`); invalid JSON → 400 `BAD_REQUEST`; `Content-Type: application/json`. Unknown JSON fields ignored except inside `error`. When a mapped disk root is unreadable, `GET /v1/snapshot` and `POST /v1/apply` return **500** + envelope `IO_ERROR` ([§6.2](chapters/06.2-sync-agent.md#error-handling)); HTTP stays up. |
| GREEN | `internal/serve` router + handlers for health, session, snapshot. |
| CHECK | Spec-Anchor; HTTP goldens under `sync-protocol/fixtures/v1/http/` if not already added in Stage 1. |

**ACs:** A4 (handshake reject), A6, A13, A15. **Also:** L3 `IO_ERROR` when roots unreadable.

### Stage 6 — Bootstrap (§6.5, §7.1 bootstrap)

| Step | Action |
|------|--------|
| RED | Compare by nodeId, className, LF-normalized source; Folders without source; `equal: true` / `false`. `disk-to-studio` MUST NOT accept `nodes`; returns snapshot nodes with revision (1 on fresh serve). `studio-to-disk` requires `nodes`; writes mapped files only; omitted mapped nodes deleted on disk; returns `applied` revision **1**; unmapped Roblox Studio instances / unmapped disk files not deleted (A20). Wrong direction/body mix → `BAD_REQUEST`. Cancel is plugin-only (no HTTP) — covered in Stage 10. |
| GREEN | Bootstrap handlers + revision table ([§6.5](chapters/06.5-sync-session.md#revision-initialization)). |
| CHECK | Spec-Anchor. |

**ACs:** A11 (rbx-sync side), A11b (rbx-sync side), A20.

### Stage 7 — Live apply and conflicts (§6.5 live, §13)

| Step | Action |
|------|--------|
| RED | upsert/delete/rename apply rules; `baseRevision` match → revision+1; new node `baseRevision` 0; stale → 409 `CONFLICT` with `applied` + `conflicts[{nodeId,currentRevision}]`; `force: true` on first-try live → `BAD_REQUEST`; `force: true` after conflict applies; Folder `source` present → `BAD_REQUEST`; rename ack uses `newNodeId`; source > 1 MiB → 413 `PAYLOAD_TOO_LARGE`, no partial write; `nodeId` outside configured roots → 400 `NOT_MAPPED` ([§6.1](chapters/06.1-sync-protocol.md#error-codes)). |
| GREEN | Apply engine. |
| CHECK | Spec-Anchor. |

**ACs:** A9 (L3 apply), A12.

### Stage 8 — Watch, SSE, poll (§6.2, §7.2, §13)

| Step | Action |
|------|--------|
| RED | Debounce 50 ms coalesces two writes into one event; self-write ignored for 500 ms; cursor monotonic from 0; `GET /v1/events` SSE `text/event-stream`, default no backlog, optional `since`; heartbeat `event: heartbeat` ≤ 15 s; `event: end` `shutdown`/`replaced`; `GET /v1/changes` default `since=0` full backlog, non-blocking; origin `disk` on stream; unauthorized stream 401 no SSE body; new session closes prior stream with `SESSION_REPLACED`; no busy loop. |
| GREEN | `internal/watch` + event log + SSE/poll. Pin `fsnotify`. |
| CHECK | Spec-Anchor. |

**ACs:** A7 (L3 events). Manual Roblox Studio remainder is Stage 14 / L5.

### Stage 9 — Plugin mapping (Luau) (§6.4)

| Step | Action |
|------|--------|
| RED | Lune tests load the **same** Stage 2 fixtures as Go: file names, nodeId, inverse, init, duplicate lua/luau refuse, forbidden stems. No Roblox Studio. |
| GREEN | Pure Luau mapping module under `studio-plugin/src` (no HttpService yet). |
| CHECK | Spec-Anchor. |

**ACs:** A5 (L4).

### Stage 10 — Plugin session (§6.3, §6.5)

| Step | Action |
|------|--------|
| RED | Mock HTTP state machine: health fail → status **rbx-sync not reachable**, no further calls (A17b); `PROTOCOL_MISMATCH`/`UNAUTHORIZED` show `message` and stop; compare `equal: true` → `GET /v1/snapshot`, skip bootstrap widget, no `bootstrap/apply`, then live; `equal: false` requires Disk / Roblox Studio / Cancel before live; Cancel → no `bootstrap/apply`, no writes, no live; Disk wins → `POST /v1/bootstrap/apply` `{direction: disk-to-studio}` without `nodes`; Roblox Studio wins → same snapshot as compare in `nodes` with `studio-to-disk`; live MUST NOT start listeners/SSE/poll before step 8. LAN host rejected. Studio-to-disk snapshot **skips** roots whose `studio` path cannot be resolved and surfaces a plugin warning; does not invent disk files ([§6.4](chapters/06.4-project-mapping.md#roblox-studio-path-resolution)). Conflict widget: Keep Roblox Studio → `POST /v1/apply` with `force: true`; Keep disk → plugin mapped state matches `GET /v1/snapshot` for that node (observable outcome; do not invent a new route). Skip leaves both sides unchanged. Poll 200 ms if stream fails; retry stream every 2 s. `SESSION_REPLACED` stops applying. Duplicate `Foo.luau`+`Foo.lua` refuse apply. ScriptEditorService vs `Source` table is unit-tested with fakes. Mapped Roblox Studio mutations batch **50 ms** before `POST /v1/apply` ([§6.3](chapters/06.3-studio-plugin.md#roblox-studio-change-detection), [§13](chapters/13-performance.md#limits)). Plugin **self-apply ignore**: DataModel/source signals caused by applying a disk event for that node are not bounced until the apply finishes ([§6.5](chapters/06.5-sync-session.md#echo-suppression)). **Re-RED (L5 A7 host):** `plugin:CreateToolbar("rbx-sync")` with a ribbon button; dockable widget showing Connect / Disconnect / bootstrap / conflict choices; defaults `127.0.0.1` / `29850`; HttpService `RequestAsync` + `CreateWebStreamClient` with token in the query string; HttpService calls pcall-wrapped; packaged `.rbxmx` Script starts the host with the `plugin` global (not require-all ModuleScripts). **Re-RED (L5 A7 DataModel):** packaged host MUST construct a DataModel + ScriptEditorService adapter (not a no-op stub) so disk-to-studio bootstrap/live create/update/delete appears in Studio; Keep disk / Keep Roblox Studio widget actions call session conflict resolution; host heartbeat flushes the 50 ms mutation batch. **Re-RED (status contrast):** status `TextLabel` MUST use a light `TextColor3` so connect/live/disconnect strings are readable on the dark dock widget. **Re-RED (L5 A7 nested root):** disk-to-studio apply and live create MUST create missing `Folder` segments of a nested `studio` path (for example `ServerScriptService.RbxSyncL5`); studio-to-disk snapshot still skips unresolved paths. |
| GREEN | Plugin connect/session modules + widget strings from spec. **Re-RED GREEN:** `studio-plugin/src/host.luau` + packager Script entry `Host.start({ plugin = plugin, httpService = … })`. **Re-RED GREEN (DataModel):** `studio-plugin/src/studio.luau` adapter; host sibling-requires `studio`; Keep handlers + RunService Heartbeat tick. **Re-RED GREEN (L5 connect status):** session/host status **choose Disk wins, Roblox Studio wins, or Cancel** / **connected** / **disconnected** on Cancel; widget Cancel/Disk/Studio handlers refresh the status label. **Re-RED (status contrast):** status `TextLabel` light `TextColor3` (readable on the dark dock widget). **Re-RED GREEN (status contrast):** `addStatus` sets near-white `TextColor3`. **Re-RED GREEN (L5 A7 nested root):** `studio.luau` creates missing mapped `studio` Folders on `applyNodes` / `createInstance`; `resolveRoot` stays lookup-only for studio-to-disk skip. |
| CHECK | Spec-Anchor. |

**ACs:** A11 (L4), A11b (L4), A17b (L4).

### Stage 11 — Plugin install helper (§6.2, §7.3)

| Step | Action |
|------|--------|
| RED | Windows dest `%LOCALAPPDATA%\Roblox\Plugins\rbx-sync.rbxmx`; WSL: `cmd.exe` `%LOCALAPPDATA%` + `wslpath`, else unique `/mnt/c/Users/*/AppData/Local/Roblox/Plugins`; else print expected Windows path exit 1; no hard-coded username; `--file` default order from [§7.3](chapters/07.3-cli.md#plugin-install-flags). Tests inject dest paths (no real Roblox Studio required). |
| GREEN | `internal/installplugin` + CLI command. Copy may use a testdata `.rbxmx` fixture. Stage 11 before Stage 12 is **intentional**: unit tests do not wait on the real packager; contributor order remains build → package → `plugin install` ([§10.1](chapters/10.1-install.md#contributor-build-optional)). |
| CHECK | Spec-Anchor. |

**ACs:** A18 (L2 path test). Manual copy into a real Plugins folder is L5.

### Stage 12 — Plugin packaging (§5, §10.1)

| Step | Action |
|------|--------|
| RED | Build produces `rbx-sync.rbxmx` from `studio-plugin/src`; artifact is a file the Stage 11 installer can copy; contributor steps in [§10.1](chapters/10.1-install.md#contributor-build-optional) are documented as commands that exist. |
| GREEN | Packager (script or `rbx-sync` helper). Do not vendor another sync tool’s packager. |
| CHECK | File-tree docs if the packager path is new under `scripts/` or `sync-agent`. |

### Stage 13 — One-command installers (§10.1, D10)

| Step | Action |
|------|--------|
| RED | `scripts/install.ps1` `-DryRun` and `scripts/install.sh --dry-run` print destination table for current OS without download/write; WSL uses **linux** asset; checksum mismatch **refuses** replace and exits **non-zero** (A24); `RBX_SYNC_RELEASE_BASE` honored; Windows dest `%LOCALAPPDATA%\rbx-sync\rbx-sync.exe` + `.rbxmx` and PATH note; Linux `$HOME/.local/bin/rbx-sync` + `$HOME/.local/share/rbx-sync/rbx-sync.rbxmx` and `export PATH` line when needed; if plugins folder reachable and `plugin install` fails, installer exits that status (not 0); if unreachable, print staged path and exit 0 after binary install. |
| GREEN | Both scripts. Tests use a fake release dir (fixture sums), not production GitHub. |
| CHECK | Spec-Anchor. |

**ACs:** A22 (L2 dry-run), A23 (L2 dry-run), A24.

### Stage 14 — Cross-compile, out of scope, acceptance (§4, §10.2, §14)

| Step | Action |
|------|--------|
| RED | CI (or `go` test helper) builds `windows/amd64` → `rbx-sync.exe` and `linux/amd64` → `rbx-sync` with `CGO_ENABLED=0`. A documented release recipe emits the three [§10.1](chapters/10.1-install.md#release-assets) files plus `SHA256SUMS` (one path per line) for those artifacts — tests may run the recipe against a temp dir (tagged GitHub publish itself is ops, not a unit test). Tests: no editor extension package shipped (A19); bootstrap/live never deletes unmapped non-script instances (A20, if not already asserted in Stage 6); repo/tests do not import or snapshot other sync tools (A21, G8). Snapshot/bootstrap HTTP completing within 30 s for ≤ 2000 mapped nodes totaling ≤ 20 MiB ([§13](chapters/13-performance.md#limits)) is **L5 / REVIEW-FINAL required**, not an L2 unit bound — record it on the Stage 14 manual checklist; do not waive it. |
| GREEN | Release-oriented build + `SHA256SUMS` recipe; out-of-scope assertions. |
| CHECK | `proj-spec/chapters/14-acceptance-criteria.md` remains the done-ness list. Record L5 manual checklist in `tmp/tasks.md`: A7 (Studio appearance), **A8** (open-tab ScriptEditorService path), A9 (Studio→disk), A16 T1, A17 T2 mirrored, A18/A22/A23 real install, §13 2000-node/30 s bootstrap. **Manual items are required for v1 done-ness and REVIEW-FINAL; they are not deferred.** |

**ACs:** A16, A17 (manual); A8 (manual L5); A19–A21; remaining A7–A9/A18/A22–A23 manual tails.

## Chapter → stage coverage

| Chapter | Stages |
|---------|--------|
| §1 Overview | (product context; no code stage) |
| §2 Goals / non-goals | 14 + AC mapping |
| §3 Architecture | 4, 5, 8, 10 |
| §3.1 Repository layout | 0 |
| §4 Stack | 0, 8 (watcher), 14 (cross-compile) |
| §5 File tree | 0, 12, 13 |
| §6.1 Protocol | 1, 5 |
| §6.2 sync-agent | 3, 4, 5 (`IO_ERROR`), 8, 11 |
| §6.3 studio-plugin | 10 |
| §6.4 Mapping | 2, 3, 9 |
| §6.5 Session | 6, 7, 8, 10 |
| §7.1 HTTP | 5, 6, 7 |
| §7.2 Events | 8, 10 |
| §7.3 CLI | 4, 11 |
| §8 Domain model | 1 |
| §9 Design decisions | D1: 10; D2: 0; D3: 6/14 A20; D4: 5; D5: 4; D6: 6/10; D7: 10; D8: 0/14; D9: 2/3/9; D10: 13 |
| §10.1 Install | 11–13 |
| §10.2 Windows/WSL | 4, 8, 14 (T2 = L5) |
| §11 Examples | 2, 6, 10 |
| §12 Testing | all (layers L1–L4); L5 in 14 |
| §13 Performance | 7 (1 MiB), 8 (disk debounce, heartbeat, poll), 10 (plugin 50 ms debounce), 14 (2000-node/30 s L5) |
| §14 Acceptance | mapped above; leftover manual in 14 |

## REVIEW-SPEC r8 findings — plan handling

Non-blocking findings are **not** spec changes (editing chapters would void spec-gate approval).

| Finding | Plan handling |
|---------|----------------|
| Keep disk HTTP source unspecified | Stage 10 tests the **observable** result: after Keep disk, plugin state matches `GET /v1/snapshot`. Do not add a new protocol route in TDD. If RED cannot encode that without a new wire field, kick back to spec (open-question / chapter), do not guess. |
| Plugin refuse duplicate `Foo.luau`+`Foo.lua` | Stage 2 fixture + Stage 9/10 refuse-apply tests. |
| Checksum mismatch exit code | Stage 13: non-zero exit (A24). |
| Missing Verification Spec-Anchor lines | TDD CHECK may add anchors when a named test is claimed; that is Path A documentation, not a behavioral revision. |
| Watcher unnamed in §4 | Pinned above as `fsnotify`. |
| `implementation-plan.md` absent from §5 | Expected post–REVIEW-SPEC; optional Stage 0 CHECK list only. |

## Dependency checklist

- [x] Shared logic / fixtures before consumers (Stages 1–2 before 3 and 9)
- [x] Config / bind before HTTP (4 before 5)
- [x] Protocol library before HTTP routes (1 before 5)
- [x] Each stage has a named test module
- [x] Each stage references specific chapters
- [x] No stage skips RED or CHECK
- [x] REVIEW-STAGE after each stage (not batched)
- [x] Final stage covers §14 (automated remainder + required L5 checklist)
- [x] Stage 0 is scaffolding code (greenfield)
- [x] Plan file is this durable artifact, not only `tmp/tasks.md`

Normative `MUST` items are **not** waived by “later stage” notes. Later-stage assignment is ordering only.
