# 5. File tree

> **Spec chapter:** Expected paths in this git repository and its packages.
> **Parent:** [README.md](../README.md)
> **Index:** [README.md](../README.md)

Paths are relative to the workspace `roblox-sync/` unless noted.

## This repository (now)

```
roblox-sync/
├── AGENTS.md
├── README.md
├── references.md
├── .gitignore
├── .github/workflows/ci.yml    # L1–L4 hooks (later stages fill tests)
├── scripts/                    # installer scaffolding (Stage 13)
├── proj-spec/
│   ├── README.md
│   ├── open-questions.md
│   ├── implementation-plan.md
│   └── chapters/
├── reviews/
├── tmp/
│   └── tasks.md
├── studio-plugin/              # package (tracked)
├── sync-agent/                 # package (tracked)
└── sync-protocol/              # package (tracked)
```

## Install scripts (target after implementation)

```
scripts/
├── install.ps1             # Windows one-command installer
└── install.sh              # Linux/WSL one-command installer
```

## sync-protocol (target after implementation)

```
sync-protocol/
├── AGENTS.md
├── README.md
├── protocolVersion.txt         # single integer, currently 1
├── schema/
│   └── v1/
│       ├── project.schema.json
│       ├── http.schema.json
│       └── events.schema.json
├── fixtures/
│   └── v1/
│       ├── mapping/            # disk path ↔ node tables
│       └── http/               # request/response goldens
└── tests/                      # L1: schema + golden validation
```

## sync-agent (target after implementation)

```
sync-agent/
├── AGENTS.md
├── README.md
├── go.mod
├── cmd/rbx-sync/main.go
├── internal/
│   ├── layout/                 # Stage 0 repository-layout tests
│   ├── serve/                  # HTTP router + handlers (L3)
│   ├── mapfs/
│   ├── watch/
│   └── installplugin/
└── testdata/
```

Binary name on PATH: `rbx-sync` (Windows: `rbx-sync.exe`).

## studio-plugin (target after implementation)

```
studio-plugin/
├── AGENTS.md
├── README.md
├── src/                        # Luau plugin scripts
└── tests/                      # pure mapping tests (no Roblox Studio)
```

Installable artifact (produced later, not hand-edited as binary): `rbx-sync.rbxmx`.

## Installed files (user machine, after §10.1)

| Environment | rbx-sync | Staged plugin | Roblox Studio plugin copy |
|-------------|----------|---------------|---------------------------|
| Windows | `%LOCALAPPDATA%\rbx-sync\rbx-sync.exe` | `%LOCALAPPDATA%\rbx-sync\rbx-sync.rbxmx` | `%LOCALAPPDATA%\Roblox\Plugins\rbx-sync.rbxmx` |
| Linux / WSL | `$HOME/.local/bin/rbx-sync` | `$HOME/.local/share/rbx-sync/rbx-sync.rbxmx` | Roblox Studio folder per [§6.2](06.2-sync-agent.md) when reachable |

## Project (user source tree, not this workspace)

The user’s **project** is the local source tree edited in an IDE. It is not the Roblox place file ([§1](01-overview.md#trees)).

```
<project>/
├── rbx-sync.json
├── src/
│   ├── server/
│   ├── client/
│   └── shared/
└── .git/
```

`rbx-sync.json` schema lives in `sync-protocol` ([§6.4](06.4-project-mapping.md)).

## Verification

- Spec-Anchor: proj-spec/chapters/05-file-tree.md#this-repository-now
- `TestLayoutHasTargetDirectories_WhenScaffoldingRequired`
  Spec-Anchor: proj-spec/chapters/05-file-tree.md#install-scripts-target-after-implementation
  Spec-Anchor: proj-spec/chapters/05-file-tree.md#sync-protocol-target-after-implementation
  Spec-Anchor: proj-spec/chapters/05-file-tree.md#sync-agent-target-after-implementation
  Spec-Anchor: proj-spec/chapters/05-file-tree.md#studio-plugin-target-after-implementation
- `TestLayoutHasGoModule_WhenSyncAgentExists`
  Spec-Anchor: proj-spec/chapters/05-file-tree.md#sync-agent-target-after-implementation
- `TestLayoutBuildsRbxSyncStub_WhenGoModuleExists`
  Spec-Anchor: proj-spec/chapters/05-file-tree.md#sync-agent-target-after-implementation
- `TestLayoutRecordsProtocolVersion1_WhenProtocolPackageExists`
  Spec-Anchor: proj-spec/chapters/05-file-tree.md#sync-protocol-target-after-implementation
- Spec-Anchor: proj-spec/chapters/05-file-tree.md#installed-files-user-machine-after-101
- Spec-Anchor: proj-spec/chapters/05-file-tree.md#project-user-source-tree-not-this-workspace

## Related chapters

- [3.1 Repository layout](03.1-repository-layout.md)
- [10.1 Install](10.1-install.md)
