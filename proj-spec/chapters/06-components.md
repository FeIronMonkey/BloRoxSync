# 6. Components

> **Spec chapter:** Index of implementable units.
> **Parent:** [README.md](../README.md)
> **Index:** [README.md](../README.md)

Implement in dependency order: protocol → mapping (shared fixtures) → rbx-sync HTTP/fs → plugin.

| Chapter | Unit | Package |
|---------|------|---------|
| [6.1](06.1-sync-protocol.md) | Wire version, schemas, error envelope | `sync-protocol` |
| [6.2](06.2-sync-agent.md) | CLI, serve, watch, apply | `sync-agent` |
| [6.3](06.3-studio-plugin.md) | Connect UI, DataModel, HTTP client | `studio-plugin` |
| [6.4](06.4-project-mapping.md) | Disk ↔ Roblox Studio node mapping | `sync-protocol` fixtures; both runtimes |
| [6.5](06.5-sync-session.md) | Bootstrap, live apply, conflicts | rbx-sync + plugin |

## Related chapters

- [4. Stack](04-stack.md)
- [7. External API](07-external-api.md)
