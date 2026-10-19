# 9. Design decisions

> **Spec chapter:** Why this shape instead of alternatives.
> **Parent:** [README.md](../README.md)
> **Index:** [README.md](../README.md)

## D1 — Plugin + localhost rbx-sync, not an IDE extension

Roblox Studio plugins cannot access the host filesystem; OS processes cannot access the DataModel. Official HttpService plugin docs describe localhost companion programs ([references.md](../../references.md)). A VS Code/Cursor extension cannot replace the plugin and is unnecessary if rbx-sync is a CLI (G3).

## D2 — One git repository, three packages

Luau and Go cannot share a build, so plugin and rbx-sync remain separate **packages**. The wire contract lives in a third package, `sync-protocol`, so neither runtime is source of truth and mapping fixtures stay language-neutral ([§3.1](03.1-repository-layout.md)). v1 uses **one git repository** for spec plus those packages: nested remotes add cross-cutting PR cost without independent consumers. `protocolVersion` still versions the wire independently of package semver.

## D3 — Scripts and Folders only in v1

Two-way sync of arbitrary instances is a large product (properties, binary models, unsynced Explorer edits). v1 delivers G1/G2 for Luau, which is the IDE editing loop. Broader instances are [OQ-002](../open-questions.md).

## D4 — Token in the query string

In-experience HttpService documents a very small allowed header set. The protocol therefore authenticates with `accessToken` as a query parameter and only relies on `Content-Type: application/json` ([§6.1](06.1-sync-protocol.md)).

## D5 — Loopback only, unprivileged port 29850

The listener MUST NOT be reachable on the LAN. Port 29850 is ≥ 1024, not 1194, and is not chosen to match any other tool (N7).

## D6 — Explicit bootstrap when trees differ

Silent last-write-wins on first connect can destroy either the git tree or the place. The user must pick Disk, Roblox Studio, or Cancel ([§6.5](06.5-sync-session.md)).

## D7 — Prefer ScriptEditorService for open tabs

Official Script documentation discourages assigning `Source` while the editor owns the buffer ([references.md](../../references.md)).

## D8 — rbx-sync language Go

rbx-sync must run on more than one OS (Windows and Linux/WSL) from **one codebase**, with installers that drop a file on `PATH` rather than a language runtime ([§4](04-stack.md), G4, G9).

Go was chosen because:

- **Cross-compile is first-class.** `GOOS`/`GOARCH` plus `CGO_ENABLED=0` produces a Windows `.exe` and a Linux ELF from the same sources. That is a *different binary per OS*, not a different language or repo.
- **No end-user toolchain.** Python, Node, and many .NET/Java options require a matching runtime on every machine. That fights one-command install on clean Windows and WSL images.
- **Stdlib HTTP** (`net/http`, `encoding/json`) covers the v1 localhost API without a web framework.
- **Filesystem watch** can still use one maintained library that is not taken from a Roblox sync tool.

Rust is a valid alternative (also static, also cross-compiles) but costs more compile time and a heavier contributor toolchain for v1. Go remains the v1 language; changing language would not remove per-OS artifacts.

## D10 — One-command install from release assets

v1 install for **users** is a documented one-liner per OS that downloads checksummed release assets ([§10.1](10.1-install.md)). Marketplace, winget, apt, Homebrew, and a signed `.msi` are allowed later and MUST NOT be required to complete G9. Contributor `go build` remains supported and is not the default user path.

## D9 — Independent mapping names (`.server` / `.client` / `init`)

v1 needs a lossless filename encoding of class. These suffixes are specified here with protocol fixtures. They are **not** an instruction to copy another tool’s project format (N7, source ban).

## Verification

- Spec-Anchor: proj-spec/chapters/09-design-decisions.md#d1--plugin--localhost-rbx-sync-not-an-ide-extension
- `TestLayoutHasPackages_WhenRepoRootResolved`
  Spec-Anchor: proj-spec/chapters/09-design-decisions.md#d2--one-git-repository-three-packages
- `TestLayoutRejectsNestedGit_WhenPackageDirsExist`
  Spec-Anchor: proj-spec/chapters/09-design-decisions.md#d2--one-git-repository-three-packages
- `TestCliServeRejectsBind_WhenHostIsNotLoopback`
  Spec-Anchor: proj-spec/chapters/09-design-decisions.md#d5--loopback-only-unprivileged-port-29850
- `TestCliServeRejectsPort_WhenPortIs1194`
  Spec-Anchor: proj-spec/chapters/09-design-decisions.md#d5--loopback-only-unprivileged-port-29850
- `TestCliServeAcceptsBind_WhenHostIsIPv4Loopback`
  Spec-Anchor: proj-spec/chapters/09-design-decisions.md#d5--loopback-only-unprivileged-port-29850
- Spec-Anchor: proj-spec/chapters/09-design-decisions.md#d3--scripts-and-folders-only-in-v1
- `TestServeCompareEqualTrue_WhenNodeIdClassAndSourceMatch`
  Spec-Anchor: proj-spec/chapters/09-design-decisions.md#d6--explicit-bootstrap-when-trees-differ
- `TestServeCompareEqualFalse_WhenSourceDiffers`
  Spec-Anchor: proj-spec/chapters/09-design-decisions.md#d6--explicit-bootstrap-when-trees-differ
- `TestServeBootstrapDiskToStudioReturnsRevision1_WhenFreshServe`
  Spec-Anchor: proj-spec/chapters/09-design-decisions.md#d6--explicit-bootstrap-when-trees-differ
- `TestServeBootstrapStudioToDiskWritesMappedOnly_WhenNodesSubmitted`
  Spec-Anchor: proj-spec/chapters/09-design-decisions.md#d6--explicit-bootstrap-when-trees-differ
- `TestLayoutHasGoModule_WhenSyncAgentExists`
  Spec-Anchor: proj-spec/chapters/09-design-decisions.md#d8--rbx-sync-language-go
- `TestLayoutDocumentsCgoDisabled_WhenSyncAgentReadmeExists`
  Spec-Anchor: proj-spec/chapters/09-design-decisions.md#d8--rbx-sync-language-go
- `TestLayoutBuildsRbxSyncStub_WhenGoModuleExists`
  Spec-Anchor: proj-spec/chapters/09-design-decisions.md#d8--rbx-sync-language-go
- `TestMappingSourceNamesEncodeClassTable_WhenFixtureExists`
  Spec-Anchor: proj-spec/chapters/09-design-decisions.md#d9--independent-mapping-names-server--client--init
- `TestMapfsClassifySourceNames_WhenFixtureExists`
  Spec-Anchor: proj-spec/chapters/09-design-decisions.md#d9--independent-mapping-names-server--client--init
- Spec-Anchor: proj-spec/chapters/09-design-decisions.md#d10--one-command-install-from-release-assets

## Related chapters

- [2. Goals and non-goals](02-goals-and-non-goals.md)
- [3. Architecture](03-architecture.md)
- [4. Stack](04-stack.md)
- [10.1 Install](10.1-install.md)
