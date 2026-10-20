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
| L4 Plugin unit | `studio-plugin` | Pure Luau mapping vs L1; session state machine with mock HTTP; Studio host (toolbar/widget/HttpService) with fake Plugin; DataModel adapter with fake instance tree |
| L5 Manual / acceptance | this repo §14 | Real Roblox Studio + rbx-sync on Windows; WSL topology T2 |

CI for v1 MUST run L1–L4. L5 is listed in [§14](14-acceptance-criteria.md) and is not required to be automated inside Roblox Studio.

## Naming

Test names follow `<unit><action>_<condition>` (example: `MapFileToNodeReturnsScript_WhenServerLuauSuffix`).

## Spec anchors

New behavioral tests MUST place one `Spec-Anchor` per line citing the chapter heading they encode. Verification sections in chapters are the index; once a test name is claimed, Verification or a matrix MUST name that test.

This spec records **expected** modules:

| Module | Covers |
|--------|--------|
| `sync-agent/internal/layout` | §3.1, §4, §5, A1–A3, G8 |
| `sync-protocol/tests` | §6.1, §8, §4 JSON conventions, §6.4 L1 mapping fixtures |
| `sync-agent/internal/mapfs` | §6.4 L2 vs L1 fixtures; §6.2 apply (LF, atomic write, delete, rename) |
| `sync-agent/internal/cli` | §7.3, §6.2 config, A14, D5 |
| `sync-agent/internal/installplugin` | §6.2 plugin install helper, §7.3 `--file`, A18 |
| `sync-agent/internal/packageplugin` | §5 / §10.1 `rbx-sync.rbxmx` packaging |
| `sync-agent/internal/watch` | §6.2 watch debounce, §6.5 self-write ignore, §13 |
| `sync-agent` serve tests | §6.2, §7.1, §7.2, §6.5 |
| `sync-agent/internal/installscripts` | §10.1 `scripts/install.ps1` / `install.sh` dry-run + checksum |
| `sync-agent/internal/release` | §4 cross-compile; §10.1 release assets; A19/A21 |
| `studio-plugin/tests` mapping | §6.4 |
| `studio-plugin/tests` session | §6.3, §6.5 |
| `studio-plugin/tests` host | §6.3 toolbar/widget/HttpService host; connect/cancel/live status strings; light status-label text color; §7.2 `CreateWebStreamClient`; Keep-disk/Studio widget actions; heartbeat tick |
| `studio-plugin/tests` studio | §6.3 DataModel create/delete/source/listeners; §6.4 studio path resolution |

## Source ban in tests

Tests MUST NOT import, snapshot, or compare against other sync tools’ fixtures or source.

## Verification

- `TestLayoutRunsLayerCommands_WhenCIWorkflowExists`
  Spec-Anchor: proj-spec/chapters/12-testing.md#layers
- `TestMappingFixturesLiveInProtocolRepo_WhenPackRequired`
  Spec-Anchor: proj-spec/chapters/12-testing.md#layers
- `PluginMappingNodeIdsMatchSpecExamples_WhenFixtureExists`
  Spec-Anchor: proj-spec/chapters/12-testing.md#layers
- `PluginSessionSkipsBootstrapAndStoresRevisions_WhenCompareEqual`
  Spec-Anchor: proj-spec/chapters/12-testing.md#layers
- `PluginHostCreatesToolbarNamedRbxSync_WhenStarted`
  Spec-Anchor: proj-spec/chapters/12-testing.md#layers
- `PluginHostAppliesDiskBootstrapToDataModel_WhenStudioNotInjected`
  Spec-Anchor: proj-spec/chapters/12-testing.md#layers
- `PluginStudioApplyNodesCreatesScript_WhenUpsertNewNodeId`
  Spec-Anchor: proj-spec/chapters/12-testing.md#layers
- `TestRepoDoesNotDependOnOtherSyncTools_WhenModulesAndTreeWalked`
  Spec-Anchor: proj-spec/chapters/12-testing.md#source-ban-in-tests

## Related chapters

- [14. Acceptance criteria](14-acceptance-criteria.md)
