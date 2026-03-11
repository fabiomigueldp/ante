# ADR-007: Scope Guardrails and Sandbox-First Isolation

Status: accepted
Date: 2026-03-11

## Context

The roadmap explicitly warns against drifting into leaderboard, gifting, spectators, chat, mixed-economy work, or premature multiplayer claims before the sandbox is finished and the free-table architecture is ready.

Without a binding scope ADR, product and architecture changes can quietly take dependencies on systems that do not exist yet.

## Options Considered

- Treat roadmap scope as advisory only.
- Allow feature work to proceed opportunistically if UI scaffolding already exists.
- Record binding scope guardrails as an accepted architecture decision.

## Decision

Ante adopts the roadmap scope guardrails as binding project policy.

Binding rules:

- complete the sandbox as a truthful standalone product before multiplayer implementation begins
- do not implement identity, networking, or economy modules in the current slice except for narrow neutral interfaces and documentation reserves
- bots remain sandbox-only
- no leaderboard, gifting, spectators, chat, or mixed-economy work before the free-table alpha scope is complete
- do not make fairness or provable claims before the roadmap's verification track is complete

The first practical consequence is sequencing:

- Track 0 docs
- Track 1 foundations
- sandbox save/load and transcript-backed product integrity
- only then later multiplayer foundations

## Consequences

Positive:

- protects the project from scope creep
- keeps the sandbox shippable on its own
- prevents accidental trust-model drift

Costs:

- some tempting future-facing work must wait
- docs and review discipline become part of normal delivery

## Migration Impact

- none directly, but this ADR constrains where future artifact namespaces and modules may appear

## Rollback Notes

Rolling back this ADR would permit roadmap drift and undermine the architecture work meant to support later tracks.

## Code Paths Impacted

- all future roadmap work
- especially `internal/storage`, `internal/session`, and `internal/tui`

## Tests Impacted

- `internal/docspec/docs_contract_test.go`
