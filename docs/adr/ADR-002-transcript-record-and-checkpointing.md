# ADR-002: TranscriptRecord Model, Chunking, and Checkpoint Hashing

Status: accepted
Date: 2026-03-11

## Context

The current code records in-memory hand history, but replay, history browsing, save/load, and future verification all require a durable, authoritative gameplay record. Session summaries and stats are not rich enough to reconstruct gameplay faithfully.

The roadmap also requires transcript chunking and checkpoint hashing before multiplayer work begins.

## Options Considered

- Continue storing lightweight hand summaries and reconstruct replay heuristically.
- Persist full engine objects directly with gob and use them as replay sources.
- Define transcript-specific records, chunks, checkpoints, and immutable snapshots with canonical encoded hashes.

## Decision

Ante adopts a transcript-first model.

Binding rules:

- `TranscriptRecord` is the authoritative append-only gameplay record.
- History, replay, and results must read transcript-backed artifacts, not infer hand flow from stats or ad hoc UI state.
- The first sandbox chunking policy is one completed hand per chunk.
- Each chunk is hash-linked to the previous committed chunk.
- A hand-boundary checkpoint is computed over canonical encoded chunk bytes and linked to the same session.
- Save/load v1 is allowed only at a committed hand boundary with a corresponding snapshot and checkpoint.

Required minimum record content over time:

- authoritative sequence number
- hand identifier
- event kind
- payload bytes or structured fields sufficient for replay
- linkage to the owning session and transcript chunk

## Consequences

Positive:

- replay becomes deterministic and durable
- save/load has a stable integrity boundary
- future verification and disputes have a compatible substrate

Costs:

- transcript schema and checkpoint logic must be introduced earlier than some UI polish items
- summary-based screens need refactoring to consume transcript-backed projections

## Migration Impact

- current in-memory `SessionHistory` is a runtime helper only, not the final authority
- old stats and save data may migrate into transcript-backed summaries where possible
- any replay feature implemented before transcript persistence must be replaced, not layered on heuristics

## Rollback Notes

Rolling back to summary-only history would reintroduce non-authoritative replay and is therefore not acceptable once transcript-backed features ship.

## Code Paths Impacted

- `internal/session`
- `internal/storage`
- `internal/tui/historyview.go`
- `internal/tui/replay.go`
- `internal/tui/results.go`

## Tests Impacted

- future `internal/session/transcript_test.go`
- future `internal/storage/transcript_test.go`
- future `internal/tui/replay_test.go`
- `internal/docspec/docs_contract_test.go`
