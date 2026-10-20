# 14. Acceptance criteria

> **Spec chapter:** When v1 is done.
> **Parent:** [README.md](../README.md)
> **Index:** [README.md](../README.md)

All items are required unless marked manual. Automated items map to [§12](12-testing.md) L1–L4.

## Packages and process

| ID | Criterion | Verify |
|----|-----------|--------|
| A1 | Three package directories exist: `sync-protocol`, `sync-agent`, `studio-plugin` | Directories in this git repo; no nested `.git` |
| A2 | Root plus each package has `AGENTS.md` and `README.md` stating the source ban | File read |
| A3 | `references.md` lists allowed research sources | File read |

## Protocol and mapping

| ID | Criterion | Verify |
|----|-----------|--------|
| A4 | `protocolVersion` is 1; handshake rejects mismatch | L3 HTTP test |
| A5 | Mapping tables in §6.4 hold for fixture pack | L1 + L2 + L4 |
| A6 | Error envelope and codes in §6.1 | L3 |

## Disk → Roblox Studio (G2)

| ID | Criterion | Verify |
|----|-----------|--------|
| A7 | Create/update/delete/rename of mapped files appears in Roblox Studio after live session | L3 events + **manual** Roblox Studio |
| A8 | Open script tabs update via ScriptEditorService path (manual) | **manual** |

## Roblox Studio → disk (G1)

| ID | Criterion | Verify |
|----|-----------|--------|
| A9 | Create/update/delete/rename of mapped scripts appears on disk | L3 apply + **manual** Roblox Studio |
| A10 | LF written on disk; UTF-8 | L2 |

## Session safety

| ID | Criterion | Verify |
|----|-----------|--------|
| A11 | Unequal trees require Disk / Roblox Studio / Cancel. Cancel writes nothing and does not enter live. Disk wins and Roblox Studio wins call `POST /v1/bootstrap/apply` before live | L3 + plugin L4 |
| A11b | Compare `equal: true` enters live without bootstrap writes; plugin revisions match `GET /v1/snapshot` | L3 + plugin L4 |
| A12 | Stale `baseRevision` → 409; `force` only after conflict | L3 |
| A13 | Health leaks neither token nor project path | L3 |
| A14 | `--bind 0.0.0.0` (and any non-loopback) rejected; `--bind 127.0.0.1` and `--bind ::1` accepted | L2 |
| A15 | Missing token on `/v1/snapshot` → 401 | L3 |

## Platform

| ID | Criterion | Verify |
|----|-----------|--------|
| A16 | Topology T1 (Windows rbx-sync) bidirectional mapped-script round-trip works | **manual** |
| A17 | Topology T2 with WSL **mirrored** networking (or the equivalent documented in [references.md](../../references.md)): Roblox Studio `GET /v1/health` succeeds, handshake succeeds, and a bidirectional mapped-script round-trip works. This is **required** for v1 done-ness (G4); a disconnected plugin does **not** satisfy A17. | **manual** |
| A17b | When health fails (any topology), the plugin stays disconnected with status **rbx-sync not reachable** (same string as [§6.3](06.3-studio-plugin.md#connect-flow) health failure) and no sync proceeds | L4 + **manual** |
| A18 | `rbx-sync plugin install` writes under `%LOCALAPPDATA%\Roblox\Plugins` on Windows | L2 path test + **manual** |
| A22 | Windows one-command install (`scripts/install.ps1`) places `rbx-sync.exe` and `rbx-sync.rbxmx` under `%LOCALAPPDATA%\rbx-sync\`, adds that directory to the user PATH when missing, and runs `plugin install` | L2 dry-run + **manual** download |
| A23 | Linux/WSL one-command install (`scripts/install.sh`) places `rbx-sync` under `$HOME/.local/bin/` and `rbx-sync.rbxmx` under `$HOME/.local/share/rbx-sync/`, prints an `export PATH` line if `$HOME/.local/bin` is not on `PATH`, and runs `plugin install` when the Roblox Studio plugins folder is reachable | L2 dry-run + **manual** download |
| A24 | Installer refuses to replace a binary or `.rbxmx` when SHA-256 does not match `SHA256SUMS` | L2 with fixture sums |

## Out of scope checks (must remain true)

| ID | Criterion |
|----|-----------|
| A19 | No editor extension shipped |
| A20 | Non-script instances under a mapped root are not deleted by bootstrap |
| A21 | No dependency on another sync tool’s source or project format |

v1 is **done** when A1–A24 (including A11b and A17b) hold and REVIEW-FINAL for the implementation plan has PASS.

## Verification

- `TestLayoutHasPackages_WhenRepoRootResolved` — A1
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#packages-and-process
  Spec-Anchor: proj-spec/chapters/03.1-repository-layout.md#packages
- `TestLayoutRejectsNestedGit_WhenPackageDirsExist` — A1
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#packages-and-process
  Spec-Anchor: proj-spec/chapters/03.1-repository-layout.md#packages
- `TestLayoutRepeatsSourceBan_WhenRootAndPackageDocsExist` — A2, G8
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#packages-and-process
  Spec-Anchor: proj-spec/chapters/03.1-repository-layout.md#packages
  Spec-Anchor: proj-spec/chapters/02-goals-and-non-goals.md#goals
- `TestLayoutListsAllowedSources_WhenReferencesExists` — A3, G8
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#packages-and-process
  Spec-Anchor: proj-spec/chapters/02-goals-and-non-goals.md#goals
  Spec-Anchor: proj-spec/chapters/05-file-tree.md#this-repository-now
- `TestProtocolRecordsVersion1_WhenProtocolVersionFileExists` — A4 (`protocolVersion` is 1)
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#protocol-and-mapping
- `TestErrorGoldensCoverEveryCode_WhenHttpFixturesExist` — A6 (L1 envelope goldens)
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#protocol-and-mapping
- `TestMappingValidProjectsValidate_WhenFixturePackExists` — A5 (L1 fixture pack)
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#protocol-and-mapping
- `TestMappingSourceNamesEncodeClassTable_WhenFixtureExists` — A5
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#protocol-and-mapping
- `TestMappingNodeIdsMatchSpecExamples_WhenFixtureExists` — A5
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#protocol-and-mapping
- `TestMappingInversePrefersLuau_WhenCreatingFiles` — A5
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#protocol-and-mapping
- `TestMapfsNodesMatchSpecExamples_WhenFixtureExists` — A5 (L2)
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#protocol-and-mapping
- `TestMapfsClassifySourceNames_WhenFixtureExists` — A5 (L2)
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#protocol-and-mapping
- `TestMapfsInversePrefersLuau_WhenCreatingFiles` — A5 (L2)
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#protocol-and-mapping
- `TestMapfsUpsertWritesLF_WhenSourceHasCRLF` — A10
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#roblox-studio--disk-g1
- `TestServeApplyUpsertsSource_WhenBaseRevisionMatches` — A9 (L3 upsert)
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#roblox-studio--disk-g1
- `TestServeApplyDeletesFile_WhenBaseRevisionMatches` — A9 (L3 delete)
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#roblox-studio--disk-g1
- `TestServeApplyRenameAcksNewNodeId_WhenBaseRevisionMatches` — A9 (L3 rename)
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#roblox-studio--disk-g1
- `TestServeChangesOriginDisk_WhenFileWritten` — A7 (L3 disk events)
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#disk--roblox-studio-g2
- `TestServeBootstrapDiskToStudioReturnsRevision1_WhenFreshServe` — G2 (disk snapshot for plugin apply)
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#disk--roblox-studio-g2
- `TestServeBootstrapStudioToDiskWritesMappedOnly_WhenNodesSubmitted` — G1 (studio snapshot written to disk)
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#roblox-studio--disk-g1
- `TestCliServeRejectsBind_WhenHostIsNotLoopback` — A14
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#session-safety
- `TestCliServeAcceptsBind_WhenHostIsIPv4Loopback` — A14
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#session-safety
- `TestCliServeAcceptsBind_WhenHostIsIPv6Loopback` — A14
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#session-safety
- `TestServeHealthOmitsTokenAndProject_WhenUnauthenticated` — A13
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#session-safety
- `TestServeSnapshotUnauthorized_WhenTokenMissing` — A15
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#session-safety
- `TestServeSessionMismatch_WhenProtocolDiffers` — A4 (handshake reject)
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#protocol-and-mapping
- `TestServeCompareEqualTrue_WhenNodeIdClassAndSourceMatch` — A11b (rbx-sync: equal compare does not write disk)
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#session-safety
- `PluginSessionSkipsBootstrapAndStoresRevisions_WhenCompareEqual` — A11b (L4)
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#session-safety
- `TestServeBootstrapDiskToStudioReturnsRevision1_WhenFreshServe` — A11 (rbx-sync disk-wins apply)
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#session-safety
- `PluginSessionCancelWritesNothing_WhenUnequal` — A11 (L4 Cancel)
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#session-safety
- `PluginSessionDiskWinsOmitsNodes_WhenApplyingBootstrap` — A11 (L4 Disk wins)
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#session-safety
- `PluginSessionStudioWinsResendsCompareSnapshot_WhenApplyingBootstrap` — A11 (L4 Roblox Studio wins)
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#session-safety
- `TestServeBootstrapStudioToDiskWritesMappedOnly_WhenNodesSubmitted` — A11 (rbx-sync studio-wins apply), A20
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#session-safety
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#out-of-scope-checks-must-remain-true
- `TestServeApplyReturnsConflict_WhenBaseRevisionStale` — A12
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#session-safety
- `TestServeApplyRejectsForce_WhenFirstTryLive` — A12
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#session-safety
- `TestServeApplyForceWrites_WhenAfterConflict` — A12
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#session-safety
- `PluginSessionShowsNotReachable_WhenHealthFails` — A17b (L4)
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#platform
- `PluginSessionMakesNoFurtherCalls_WhenHealthFails` — A17b (L4)
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#platform
- `TestInstallpluginResolvesWindowsDest_WhenGOOSIsWindows` — A18 (L2)
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#platform
- `TestInstallPs1DryRunPrintsWindowsDest_WhenDryRun` — A22 (L2 dry-run)
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#platform
- `TestInstallShDryRunPrintsLinuxDest_WhenDryRun` — A23 (L2 dry-run)
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#platform
- `TestInstallShHasUnixLineEndings_WhenCheckedIn` — A23 (LF so POSIX `sh` on WSL can run)
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#platform
- `TestInstallPs1RefusesReplace_WhenChecksumMismatches` — A24
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#platform
- `TestInstallShRefusesReplace_WhenChecksumMismatches` — A24
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#platform
- `TestRepoShipsNoEditorExtension_WhenTreeWalked` — A19
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#out-of-scope-checks-must-remain-true
- `TestServeBootstrapStudioToDiskWritesMappedOnly_WhenNodesSubmitted` — A20
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#out-of-scope-checks-must-remain-true
- `TestRepoDoesNotDependOnOtherSyncTools_WhenModulesAndTreeWalked` — A21
  Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#out-of-scope-checks-must-remain-true

## Related chapters

- [2. Goals and non-goals](02-goals-and-non-goals.md)
- [12. Testing](12-testing.md)
