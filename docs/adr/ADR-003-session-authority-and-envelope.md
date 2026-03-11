# ADR-003: Session Authority and a Single Ordered Authority Envelope

Status: accepted
Date: 2026-03-11

## Context

The current runtime exposes event state and action-request state on separate channels. That split forces the TUI to merge authority information heuristically and creates risk of stale prompt display, stale legal actions, and UI states that do not match the latest authoritative sequence.

The roadmap requires prompt visibility and action availability to derive from a single authoritative message stream or equivalent reducer contract.

## Options Considered

- Keep separate event and prompt channels and add more ordering hacks in the TUI.
- Move authority ordering into the TUI model.
- Make `Session Authority` emit one ordered authority envelope that contains snapshot, prompt, notices, and sequence metadata.

## Decision

Ante adopts one ordered authority envelope emitted by `Session Authority`.

Binding rules:

- one monotonic sequence number orders all reducer-facing updates
- if a human prompt exists, it belongs to the same envelope sequence as the snapshot it describes
- stale prompts are rejected by sequence, not guessed around in the renderer
- the TUI consumes this stream and reduces it into `GameVM`
- the TUI does not independently decide when legal actions are valid
- the same ordered envelope also carries session-control prompts that are not engine poker actions, including the between-hands readiness barrier
- after every completed hand, `Session Authority` may pause in a between-hands synchronization state and require explicit readiness before the next hand begins

In the sandbox, `Session Authority` is a single local process. This is acceptable only because sandbox chips are ephemeral. This ADR does not authorize a future host-authoritative trust model for persistent-value tables.

## Consequences

Positive:

- prompt, log, and legal actions can change atomically
- reducer tests can prove ordering behavior clearly
- later network protocol work has a clearer authority contract to preserve
- hand-boundary save/load and future P2P settlement both have a formal sync barrier instead of relying on auto-advance timing

Costs:

- session and gameplay screen interfaces must change
- some existing tests around channel behavior will need updates

## Migration Impact

- no durable artifact migration is required directly by this ADR
- save/load and transcript code should target the new envelope contract to avoid double work

## Rollback Notes

Rolling back to split prompt and event channels would reintroduce state-merging ambiguity and block the reducer architecture.

## Code Paths Impacted

- `internal/session/session.go`
- future `internal/session/envelope.go`
- `internal/tui/game.go`

## Tests Impacted

- future `internal/session/envelope_test.go`
- future `internal/session/gamevm_test.go`
- `internal/tui/game_test.go`
- `internal/docspec/docs_contract_test.go`
