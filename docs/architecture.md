# Architecture

## Project Scope, Product Boundaries, and Non-Goals

Ante currently ships as a local sandbox poker product. The sandbox must remain a coherent product even if every multiplayer track is delayed. The current implementation target is a deterministic single-process runtime with local persistence, Bubble Tea rendering, AI opponents, and durable artifacts for save/load, transcripts, stats, and replay.

The project has three product domains that must remain explicitly separated:

1. Sandbox: local-only play with bots, ephemeral chips, local saves, local transcripts, and local stats.
2. Free-play multiplayer: future heads-up cash tables with persistent free chips, identity, signatures, networking, and symmetric settlement.
3. Future paid economy: reserved only as an architectural boundary. No paid feature work is in scope in the current roadmap slice.

This document treats the explicit separation between sandbox, free-play multiplayer, and any future paid economy as a hard architectural rule, not a roadmap suggestion.

The separation is not just a UI choice. It is a storage, protocol, and trust-boundary rule:

- sandbox chips never convert into free-play balances
- free-play balances never mix with sandbox artifacts
- paid economy concerns must not leak into sandbox or free-play storage models

Out of scope before the free-table alpha is publicly usable:

- leaderboard features
- gifting or off-table transfers
- spectators
- chat
- mixed-economy tables
- marketing claims about fairness or auditability

## Scope Guardrails

The following guardrails are binding on code and documentation:

- The engine remains deterministic and is the only authority on poker rules.
- The sandbox must work with zero dependency on identity, networking, P2P discovery, relay infrastructure, or balance verification modules.
- Bots are sandbox-only.
- Multiplayer scope, when it starts later, is free-play heads-up cash only.
- Host-authoritative gameplay is forbidden for any persistent-value table. A temporary host-authoritative harness may exist later only for network bringup and must be clearly marked as non-production.
- All trust-sensitive timestamps must flow through `TimeAnchorProvider`.
- All durable local state introduced after Track 1 must flow through `ArtifactStore`.
- TUI state that affects prompts, notices, and legal actions must be presented atomically from one authoritative reducer-facing stream.

## Package Map

Current packages:

- `cmd/ante`: local sandbox TUI entrypoint.
- `cmd/sim`: deterministic simulation runner for engine verification.
- `internal/engine`: pure poker rules, deterministic state transitions, betting, showdown, blind logic, hand history primitives, tournament and cash-game rules.
- `internal/session`: local session authority, bot orchestration, snapshot emission, transcript append, session metrics, save/load reconstruction.
- `internal/tui`: rendering and input only. The TUI consumes reducer-facing state and emits player intent; it does not own persistence or rules.
- `internal/storage`: `ArtifactStore`, migrations, typed artifact encoding, manifests, time anchors, local persistence, and storage-backed query helpers.
- `internal/audio`: local presentation concern only.

Reserved future packages, not to be implemented in the current slice except as documentation references or narrow interfaces:

- `internal/identity`: local identity material, key delegation, and recovery flows.
- `internal/protocol`: future wire types and canonical signed payload definitions.
- `internal/network`: future transport, discovery, reconnect, and peer-scoring runtime.
- `internal/freeplay`: future free-play economy verification and `BalanceChain` logic.

Ownership rules:

- `internal/engine` must not import TUI, storage, or networking packages.
- `internal/session` may depend on `engine` and `storage`, but not on renderer details.
- `internal/tui` may depend on `session` reducer-facing contracts and storage-backed queries, but it must not write raw files.
- `internal/storage` must not depend on TUI.
- Future identity, network, and economy packages must not back-reference the TUI.

## Authoritative Runtime Data Flow

The target runtime flow is:

`user input -> session authority -> transcript -> snapshot -> reducer -> renderer`

Each stage has one responsibility:

- user input: the TUI captures intent only
- session authority: validates the active prompt, applies engine transitions, assigns sequence numbers, and decides when artifacts are appended
- transcript: append-only authoritative event record for replay, debugging, save/load linkage, and future verification
- snapshot: immutable state capture at a known sequence boundary
- reducer: produces `GameVM` from the ordered authority envelope
- renderer: draws only from `GameVM`; it does not infer authority by stitching together loose fields

This flow replaces the current split-brain pattern where prompt state and event state can arrive on different channels and then be merged heuristically in the gameplay model.

## ArtifactStore Responsibilities, Artifact Types, and Namespacing Rules

`ArtifactStore` is the only typed persistence boundary for durable local data after Track 1. Its responsibilities are:

- create, load, list, and delete sandbox save artifacts
- append and query transcript chunks, checkpoints, and summaries
- persist sandbox stats and session summaries
- persist local configuration artifacts
- persist time-anchor artifacts and migration manifests
- expose explicit version metadata and compatibility results

Initial local namespaces:

- `sandbox/saves/`
- `sandbox/transcripts/`
- `sandbox/history/`
- `sandbox/stats/`
- `local/config/`
- `local/time_anchors/`
- `local/migrations/`

Reserved future namespaces:

- `identity/`
- `free/`
- `cache/`
- `disputes/`

Rules:

- The TUI and session layers never write gob or JSON files directly.
- Feature code does not invent ad hoc filenames; it asks `ArtifactStore` for typed operations.
- Physical filenames are implementation details. Artifact semantics, versioning, and identifiers are the stable contract.

## Transcript and Snapshot Lifecycle

`TranscriptRecord` is the authoritative append-only gameplay log. It is the source of truth for replay, history browsing, debugging, and future dispute tooling. Session summaries and stats are projections derived from transcript-backed facts; they are not the authority for reconstructing gameplay.

Initial sandbox lifecycle:

1. `Session Authority` emits ordered authority envelopes with a monotonic sequence number.
2. Engine-derived events are normalized into transcript records.
3. Records are accumulated into a transcript chunk for one completed hand.
4. At the hand boundary, the runtime commits:
   - the transcript chunk
   - a checkpoint hash for that chunk
   - an immutable snapshot linked to the same session and hand boundary
5. The session authority enters an explicit between-hands synchronization barrier and waits for readiness before the next hand begins.
6. Save/load v1 may only target these hand-boundary snapshots and resumes into the waiting-ready boundary rather than auto-starting the next hand.

Transcript chunking and checkpoint hashing rules:

- Chunking unit in the sandbox is one completed hand.
- Each chunk is linked to the previous chunk hash.
- Each checkpoint is computed over canonical bytes, never raw JSON.
- Checkpoints are created only at deterministic sequencing boundaries.
- The barrier between hands is an authoritative sync boundary owned by `Session Authority`, not by the renderer.
- Session-control intents such as next-hand readiness or leaving the table are session-layer lifecycle inputs, not engine poker actions.
- Replay, history, and results must read transcript-backed artifacts, not best-effort reconstructions from summaries.

## Stable Linkage Between Session, Snapshot, and Transcript Identifiers

Identifiers must be stable enough that tooling, migrations, and future verification layers can reason about related artifacts without filename guessing.

The accepted identifier model is:

- `session_id`: one immutable identifier for the whole sandbox run, format `ses_<32 lowercase hex>`
- `transcript_id`: exactly one transcript stream for a session, format `trn_<32 lowercase hex>` and derived from the same base token as the session
- `chunk_id`: one per transcript chunk, format `tch_<32 lowercase hex>_<6-digit chunk index>`
- `checkpoint_id`: one per hand-boundary checkpoint, format `ckp_<32 lowercase hex>_<6-digit hand index>`
- `snapshot_id`: one immutable snapshot boundary, format `snp_<32 lowercase hex>_<6-digit hand index>_<9-digit seq>`

Required relationships:

- one session owns one transcript stream
- one completed hand yields at most one committed chunk, one checkpoint, and one save-eligible snapshot in v1
- stats and session summaries must reference `session_id` and the latest committed transcript/checkpoint identifiers

## TimeAnchorProvider Contract, Provenance, and Error Handling

`TimeAnchorProvider` is mandatory infrastructure, not polish. Persisted timestamps and any later trust-sensitive timing logic must not call `time.Now()` directly from business logic.

The contract must provide:

- a timestamp value
- provenance metadata describing the source, such as `local_clock`
- explicit error surfaces when a timestamp cannot be produced safely

Initial sandbox implementation:

- `LocalTimeAnchorProvider` may use the local system clock
- every persisted artifact must store both the timestamp and the provider provenance
- local-clock trust is acceptable for sandbox artifacts because sandbox chips have no persistent value

Future implication:

- later free-play refill, lock, settlement, and identity flows may replace or augment the provider, but the calling code should not change its shape

## Free-Table Multiplayer Layering and Future Boundaries

Multiplayer is not part of the current implementation slice, but the sandbox architecture must reserve clean seams for it. The intended layering is:

- `engine`: deterministic rules only
- `session`: local authority today, network-aware authority later
- `protocol`: future canonical payload definitions and wire envelopes
- `network`: future transport and reconnect runtime
- `freeplay`: future balance verification and settlement rules

The sandbox runtime must not preload or depend on these future modules. The code written now should only reserve neutral interfaces and durable artifact semantics that later modules can reuse.

## BalanceChain Overview, Identity Boundaries, and Networking Overview

`BalanceChain`, identity bundles, session-key delegation, and networking are future concerns. They matter now only as constraints on what the sandbox must not do.

- The sandbox transcript model must be replayable from its own artifacts alone.
- Sandbox artifacts must not masquerade as signed multiplayer evidence.
- Future free-play balances will require a separate trust model, separate namespaces, and symmetric settlement rules.
- No current sandbox artifact may be treated as a balance proof, lock proof, or settlement proof.

## Failure Domains, Crash Recovery, Migrations, and Trust Roots

Current sandbox trust roots:

- the local process for session authority
- the deterministic engine for rules
- the local filesystem for artifact durability
- the local time anchor for timestamps

Crash recovery boundaries:

- the runtime may recover only from committed hand-boundary artifacts in the first save/load slice
- a resumed hand-boundary save returns to the between-hands waiting-ready state and never auto-starts the next deal
- mid-hand state is explicitly unsupported in v1 and must fail loudly
- transcript chunks and snapshots must be committed in an order that avoids a save pointing at a nonexistent checkpoint

Migration expectations:

- legacy `config.json`, `stats.gob`, and `saves/slot_*.gob` are migration sources only
- migration must either succeed deterministically or return a typed compatibility error visible to the user
- migration manifests belong to `ArtifactStore`, not to feature-local code

Trust assumptions that do not carry into multiplayer:

- the local sandbox may trust its own process ordering
- the local sandbox may trust bot actions because bots are local-only entities
- the local sandbox may use a local clock for timestamps
- none of these assumptions are sufficient for persistent-value tables later
