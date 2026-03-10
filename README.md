# Ante

Ante is a terminal poker game built in Go with a custom engine, AI opponents, simulation tooling, and a Bubble Tea TUI.

## Features

- Play Texas Hold'em in a polished terminal interface
- Choose between tournament, cash game, and heads-up duel modes
- Face AI opponents with distinct characters and difficulty settings
- Review hand history and player statistics
- Save configuration and session data locally
- Run a simulation entrypoint to stress-test engine stability

## Project Structure

- `cmd/ante`: main TUI application
- `cmd/sim`: simulation and engine smoke-test runner
- `internal/engine`: poker rules, betting, pots, showdowns, tournaments, and table logic
- `internal/ai`: bot behavior and decision evaluation
- `internal/session`: orchestration layer between engine, AI, and UI
- `internal/tui`: Bubble Tea interface and screens
- `internal/storage`: local persistence for config, saves, and stats

## Requirements

- Go 1.26 or newer

## Quick Start

```bash
go run ./cmd/ante
```

Run the simulator:

```bash
go run ./cmd/sim -hands 1000
```

## Build

Build the TUI application:

```bash
go build -o ./bin/ante ./cmd/ante
```

Build the simulator:

```bash
go build -o ./bin/sim ./cmd/sim
```

## Test

```bash
go test ./... -count=1
```

## Controls

- `Up` / `Down`: navigate menus
- `Enter`: confirm selections
- `Esc`: go back or open the pause menu
- `Q`: fold or quit, depending on context
- `W`: check
- `E`: call
- `T`: raise or bet
- `A`: all-in
- `0-9`: enter bet amounts

## Local Data

Ante stores local user data under:

- `~/.ante/config.json`

Additional save and stats files are managed by the storage layer in the same application directory.

## Quality Status

The project currently includes automated tests across the core engine, session, AI, and TUI layers. The repository is also configured with GitHub Actions to run build and test checks on every push and pull request.

## Roadmap

- Improve release automation for downloadable binaries
- Expand documentation with gameplay screenshots or terminal captures
- Add benchmark and profiling workflows for engine performance

## License

This project is released under the MIT License. See `LICENSE` for details.
