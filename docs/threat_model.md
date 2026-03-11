# Threat Model

## Assets and Trust Boundaries

The current product is a local sandbox, but the architecture must already separate local-only trust from future multiplayer trust. The main assets are:

- deterministic engine correctness
- transcript integrity
- snapshot integrity
- save/load artifact integrity
- stats and session summary integrity
- local configuration integrity
- time-anchor provenance
- future identity and economy boundaries, even before they exist in code

Current trust boundaries:

- local process boundary: trusted for sandbox sequencing only
- local filesystem boundary: untrusted for tamper resistance, trusted only as storage medium
- local clock boundary: acceptable for sandbox timestamps, not sufficient for future economic decisions
- future network boundary: fully untrusted until explicitly designed and verified

## Actor Model

Relevant actors:

- Local player: can modify local files, kill the process, or run altered binaries.
- Remote peer: future multiplayer counterparty; must be treated as potentially Byzantine.
- Relay: future transport helper that may fail, censor, delay, or reorder traffic.
- Bootstrap node: future discovery dependency that may be unavailable, malicious, or outdated.
- Storage attacker: any actor with read/write access to local artifact files.
- Clock manipulator: any actor who can skew the local clock or intercept time-provider dependencies.
- Protocol attacker: any actor who sends malformed, replayed, equivocated, or out-of-order messages.

## Sandbox Isolation and Why It Matters

The current sandbox runtime is intentionally simpler and more trusting than the future multiplayer runtime. That is acceptable only because sandbox chips are ephemeral and local. The sandbox must remain isolated because:

- the current local process is the sole session authority
- the current filesystem has no cryptographic tamper resistance
- the current TUI is optimized for local UX, not adversarial message handling
- bots are local entities and are not governed by multiplayer fairness rules

If sandbox artifacts were allowed to bridge into free-play balances later, the architecture would inherit trust assumptions that are invalid for any persistent-value table.

## Persistence Attack Surface

Persistence risks in the current and near-future sandbox:

- modified save artifacts that attempt to load impossible table states
- stale or partially migrated config artifacts
- corrupted transcript chunks or missing checkpoints
- mismatched snapshot and transcript references
- deletion or truncation of migration manifests

Required mitigations:

- typed artifact loading through `ArtifactStore`
- explicit versioning and compatibility checks
- deterministic linkage between session, transcript, checkpoint, and snapshot identifiers
- loud failure on incompatible or unsupported artifacts
- transcript-backed replay and results rather than inference from summaries

## Transcript, Replay, and Snapshot Attack Surface

Threats:

- transcript tampering after the fact
- replaying stale prompt state into a newer UI state
- reconstructing replay from incomplete summaries instead of authoritative artifacts
- loading a snapshot whose referenced checkpoint does not exist or does not match

Required mitigations:

- append-only transcript model
- hand-boundary chunking and checkpoint hashing
- canonical encoding for any bytes that are hashed now or signed later
- reducer-side stale prompt rejection based on authoritative sequence numbers
- history, replay, and results sourced from transcript/snapshot artifacts only

## Identity, Networking, and Balance Verification Attack Surface

These are future attack surfaces, but the current architecture must reserve the right boundaries now.

Threats that matter later:

- impersonation of player identity
- forged session-key delegation
- network message replay or reordering
- equivocation about the latest balance head
- conflicting settlements or checkpoints
- relay censorship or DHT poisoning

Current design implication:

- do not bake sandbox-specific trust assumptions into artifact semantics
- do not make the TUI the owner of trust-critical state transitions
- do not create local-only shortcuts that later need to become signed multiplayer evidence

## Replay, Tampering, Impersonation, Equivocation, and Clock-Skew Abuse

Replay:

- stale prompts or duplicated authority messages can surface illegal actions if the UI state is not derived from one ordered stream

Tampering:

- local artifacts can be edited by a user or malware; the sandbox may detect incompatibility and corruption but cannot treat the local disk as a trust root

Impersonation:

- not applicable to sandbox gameplay today, but future identity artifacts must never be inferred from sandbox player names or local save metadata

Equivocation:

- future free-play systems must detect conflicting `BalanceChain` heads or checkpoints; the sandbox transcript model should already be deterministic and unambiguous so it can later feed dispute bundles without semantic drift

Clock-skew abuse:

- sandbox timestamps are informational, but any future refill or lock logic would be vulnerable if code paths continue to read `time.Now()` directly without provenance

Lock-doubling:

- future free-play tables must prevent more than one active session lock per identity; the current sandbox must not introduce any wallet-like or lock-like value system that could blur this boundary

## Bots Are Sandbox-Only by Design

Bots are allowed in the sandbox because the sandbox is a local product. Bots are excluded from multiplayer scope because:

- they would complicate fairness and identity claims
- they would distort the trust model for P2P tables
- they would blur whether a remote action came from a person or an automated agent

This roadmap does not require the protocol to detect or block bots. It requires that multiplayer product design not include them.

## Host-Authoritative Bringup Harness Risks

Later network bringup may use a host-authoritative harness to prove transport and reconnect mechanics. That harness carries explicit risks:

- one side becomes an unjustified trust root
- state ordering can diverge from symmetric transcript expectations
- economic outcomes could be faked or selectively omitted

Therefore:

- host-authoritative bringup is allowed only as a temporary development harness later
- it must be labeled as insecure and non-production
- it is prohibited for any table where chips have persistent value

## Incident Response Expectations

Expected responses for the current roadmap stage:

- corrupted or incompatible artifacts must fail closed with a visible error
- migrations must be logged through manifests so failures can be diagnosed
- transcript and snapshot mismatches must block replay and resume rather than guessing
- documentation must remain explicit about unsupported states, especially mid-hand resume in the first slice

Future incident response areas already reserved by this threat model:

- identity compromise
- transcript corruption at rest
- time-anchor outage or drift
- settlement mismatch or equivocation
- discovery or relay outage

## Unresolved Risks

Known unresolved risks at this stage:

- local filesystem tampering cannot be fully prevented in the sandbox
- local time provenance is weak compared to later persistent-value requirements
- the current codebase still contains split UI/session state that can produce stale prompt behavior until the reducer boundary lands
- until transcript-backed history and replay are implemented, some existing UI surfaces remain truthful only at the scaffolding level

These risks are acceptable only because the current product scope is the sandbox and sandbox chips have no transferable value.
