# ADR-006: Canonical Encoding and Stable Artifact Identifiers

Status: accepted
Date: 2026-03-11

## Context

Transcript checkpoints and future signed states require bytes that are deterministic across runs and machines. Raw JSON is not stable enough for that role, and filename-only relationships are too fragile for migrations and tooling.

The roadmap also requires stable linkage between transcript identifiers, snapshot identifiers, and session identifiers.

## Options Considered

- Use raw JSON for hashes and rely on filenames for linkage.
- Keep gob encodings as the durable canonical format.
- Adopt canonical binary encoding for hashed or signed payloads and define stable identifier formats with deterministic relationships.

## Decision

Ante adopts canonical deterministic encoding for every payload that is hashed now or signed later.

Binding rules:

- raw JSON bytes must never be used as canonical hash or signature inputs
- gob is not the canonical encoding for replay-critical or signature-critical artifacts
- checkpoint hashing uses canonical bytes only
- future signed states must use the same rule

Accepted stable identifier formats:

- `session_id`: `ses_<32 lowercase hex>`
- `transcript_id`: `trn_<same 32 lowercase hex base token>`
- `chunk_id`: `tch_<same 32 lowercase hex base token>_<6-digit chunk index>`
- `checkpoint_id`: `ckp_<same 32 lowercase hex base token>_<6-digit hand index>`
- `snapshot_id`: `snp_<same 32 lowercase hex base token>_<6-digit hand index>_<9-digit seq>`

The shared base token ties related artifacts to one session without relying on directory names alone.

## Consequences

Positive:

- replay-critical integrity checks become deterministic
- future signatures can build on the same encoding rules
- tooling can reason about related artifacts from identifiers directly

Costs:

- encoding helpers and tests must be introduced before cryptographic features exist
- legacy gob artifacts require migration or explicit rejection

## Migration Impact

- old save and stat formats remain compatibility inputs only
- new replay-critical artifacts must adopt the stable ID model immediately

## Rollback Notes

Rolling back to raw JSON or filename-only linkage would invalidate the roadmap's checkpoint and signature requirements.

## Code Paths Impacted

- `internal/storage`
- `internal/session`

## Tests Impacted

- future `internal/storage/canonical_test.go`
- future `internal/session/transcript_test.go`
- `internal/docspec/docs_contract_test.go`
