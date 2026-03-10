# Ante

Ante is a terminal Texas Hold'em poker game built in Go. It features a custom internal engine, challenging AI opponents, robust simulation tooling, and a sleek Bubble Tea Terminal User Interface (TUI).

## Screenshots

### Main Menu
![Main Menu UI](main-menu.png)

### Gameplay
![Gameplay UI](gameplay.png)

## Overview

The goal of Ante is to provide a complete and engaging poker experience entirely within your terminal. Whether you want to play a quick cash game or test your skills in a tournament, Ante offers a rich environment with distinct AI personalities.

## Features

- Texas Hold'em Gameplay: Experience a polished terminal interface with clear and intuitive readouts.
- Multiple Game Modes: Select between tournament, cash game, and heads-up duel formats.
- AI Opponents: Play against bots featuring distinct characters and adjustable difficulty levels.
- Analytics: Review detailed hand history and track your player statistics over time.
- State Persistence: Save your configuration, progress, and session data locally automatically.
- Engine Stability: Run the included simulation entrypoint to stress-test the custom core engine.

## Project Structure

- `cmd/ante`: Main TUI executable application.
- `cmd/sim`: Simulation and engine smoke-test runner executable.
- `internal/engine`: Core poker rules, betting logic, pot mechanics, showdowns, tournaments, and overall table representation.
- `internal/ai`: Bot behavioral models, heuristics, and decision evaluation algorithms.
- `internal/session`: Central orchestration layer connecting the engine, AI systems, and UI components.
- `internal/tui`: All Bubble Tea interface implementations, views, and specific screen rendering logic.
- `internal/storage`: Local persistence layer handling configurations, game saves, and statistics read/write operations.

## Requirements

- Go 1.26 or newer

## Quick Start

Run the Ante TUI application directly:

```bash
go run ./cmd/ante
```

Run the engine simulator (e.g., simulating 1000 hands):

```bash
go run ./cmd/sim -hands 1000
```

## Build Instructions

Compile the TUI application to your local `bin` directory:

```bash
go build -o ./bin/ante ./cmd/ante
```

Compile the simulator runner:

```bash
go build -o ./bin/sim ./cmd/sim
```

## Test

Run the automated test suite without cached results:

```bash
go test ./... -count=1
```

## Controls

The interface relies on common keyboard navigation:

- `Up` / `Down`: Navigate menu items.
- `Enter`: Confirm selection.
- `Esc`: Go back to the previous screen or open the pause menu.
- `Q`: Fold your hand or quit the application (context-dependent).
- `W`: Check the current betting round.
- `E`: Call the current bet.
- `T`: Raise or make a new bet.
- `A`: Push all-in.
- `0-9`: Directly enter specific bet amounts.

## Local Data Storage

Ante automatically manages and persists local user data:

- Main configuration file: `~/.ante/config.json`
- Save files and statistics are managed by the storage layer in the same application directory.

## Quality Assurance

The project currently includes automated tests across the core engine, session, AI, and TUI layers. The repository is also configured with GitHub Actions to run build and test checks on every push and pull request.

## Roadmap

- Improve release automation pipelines for downloadable binaries.
- Add comprehensive benchmark suites and profiling workflows to further optimize engine performance.

## License

This project is licensed and released under the terms of the MIT License. See `LICENSE` for details.
