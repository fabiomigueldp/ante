# Contributing

Thank you for considering a contribution to Ante.

## Development Setup

1. Install Go 1.26 or newer.
2. Clone the repository.
3. Run `go test ./... -count=1` before submitting changes.

## Guidelines

- Keep changes focused and well-scoped.
- Preserve existing architecture boundaries between `engine`, `session`, `ai`, `tui`, and `storage`.
- Add or update tests when changing behavior.
- Use clear commit messages and open pull requests with a concise summary.

## Pull Requests

Please include:

- a short description of the change
- any relevant reasoning or tradeoffs
- test evidence when behavior changes
