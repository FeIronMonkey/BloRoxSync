# 12. Testing

> **Spec chapter:** Test layers and which package owns them.
> **Parent:** [README.md](../README.md)
> **Index:** [README.md](../README.md)

## Layers

| Layer | Package | What |
|-------|------|------|
| L1 Fixtures | `sync-protocol` | JSON Schema + mapping/http goldens; tests that goldens validate |
| L2 rbx-sync unit | `sync-agent` | `mapfs` vs L1; bind/token; apply/conflict; self-write ignore |
| L3 rbx-sync HTTP | `sync-agent` | `httptest` against [§7.1](07.1-http-api.md) goldens |
| L4 Plugin unit | `studio-plugin` | Pure Luau mapping vs L1; session state machine with mock HTTP |
| L5 Manual / acceptance | this repo §14 | Real Roblox Studio + rbx-sync on Windows; WSL topology T2 |

CI for v1 MUST run L1–L4. L5 is listed in [§14](14-acceptance-criteria.md) and is not required to be automated inside Roblox Studio.

## Naming

Test names follow `<unit><action>_<condition>` (example: `MapFileToNodeReturnsScript_WhenServerLuauSuffix`).

## Spec anchors

New behavioral tests MUST place one `Spec-Anchor` per line citing the chapter heading they encode. Verification sections in chapters are the index; once a test name is claimed, Verification or a matrix MUST name that test.

This spec records **expected** modules:

| Module | Covers |
|--------|--------|
| `sync-protocol` schema tests | §6.1, §8 |
| `sync-agent/internal/mapfs` | §6.4 |
| `sync-agent` serve tests | §6.2, §7.1, §7.2, §6.5 |
| `scripts/install.ps1` / `install.sh` dry-run + checksum | §10.1 |
| `studio-plugin/tests` mapping | §6.4 |
| `studio-plugin/tests` session | §6.3, §6.5 |

## Source ban in tests

Tests MUST NOT import, snapshot, or compare against other sync tools’ fixtures or source.

## Verification

- Spec-Anchor: proj-spec/chapters/12-testing.md#layers
- Spec-Anchor: proj-spec/chapters/12-testing.md#source-ban-in-tests

## Related chapters

- [14. Acceptance criteria](14-acceptance-criteria.md)
