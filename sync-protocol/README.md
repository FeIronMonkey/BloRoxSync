# sync-protocol

Versioned **wire contract** for roblox-sync (JSON Schema, mapping fixtures, golden HTTP examples). Both `sync-agent` and `studio-plugin` must conform to the same `protocolVersion`.

JSON Schema 2020-12 files live in `schema/v1/`. L1 tests in `tests/` validate those schemas and `fixtures/v1/http` goldens (Go module plus `github.com/santhosh-tekuri/jsonschema/v6`). Mapping fixtures are a later stage.

## Existing Roblox filesystem-sync tools — source ban

**YOU MUST NOT READ THEIR CODE.** You may read **descriptions** of how existing sync tools work and **API documentation**. You must **not** read any of their source code, even if stuck.

Full rule: [AGENTS.md](AGENTS.md).
