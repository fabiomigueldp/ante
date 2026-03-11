# ADR-005: TimeAnchorProvider Is Mandatory for Persisted Timestamps

Status: accepted
Date: 2026-03-11

## Context

The current code calls `time.Now()` directly for persisted data such as hand record timestamps. The roadmap treats `TimeAnchorProvider` as mandatory infrastructure because later refill, settlement, and verification flows need provenance-aware timestamps.

## Options Considered

- Continue calling `time.Now()` in business logic and refactor later.
- Wrap time reads in helper functions without provenance or explicit errors.
- Introduce `TimeAnchorProvider` now and route persisted timestamps through it.

## Decision

Ante adopts `TimeAnchorProvider` as the required boundary for persisted or trust-sensitive timestamps.

Binding rules:

- code that creates persisted artifacts must request a time anchor from the provider
- the resulting artifact stores timestamp value and provider provenance
- direct `time.Now()` calls are not allowed in business logic that writes durable artifacts
- sandbox may start with `LocalTimeAnchorProvider` using the local clock and provenance `local_clock`

## Consequences

Positive:

- timestamp creation becomes testable and explicit
- future trust-sensitive flows can swap providers without redesigning call sites
- provenance is captured early rather than retrofitted later

Costs:

- more dependency injection in storage and session code
- some simple runtime paths become slightly more explicit

## Migration Impact

- new artifact schemas need timestamp provenance fields
- old artifacts without provenance are migration inputs and may receive default legacy provenance where appropriate

## Rollback Notes

Rolling back would reintroduce hidden clock dependencies and make later free-play timing rules harder to secure.

## Code Paths Impacted

- `internal/storage`
- `internal/session`

## Tests Impacted

- future `internal/storage/time_anchor_test.go`
- future migration tests for legacy timestamps
- `internal/docspec/docs_contract_test.go`
