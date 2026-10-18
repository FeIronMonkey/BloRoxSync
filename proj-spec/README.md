# roblox-sync Specification — Chapter Index

## How to use these chapters

- Read **§1–§3** (including §3.1) before writing code — product, architecture, and how the packages in this git repo interact
- Implement **§6** in order: protocol → mapping fixtures → rbx-sync → plugin → session
- Wire details: **§7–§8**; platform: **§10**
- Done-ness: **§12** and **§14**
- Allowed research: [`../references.md`](../references.md) (do not read existing sync-tool source)

## Change history

| Topic | Spec | Plan | Log |
|-------|------|------|-----|
| *(none yet)* | | | |

Reserve `proj-spec/change-specs/`, `change-plans/`, and `change-logs/` for future change-mode work.

## Part I — Product Definition

| Chapter | Title |
|---------|-------|
| [1](chapters/01-overview.md) | Overview |
| [2](chapters/02-goals-and-non-goals.md) | Goals and non-goals |

## Part II — Design

| Chapter | Title |
|---------|-------|
| [3](chapters/03-architecture.md) | Architecture |
| [3.1](chapters/03.1-repository-layout.md) | Repository layout |
| [4](chapters/04-stack.md) | Stack |
| [5](chapters/05-file-tree.md) | File tree |
| [9](chapters/09-design-decisions.md) | Design decisions |

## Part III — Components

| Chapter | Title |
|---------|-------|
| [6](chapters/06-components.md) | Components (index) |
| [6.1](chapters/06.1-sync-protocol.md) | sync-protocol |
| [6.2](chapters/06.2-sync-agent.md) | sync-agent |
| [6.3](chapters/06.3-studio-plugin.md) | studio-plugin |
| [6.4](chapters/06.4-project-mapping.md) | Project mapping |
| [6.5](chapters/06.5-sync-session.md) | Sync session |

## Part IV — Interface

| Chapter | Title |
|---------|-------|
| [7](chapters/07-external-api.md) | External API (index) |
| [7.1](chapters/07.1-http-api.md) | HTTP API |
| [7.2](chapters/07.2-event-stream.md) | Event stream |
| [7.3](chapters/07.3-cli.md) | CLI |
| [8](chapters/08-domain-model.md) | Domain model |

## Part V — Operations

| Chapter | Title |
|---------|-------|
| [10](chapters/10-operations.md) | Operations (index) |
| [10.1](chapters/10.1-install.md) | Install |
| [10.2](chapters/10.2-windows-and-wsl.md) | Windows and WSL |
| [13](chapters/13-performance.md) | Performance |

## Part VI — Validation

| Chapter | Title |
|---------|-------|
| [11](chapters/11-examples.md) | Examples |
| [12](chapters/12-testing.md) | Testing |
| [14](chapters/14-acceptance-criteria.md) | Acceptance criteria |

Open questions (non-blocking backlog): [open-questions.md](open-questions.md)
