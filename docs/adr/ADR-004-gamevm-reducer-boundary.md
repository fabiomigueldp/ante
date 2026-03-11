# ADR-004: Incremental GameVM Reducer Boundary for the Gameplay Screen

Status: accepted
Date: 2026-03-11

## Context

`internal/tui/game.go` currently mixes transport handling, state mutation, and rendering. The roadmap requires a TUI-facing reducer such as `GameVM`, but a full rewrite of every screen before the first vertical slice would slow delivery and broaden risk.

## Options Considered

- Rewrite the entire TUI around a global reducer before landing save/load.
- Leave the gameplay screen as-is and postpone reducer work until much later.
- Introduce `GameVM` incrementally, starting only with the gameplay screen and the prompt/log/legal-action atomicity problem.

## Decision

Ante adopts an incremental `GameVM` reducer rollout.

Binding rules:

- the first reducer scope is `internal/tui/game.go` only
- the reducer consumes the single ordered authority envelope from `Session Authority`
- gameplay rendering depends on `GameVM`, not on independently managed prompt and event fields
- other screens may remain unchanged until the first save/load slice is working

This ADR intentionally limits scope. It is not permission to keep ad hoc gameplay state forever; it is permission to phase the refactor to protect the first vertical slice.

## Consequences

Positive:

- solves the highest-risk UI coherence problem first
- keeps the refactor narrow enough to ship incrementally
- creates a reusable pattern for later screens

Costs:

- temporary coexistence of reducer-driven gameplay and older non-gameplay screens
- extra adapter code during the transition period

## Migration Impact

- no artifact migration directly
- tests for gameplay rendering must shift from field mutation behavior to reducer contracts

## Rollback Notes

Rollback is possible before other screens adopt reducer patterns, but doing so would restore the gameplay atomicity problem the roadmap explicitly forbids.

## Code Paths Impacted

- future `internal/session/gamevm.go`
- future `internal/session/reducer.go`
- `internal/tui/game.go`

## Tests Impacted

- future `internal/session/gamevm_test.go`
- `internal/tui/game_test.go`
- `internal/docspec/docs_contract_test.go`
