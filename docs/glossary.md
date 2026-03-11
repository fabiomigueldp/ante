# Glossary

## Canonical Definitions

## ArtifactStore

Typed persistence boundary responsible for durable local artifacts such as sandbox saves, transcript chunks, snapshots, session summaries, stats, configuration artifacts, time anchors, and migration manifests. Code outside `internal/storage` must not write raw artifact files directly.

## TranscriptRecord

One append-only authoritative record in a session transcript. A transcript record represents a normalized gameplay or authority event at a specific sequence number and belongs to exactly one transcript chunk.

## TranscriptChunk

A durable append-only segment of transcript records. In the first sandbox implementation, one completed hand maps to one chunk.

## Checkpoint

Hash-linked integrity boundary for a committed transcript chunk. A checkpoint belongs to a deterministic sequencing boundary and is computed over canonical encoded bytes.

## Snapshot

Immutable serialized view of session or table state captured at a known authoritative sequence boundary. Save/load v1 may only use hand-boundary snapshots.

## Session Authority

The component that orders legal actions, validates the active prompt, advances deterministic state, appends transcript records, emits snapshots, and assigns authoritative sequence numbers. In the sandbox this is a single local process. In future persistent-value multiplayer it must not become a sole trust root.

## Prompt Envelope

Typed message describing whose turn it is, which legal actions exist, and which authoritative sequence number the prompt belongs to. A prompt envelope is invalid if its sequence is stale.

## GameVM

The TUI-facing reducer state derived from authoritative envelopes and snapshots. Rendering depends on `GameVM`, not on ad hoc reads across session internals.

## Reducer

Pure state transition function that accepts the current `GameVM` and one typed authority message, then returns the next `GameVM`.

## TimeAnchorProvider

Boundary that supplies timestamps and provenance for persisted or trust-sensitive features. Business logic must not call `time.Now()` directly when creating persisted artifacts.

## Canonical Encoding

Deterministic binary encoding used for bytes that are hashed now or signed later. Raw JSON bytes are not canonical encoding.

## Signed State

Reserved future term for a canonical payload whose exact bytes are signed and verified directly.

## Identity Bundle

Reserved future term for locally stored cryptographic identity material. It does not exist in the sandbox implementation and must not share namespaces with sandbox artifacts.

## BalanceChain

Reserved future term for the signed free-play balance ledger of one identity. It is not part of the sandbox and must remain storage-isolated from sandbox artifacts.

## Dispute Bundle

Reserved future term for the transcript slice, checkpoints, and related signed states needed to prove the latest mutually signed result of a persistent-value session.

## Economy Isolation

Rule that sandbox chips, free-play chips, and any future paid economy remain fully separated in storage, protocol, and UX.

## Free Table Tier

Reserved future policy bucket for free-play heads-up cash tables, including blinds and buy-in limits.

## Session Summary

Durable projection artifact for one finished session. It may feed stats and results, but it is not the source of truth for replaying the hand flow.

## Naming Rules

Artifact naming must be ASCII, lowercase, deterministic, and prefix-based.

Stable identifier formats:

- `session_id`: `ses_<32 lowercase hex>`
- `transcript_id`: `trn_<32 lowercase hex>`
- `chunk_id`: `tch_<32 lowercase hex>_<6-digit chunk index>`
- `checkpoint_id`: `ckp_<32 lowercase hex>_<6-digit hand index>`
- `snapshot_id`: `snp_<32 lowercase hex>_<6-digit hand index>_<9-digit seq>`
- reserved future `identity_id`: `idn_<32 lowercase hex>`
- reserved future `table_id`: `tbl_<32 lowercase hex>`
- reserved future `lock_id`: `lck_<32 lowercase hex>_<6-digit seq>`
- reserved future `settlement_id`: `stl_<32 lowercase hex>_<6-digit seq>`
- reserved future `refill_id`: `rfl_<32 lowercase hex>_<6-digit seq>`

Artifact category names:

- sandbox artifacts live under `sandbox/...`
- local non-gameplay artifacts live under `local/...`
- reserved future free-play artifacts live under `free/...`
- reserved future identity artifacts live under `identity/...`

Preferred names in code and docs:

- use `artifact` for durable typed state
- use `transcript` for replay-critical append-only records
- use `snapshot` for immutable state captures
- use `summary` for derived finished-session projections
- use `prompt envelope` for turn/action availability state

Reserved future names:

- `lock` means a future free-play chip reservation for a table join
- `settlement` means a future co-signed balance-affecting session result
- `refill` means a future free-play balance top-up policy action
- `identity` means a future cryptographic player identity and never a sandbox display name
- `table` means a future network-addressable multiplayer table and never a local menu slot

## Disallowed Ambiguous Terminology

Use these replacements consistently:

- do not say `event log` when you mean `transcript`
- do not say `view state` when you mean `GameVM`
- do not say `save file` when you mean `save artifact`
- do not say `session state` when you mean either `snapshot` or live authority state; choose the precise term
- do not say `wallet`, `bank`, or `balance` for sandbox chips
- do not call sandbox summaries or saves `proofs`

## Practical Definitions That Guide Code

- `history` is a browsing projection built from transcript-backed artifacts.
- `replay` is deterministic reconstruction from transcript-backed artifacts.
- `results` is a post-session presentation built from transcript-backed summaries and metrics.
- `stats` is an aggregate projection over session summaries.
- `save/load v1` means resume from the last committed hand-boundary snapshot only.

## Definitions for Future-Boundary Terms

- `economy isolation`: no conversion and no shared namespace between sandbox and free-play state
- `signed states`: future canonical encoded payloads with direct signature verification
- `dispute bundles`: future artifact bundles containing transcript evidence and mutually signed states
- `free table tiers`: future standardized free-play table policy buckets
