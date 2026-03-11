# ADR-001: ArtifactStore as the Sole Typed Persistence Boundary

Status: accepted
Date: 2026-03-11

## Context

The current project persists configuration as JSON and save/stat data as feature-local gob files. That model couples persistence details to individual features and makes migrations, compatibility checks, and future transcript-backed flows harder than necessary.

The roadmap requires a typed persistence boundary that can serve the sandbox now and reserve space for future identity and free-play artifacts later.

## Options Considered

- Keep feature-local file formats and add more helper functions beside them.
- Introduce a general-purpose key/value store and let features define their own record shapes.
- Introduce `ArtifactStore` as a typed boundary with explicit artifact kinds, namespaces, versioning, and migration support.

## Decision

Ante adopts `ArtifactStore` as the only typed persistence boundary for durable local artifacts after Track 1.

Binding rules:

- All new durable state flows through `internal/storage`.
- TUI and session code may request typed operations, but they do not write files directly.
- Config, saves, stats, transcript chunks, checkpoints, snapshots, session summaries, time anchors, and migration manifests are all artifact types.
- The first concrete implementation is a local filesystem-backed store.
- Artifact semantics are stable; physical filenames and encodings are internal details.

Namespaces:

- `sandbox/*` for sandbox gameplay artifacts
- `local/*` for local non-gameplay artifacts such as config, time anchors, and migrations
- `identity/*`, `free/*`, and `cache/*` are reserved future namespaces and are not to be implemented in this PR slice

## Consequences

Positive:

- one compatibility layer for migration and version checks
- one place to enforce naming, versioning, and manifest rules
- cleaner separation between runtime logic and persistence details

Costs:

- short-term refactor work to adapt existing config, save, and stats APIs
- more upfront structure before user-facing features are completed

## Migration Impact

- `config.json`, `stats.gob`, and `saves/slot_*.gob` become migration sources
- compatibility must be explicit, deterministic, and user-visible on failure
- migration manifests are owned by `ArtifactStore`

## Rollback Notes

Rollback is safe only before new artifact-backed data is written. Once artifacts and migration manifests exist, rollback requires keeping compatibility readers or explicit cleanup tooling.

## Code Paths Impacted

- `internal/storage`
- `internal/session`
- `internal/tui/loadgame.go`
- `internal/tui/statsview.go`
- `internal/tui/game.go`

## Tests Impacted

- future `internal/storage/artifact_store_test.go`
- future `internal/storage/migration_test.go`
- `internal/docspec/docs_contract_test.go`
