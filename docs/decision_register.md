# Decision Register

## Purpose

This register defines how Ante records architecture decisions that affect durable artifacts, runtime authority, trust boundaries, and migration behavior. ADRs are first-class project artifacts. If code changes one of these contracts, the corresponding ADR must be updated in the same change.

## ADR Template Definition

Every ADR must contain the following fields or sections:

- Title
- Status
- Date
- Context
- Options Considered
- Decision
- Consequences
- Migration Impact
- Rollback Notes
- Code Paths Impacted
- Tests Impacted

Template:

```text
# ADR-XXX: Title

Status: accepted|proposed|superseded|rejected
Date: YYYY-MM-DD

## Context
...

## Options Considered
- Option A
- Option B

## Decision
...

## Consequences
...

## Migration Impact
...

## Rollback Notes
...

## Code Paths Impacted
- path/to/code

## Tests Impacted
- path/to/test
```

## Status Model

- `proposed`: actively considered but not yet binding
- `accepted`: binding project decision
- `superseded`: once accepted, now replaced by a newer ADR
- `rejected`: considered and intentionally not adopted

Only `accepted` ADRs define the current architecture. `proposed` ADRs may reserve future decision slots but must not be implemented as though they were settled.

## Per-Decision Fields

Each ADR must answer these practical questions:

- what problem forced the decision
- which options were considered and rejected
- what exact decision is now binding
- what code paths are governed by the decision
- which tests prove the decision remains true
- what migration work is required if the decision touches persisted artifacts
- what rollback is possible and what cannot be rolled back safely

## Index of Accepted Decisions

Accepted now:

- `ADR-001`: ArtifactStore as the sole typed persistence boundary
- `ADR-002`: TranscriptRecord model, chunking, and checkpoint hashing
- `ADR-003`: Session Authority and a single ordered authority envelope
- `ADR-004`: Incremental GameVM reducer boundary for the gameplay screen
- `ADR-005`: TimeAnchorProvider is mandatory for persisted timestamps
- `ADR-006`: Canonical encoding and stable artifact identifiers
- `ADR-007`: Scope guardrails and sandbox-first isolation

## Reserved ADR Entries

Reserved and not yet accepted:

- `ADR-008`: heads-up-only P2P scope
- `ADR-009`: network stack choice
- `ADR-010`: identity key hierarchy
- `ADR-011`: BalanceChain policy

The canonical encoding decision is already occupied by `ADR-006` and is therefore not available as a future placeholder.

## Cross-Reference Rules for Code and Tests

When a code path is materially governed by an ADR:

- reference the ADR file in the pull request description
- include the affected package paths under `Code Paths Impacted`
- include the proving tests under `Tests Impacted`
- if persistent artifact semantics change, update migration notes and compatibility expectations in the ADR

When a test encodes an architectural contract:

- the test name should mention the contract in plain language
- the ADR and the test should evolve together

## Current Documentation Map

Binding architectural documents for Track 0:

- `docs/architecture.md`
- `docs/threat_model.md`
- `docs/glossary.md`
- `docs/decision_register.md`
- `docs/adr/ADR-001-artifact-store.md`
- `docs/adr/ADR-002-transcript-record-and-checkpointing.md`
- `docs/adr/ADR-003-session-authority-and-envelope.md`
- `docs/adr/ADR-004-gamevm-reducer-boundary.md`
- `docs/adr/ADR-005-time-anchor-provider.md`
- `docs/adr/ADR-006-canonical-encoding-and-stable-ids.md`
- `docs/adr/ADR-007-scope-guardrails.md`
