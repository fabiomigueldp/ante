# Roadmap

Ante is currently a local terminal poker application written in Go: a deterministic engine, Bubble Tea TUI, AI opponents, and a partial persistence layer. The core rules work, but the product surface is still split between finished systems and visible scaffolding. The project now needs two things at once: close the sandbox as a complete, coherent product, and establish the technical foundations for a protocol-first free-table multiplayer architecture.

This roadmap therefore has two equally important commitments. First, the sandbox must become a finished, independent product with truthful UX, durable artifacts, and a premium table experience. Second, the multiplayer plan must be specified precisely enough that identity, networking, transcripts, and economy rules do not drift into improvised host-authoritative behavior later.

The multiplayer scope in this roadmap is strictly free-play heads-up cash. Paid tables are out of scope, but the architecture must still preserve a hard separation between sandbox state, free-play persistent balances, and any future paid economy so that no later expansion invalidates the core trust model.

## Constraints

- The engine remains deterministic. Multiplayer must be built around the existing deterministic rules engine, not around duplicated rules on each client.
- Single-player sandbox quality matters. Tracks 0-3 are not throwaway work; they define artifacts, transcripts, UI contracts, and operational boundaries reused by later tracks.
- The sandbox is a finished product in its own right. It must function with zero dependency on identity, networking, or multiplayer economy modules.
- Sandbox chips are ephemeral. They are never convertible into free-play persistent balances, paid balances, or any off-table value.
- Free and paid economies are completely isolated. No conversion, no mixed tables, and no bridging. Paid tables are out of scope for this roadmap, but the isolation rule must still shape the architecture in every track.
- Multiplayer scope is strictly free tables.
- Bots are exclusive to the sandbox. The P2P protocol does not integrate, detect, or block bots, but bots are not part of the multiplayer product design.
- Heads-up cash is the sole format for the P2P protocol. Tournaments, 6-max, and 9-max remain sandbox-only.
- No leaderboard is in scope before or during the free-table alpha.
- No gifts, no off-table transfers, and no wallet-like value movement are in scope in any track in this roadmap.
- No spectators are in scope in any track in this roadmap.
- No chat is in scope in any track in this roadmap.
- No marketing-driven positioning changes should be made before the free-table multiplayer loop is publicly usable in alpha.
- No "provably fair" or equivalent claim may be used before Track 7 is complete and independently verifiable artifacts exist.
- `crypto/rand` is mandatory in every protocol and cryptographic code path. `math/rand` is acceptable only in AI decision-making, audio synthesis, and the simulation runner.
- All signed states use canonical deterministic encoding. Nothing is signed over raw JSON.
- Every session must be replayable and auditable from its transcript alone.
- Disputes are symmetric: either side can submit the latest mutually signed state.
- Host-authoritative game logic is forbidden for any table where chips have persistent value. It may exist only as a development harness during network bringup and must be explicitly marked as such.
- `TimeAnchorProvider` is required before the product may be considered stable. It is not an optional enhancement.
- Refill logic must use a lazy-applied model. The application computes refill entitlement when a relevant operation occurs and persists the result. It must not depend on background timers, daemons, or periodic local tasks.
- All persistence added after Track 1 must go through `ArtifactStore` abstractions. TUI screens and session code must not write raw gob files directly.
- TUI state changes that affect prompts, logs, and legal actions must be atomic from the user's point of view. Visible state cannot be composed from independent unsynchronized streams.
- Reconnect and resume flows must not duplicate authority, duplicate transcript segments, or duplicate economy effects.

## Glossary

- `ArtifactStore`: the typed persistence boundary for local durable data such as sandbox saves, sandbox transcripts, stats, identity artifacts, time anchors, free-play balance state, and migration metadata.
- `Snapshot`: an immutable serialized view of table or session state captured at a known sequencing boundary.
- `TranscriptRecord`: the append-only authoritative record of gameplay events, metadata, signatures, and checkpoints required for replay, debugging, synchronization, and dispute resolution.
- `Session Authority`: the sequencing role that orders legal actions, prompt issuance, checkpointing, and transcript append. In the sandbox it may be a single local process; in persistent-value multiplayer it is a liveness coordinator, not a sole trust root.
- `GameVM`: the TUI-facing view-model reduced from authoritative snapshots and prompt messages. Rendering depends on `GameVM`, not on ad hoc reads across session internals.
- `Reducer`: the pure state transition function that applies a typed UI or protocol message to the current `GameVM` and returns the next `GameVM`.
- `Prompt Envelope`: the typed message that describes whose turn it is, what legal actions exist, and which authoritative sequence number the prompt belongs to.
- `Canonical State Encoding`: the deterministic binary encoding used for all signed states, signed envelopes, and replay-critical artifacts.
- `Signed State`: a canonical payload whose bytes are signed and verified directly, without lossy or ambiguous re-encoding.
- `Identity Bundle`: the locally stored cryptographic key material, encrypted seed, and signed metadata used to authenticate a device or player to multiplayer services.
- `IdentityGenesis`: the first signed identity artifact derived from the root seed, including the identity public key, wallet address, timestamp, proof-of-work fields, and signature.
- `SessionKeyDelegation`: the signed delegation from a long-lived identity key to an ephemeral per-session key, scoped to a table identifier and validity window.
- `ZKCredential`: a bounded extension object carrying an optional zero-knowledge credential claim for future policy checks. It is accepted structurally before it is processed semantically.
- `TableOffer`: the signed protobuf message advertising a free table on the network.
- `JoinRequest`: the signed protobuf message requesting entry into a specific free table.
- `BalanceChain`: the hash-linked, signed ledger of free-play chip state for one identity.
- `Dispute Bundle`: the transcript slice, mutually signed checkpoints, and related signed states used to prove the latest valid balance-affecting session result.
- `TimeAnchorProvider`: the boundary for obtaining time values used by persisted or trust-sensitive features. It must expose provenance and error handling instead of allowing direct uncontrolled `time.Now()` use in business logic.
- `Lazy-applied refill`: a refill policy that stores the last applied anchor, computes any refill delta only when the player loads, joins, or requests a rebuy, applies that delta once, and persists the new anchor.
- `HeadAnnouncement`: the signed gossip message advertising a `BalanceChain` head hash and sequence for equivocation detection.
- `Peer Score`: the network reputation value used to penalize malformed, abusive, or low-quality peer behavior.
- `Free Table Tier`: the standardized policy bucket used for free heads-up tables, including blinds and allowed buy-in range.
- `Economy Isolation`: the rule that sandbox chips, free-play persistent chips, and any future paid economy remain fully separated in storage, protocol, and user experience.

## Track 0. Scope Lock and Architecture Documentation

### Objective

Create the shared technical vocabulary and design documents required to keep the architecture coherent while the project moves from a local sandbox to a protocol-first free-table multiplayer system.

### Dependencies

None. This is the baseline track.

### Deliverables

- [ ] Create `docs/architecture.md`. Required sections:
  - project scope, product boundaries, and non-goals
  - explicit separation between sandbox, free-play multiplayer, and any future paid economy
  - package map for `cmd/*`, `internal/engine`, `internal/session`, `internal/tui`, `internal/storage`, and future modules introduced by this roadmap
  - authoritative runtime data flow: user input -> session authority -> transcript -> snapshot -> reducer -> renderer
  - `ArtifactStore` responsibilities, artifact types, namespacing rules, and ownership boundaries
  - transcript and snapshot lifecycle, including replay, checkpointing, and recovery expectations
  - `TimeAnchorProvider` contract, provenance model, fallback behavior, and error surfaces
  - free-table multiplayer layering and the distinction between local sandbox runtime and P2P runtime
  - `BalanceChain` overview, identity boundaries, and networking stack overview
  - failure domains, crash recovery boundaries, migration expectations, and trust-root assumptions
- [ ] Create `docs/threat_model.md`. Required sections:
  - assets and trust boundaries
  - actor model: local player, remote peer, relay, bootstrap node, storage attacker, clock manipulator, protocol attacker
  - attack surfaces in persistence, transcript handling, identity, networking, balance verification, and time anchoring
  - replay, tampering, impersonation, equivocation, clock-skew abuse, and lock-doubling scenarios
  - sandbox-only bot policy and why bots are excluded from multiplayer scope
  - host-authoritative bringup harness risks and explicit prohibition for persistent-value tables
  - incident response expectations and unresolved risks
- [ ] Create `docs/decision_register.md`. Required sections:
  - ADR template definition
  - status model (`proposed`, `accepted`, `superseded`, `rejected`)
  - per-decision fields: id, date, context, options considered, decision, consequences, migration impact, rollback notes
  - index of accepted decisions
  - cross-reference rules for code paths and tests impacted by each decision
  - reserved ADR entries for canonical state encoding, heads-up-only P2P scope, network stack choice, identity key hierarchy, and `BalanceChain` policy
- [ ] Create `docs/glossary.md`. Required sections:
  - canonical definitions for the terms used in this roadmap
  - naming rules for artifacts, transcripts, prompts, identities, tables, locks, settlements, and refills
  - disallowed ambiguous terminology and preferred replacements
  - definitions for economy isolation, signed states, dispute bundles, and free-table tiers
- [ ] Record the scope guardrails from this roadmap in the documentation set so later tracks cannot quietly drift into leaderboard, gifting, off-table transfer, spectator, chat, or mixed-economy work.

### Acceptance Criteria

- The `docs/` directory exists and contains the four documents above.
- Each document contains the required sections listed in this roadmap, but the roadmap itself does not inline their full contents.
- Track 1 work does not begin until the terminology for `ArtifactStore`, `TranscriptRecord`, `Session Authority`, `GameVM`, `BalanceChain`, and `TimeAnchorProvider` is documented.

## Track 1. Core Architecture and Runtime Boundaries

### Objective

Replace ad hoc storage, mixed UI/session state, and implicit timing assumptions with explicit boundaries that later sandbox and multiplayer tracks can safely build upon.

### Dependencies

Track 0 must be complete.

### Deliverables

- [ ] Introduce `ArtifactStore` in `internal/storage` as the typed persistence boundary for:
  - sandbox snapshots and resumable saves
  - sandbox transcripts and replay records
  - sandbox stats and session summaries
  - identity artifacts
  - time anchors and future refill state
  - migration metadata and artifact manifests
- [ ] Add artifact versioning and migration support so current `internal/storage/save.go` gob saves, `internal/storage/stats.go` gob stats, and `internal/storage/config.go` JSON configuration can be read, migrated, or explicitly rejected with user-visible errors.
- [ ] Refactor `internal/session/session.go` so prompt state is no longer exposed through unrelated outbound channels. Replace the current `Events` plus `ActionReq` split with an ordered typed envelope carrying:
  - authoritative sequence number
  - hand and table identity
  - snapshot payload
  - prompt payload when action is required
  - structured UI-facing notices and errors
- [ ] Extract a TUI-facing reducer such as `GameVM` so `internal/tui/game.go` no longer mixes transport, mutation, and rendering logic in `handleSessionEvent`, `handleActionReq`, and `renderActionBar`.
- [ ] Enforce renderer boundaries so the TUI composes from reducer state rather than reading loosely coupled fields like `needsAction`, `lastAction`, and transient message state independently.
- [ ] Introduce deterministic transcript chunking and checkpoint hashing so replay-critical artifacts are hash-linked before any multiplayer or dispute layer exists.
- [ ] Define stable linkage between transcript identifiers, snapshot identifiers, and session identifiers so replay and migration tools can reason about artifacts deterministically.
- [ ] Introduce `TimeAnchorProvider` as a first-class boundary and route persisted trust-sensitive timestamps through it. This includes session summary timestamps, transcript timestamps, save timestamps, and all later refill logic.
- [ ] Define package ownership boundaries for future modules. At minimum:
  - `internal/engine` owns pure poker rules and deterministic state transitions
  - `internal/session` owns local authority orchestration and sandbox session accumulators
  - `internal/tui` owns rendering and input only
  - `internal/storage` owns artifacts and migrations only
  - future identity, networking, and economy packages must not back-reference TUI concerns
- [ ] Add migration and sequencing tests covering:
  - legacy save/stat/config migration into artifact-backed storage
  - prompt and event ordering with explicit sequence numbers
  - stale prompt rejection in the reducer
  - checkpoint hash continuity
  - `TimeAnchorProvider` injection in timestamped artifacts

### Acceptance Criteria

- Local sandbox play still works after the boundary refactor.
- No new feature work in later tracks writes directly to raw gob save files from the TUI or session layers.
- Prompt visibility and action availability derive from a single authoritative message stream or equivalent reducer contract.
- Transcript chunks, snapshots, and session identifiers are linked deterministically enough for later replay and signature work.
- `TimeAnchorProvider` exists, is injected where persisted timestamps are created, and is treated as mandatory infrastructure rather than optional polish.

## Track 2. Sandbox Completion and Product Integrity

### Objective

Close the currently visible single-player product gaps using the architecture introduced in Track 1, while preserving deterministic artifacts, transcript quality, and the premium TUI direction.

### Dependencies

Track 1 is required. Every item below must use `ArtifactStore` instead of bespoke persistence and must respect the reducer and sequencing boundaries introduced in Track 1. The internal implementation order in this track is fixed and must remain as listed.

### Deliverables

#### 2.1 Save / Load / Continue

##### What exists today

- `internal/storage/save.go` defines `SaveSlot`, `TableSaveData`, `PlayerSaveData`, `GameConfig`, and `HandRecordSave`.
- `storage.SaveGame()`, `storage.LoadGame()`, `storage.DeleteSave()`, and `storage.ListSaves()` serialize five slot files using gob.
- `internal/tui/loadgame.go` lists saved slots and supports deletion.
- The pause menu in `internal/tui/game.go` exposes a save action bound to `S`.

##### What is missing

- Save is not wired. In `internal/tui/game.go`, `handlePauseKey()` still takes the placeholder path that calls `setMessage("Game saved!")` without constructing a real save artifact or calling persistence.
- Load does not resume. In `internal/tui/loadgame.go`, pressing `Enter` on a populated slot hits a `TODO` path and returns to the menu.
- The current `SaveSlot` schema is not sufficient for reliable mid-hand resume. It omits authoritative prompt state, current street sequencing details, action seat, side-pot state, deck order, and other data needed to restore an in-progress hand exactly.
- The current file format is feature-local. It does not participate in artifact versioning, transcript linkage, or migration.

##### Completion requires

1. Replace direct feature ownership of gob saves with `ArtifactStore` operations for sandbox session snapshot artifacts and save-slot metadata under `~/.ante/sandbox/saves/`.
2. Add a `buildSaveArtifact()` or equivalent on `Session` that captures the minimum resumable state. If only hand-boundary resume is supported initially, that restriction must be explicit in the UI and artifact schema.
3. Extend the save schema if mid-hand resume is supported. Required state includes current street, action seat, player bets, side pots, current board, deck order or equivalent deterministic recovery material, legal prompt state, and bot or authority sequencing state.
4. Implement `resumeSession()` or equivalent reconstruction logic that rebuilds `session.Session`, the underlying table state, bot instances with correct seeds, blind level state, and transcript linkage from the saved artifact.
5. Wire the pause menu save action in `internal/tui/game.go` and the load flow in `internal/tui/loadgame.go` through the new artifact-backed persistence path.
6. Add migration or compatibility handling for existing gob saves in `internal/storage/save.go`.

##### Deliverables

- [ ] Replace raw save-slot persistence in `internal/storage/save.go` with `ArtifactStore` save snapshot operations and migration support for legacy gob files.
- [ ] Implement session snapshot construction from `internal/session/session.go` without leaking renderer concerns into storage.
- [ ] Implement session reconstruction from saved artifacts and wire the result to `ScreenGame` via `internal/tui/app.go` and `internal/tui/loadgame.go`.
- [ ] Make the pause menu in `internal/tui/game.go` perform a real save and show a real result state.
- [ ] Add round-trip tests for save creation, artifact load, and resumed gameplay integrity.

##### Acceptance Criteria

- Saving from the pause menu persists a real artifact and updates the load screen with correct slot metadata.
- Loading a saved slot transitions directly back into `ScreenGame` with a reconstructed `session.Session`.
- Unsupported save scenarios fail clearly and deterministically; they do not claim success.
- Existing gob saves are either migrated or rejected with an explicit compatibility message.

#### 2.2 Statistics Persistence

##### What exists today

- `internal/storage/stats.go` defines `SessionStats` and `StatsStore` with aggregate queries such as total sessions, total hands played, tournament wins, total profit, average finish, win rate, best hand, and recent sessions.
- `storage.LoadStats()` and `storage.SaveStats()` persist stats using gob.
- `internal/tui/statsview.go` reads the persisted store and renders the statistics screen.
- `SessionStats` already includes fields for richer per-session detail such as `FlopsSeen`, `ShowdownsWon`, `AllInsWon`, `BiggestPot`, `LargestWin`, `LongestStreak`, and `BestHand`.

##### What is missing

- No automatic recording exists. `emitSessionEnd()` in `internal/session/session.go` does not construct `SessionStats` and does not persist anything.
- The session loop does not currently accumulate all of the per-session counters described by `SessionStats`.
- The existing stats persistence is isolated from the future artifact model and cannot be cleanly correlated with transcripts or save artifacts.

##### Completion requires

1. Add a session-level accumulator to `internal/session/session.go` that records hands won, flops seen, showdowns seen and won, all-ins seen and won, biggest pot, largest win, streaks, and best hand.
2. Update the accumulator as hands progress rather than trying to reconstruct everything at session end.
3. At session end, build `SessionStats`, persist it through `ArtifactStore`, and preserve a link to related transcript and session identifiers.
4. Keep `internal/tui/statsview.go` reading through storage abstractions rather than directly depending on runtime session internals.
5. Reuse the same accumulator in the results screen so Track 2.5 does not recalculate from scratch.

##### Deliverables

- [ ] Add a session metrics accumulator to `internal/session/session.go` and feed it during live play.
- [ ] Persist session statistics via `ArtifactStore` instead of raw gob-only feature code.
- [ ] Update `internal/tui/statsview.go` to load artifact-backed stats and handle migrated historical data.
- [ ] Add tests for stat accumulation, persistence, and retrieval.

##### Acceptance Criteria

- Completing a session writes one durable stats artifact without manual intervention.
- The statistics screen reflects newly completed sessions immediately after returning to it.
- At least the fields already present in `SessionStats` are populated from real session data rather than remaining mostly zeroed placeholders.

#### 2.3 Hand History Browser

##### What exists today

- `internal/tui/historyview.go` opens from the menu and lists entries derived from `StatsStore`.
- Each row currently represents a session summary rather than individual hands.
- Pressing `Enter` in `HistoryViewModel.Update()` still hits a `TODO` path.
- `internal/engine/history.go` defines `HandRecord` and `SessionHistory`, and `internal/session/session.go` records `HandRecord` objects in `recordHand()`, but that history is memory-only.

##### What is missing

- There is no persistent transcript or hand record store. All hand-level detail disappears when the process exits.
- The browser cannot drill from session to hand.
- The current screen is wired to stats summaries rather than the real transcript structure required for replay and debugging.

##### Completion requires

1. Persist transcript or hand-record artifacts as durable sandbox session history under `~/.ante/sandbox/transcripts/` and `~/.ante/sandbox/history/`, either incrementally during play or at session end.
2. Keep the persisted structure rich enough for replay. `engine.HandRecord` and transcript artifacts must preserve blinds, dealer seat, player snapshots, board, actions, and event sequence.
3. Update `internal/tui/historyview.go` to support two navigation levels: session list and hand list.
4. Use artifact identifiers rather than transient slice indexes to address records.
5. Add retention and pruning rules so history growth remains bounded and deliberate.

##### Deliverables

- [ ] Add transcript or hand-record persistence through `ArtifactStore`, sourced from `internal/session/session.go` and `internal/engine/history.go`.
- [ ] Update `internal/tui/historyview.go` to browse sessions and then individual hands.
- [ ] Replace the `TODO` enter handler with working navigation into session details or replay selection.
- [ ] Add retention policy and migration tests for persisted history.

##### Acceptance Criteria

- Hand history survives process restart.
- Users can select a session and inspect the individual hands recorded within it.
- History entries are backed by durable transcript data, not just by session-level stat summaries.

#### 2.4 Hand Replay

##### What exists today

- `internal/tui/replay.go` defines `ReplayModel` and allows stepping forward and backward through the actions of a `HandRecord`.
- The replay screen already shows blinds, dealer seat, players, board area, action list, and a progress bar.
- Keyboard navigation is implemented for left and right stepping, home and end jumps, and escape to return.

##### What is missing

- Replay is not connected to the history browser. `internal/tui/app.go` does not pass a `HandRecord` into `ReplayModel` when `ScreenReplay` is activated.
- `ReplayModel.boardAtStep()` is heuristic and explicitly incomplete. It shows the full board too early and does not reveal streets based on authoritative boundaries.
- Hole cards and showdown reveals are not surfaced properly from persisted history.

##### Completion requires

1. Use transcript metadata or explicit street boundaries rather than action-count heuristics to determine when flop, turn, and river become visible.
2. Wire replay selection from `internal/tui/historyview.go` through `switchScreenMsg` into `internal/tui/app.go` and into `NewReplayModel()`.
3. Extend replay data so human hole cards and showdown-revealed cards can be displayed when the transcript supports them.
4. Keep replay deterministic against the stored transcript rather than reconstructing loosely from partial summaries.

##### Deliverables

- [ ] Fix `internal/tui/app.go` so `ScreenReplay` receives a real replay selection payload instead of the current dead `*session.Session` assertion path.
- [ ] Replace the approximate `boardAtStep()` logic in `internal/tui/replay.go` with transcript-based street progression.
- [ ] Surface recorded hole cards and showdown reveals where the transcript allows them.
- [ ] Add replay tests using real persisted `HandRecord` or transcript artifacts.

##### Acceptance Criteria

- Selecting a hand from the history browser opens a working replay.
- Board cards appear street by street at the correct points in the replay timeline.
- Replay output is reproducible from durable transcript data and not dependent on ad hoc heuristics.

#### 2.5 Results Screen

##### What exists today

- `internal/tui/results.go` renders a results screen with the main result message, hands played, and final standings sorted by stack.
- `ResultsModel.renderStats()` already attempts to summarize session history.
- The code acknowledges the need for richer history-derived values, but the current implementation does not actually populate them.

##### What is missing

- Biggest pot is not populated. `internal/tui/results.go` contains a comment noting that amounts are not easily extracted from the event interface in the current implementation.
- No best hand, streak, or memorable session highlights are displayed.
- No comparison against prior results exists.
- The results screen does not yet consume the session accumulator required by Track 2.2.

##### Completion requires

1. Use the same session accumulator built for stats persistence rather than re-scanning ad hoc runtime structures.
2. Populate biggest pot, best hand, largest win, longest streak, and other stable highlights from that accumulator.
3. Where historical stats exist, add optional comparison against historical averages without making the results screen depend on the stats screen directly.
4. Preserve support for both tournament and cash-style summaries.

##### Deliverables

- [ ] Feed `internal/tui/results.go` from the Track 2.2 session accumulator.
- [ ] Populate biggest pot, best hand, largest win, and streak values in the results UI.
- [ ] Add optional comparison against historical averages when artifact-backed stats are available.
- [ ] Add tests for results rendering based on completed session metrics.

##### Acceptance Criteria

- The results screen shows real highlights, not placeholders.
- Biggest pot and similar values are computed from authoritative data without fragile interface inspection hacks.
- Tournament and cash-game results both display coherent summaries.

#### 2.6 Cash Game Continuity

##### What exists today

- `internal/engine/cashgame.go` already supports `OfferRebuy()`, `ReplacePlayer()`, and `CashOut()`.
- `internal/tui/help.go` describes cash games as an open-ended mode with fixed blinds and voluntary exit.
- The session loop can run cash games, but it still behaves like a bounded single-session lifecycle.

##### What is missing

- There is no open-ended session flow. The current runtime ends when the human busts or when the table reaches a generic end condition.
- There is no voluntary "Stand Up" or "Leave Table" action that cashes out cleanly.
- Bot replacement or rebuy behavior is not wired to maintain table continuity.
- Cash-game profit is not carried into results and stats in a first-class way.

##### Completion requires

1. Extend `internal/session/session.go` so cash games do not terminate simply because one stack hit zero.
2. Offer human rebuys at appropriate points and decide whether busted bots are replaced or automatically re-bought to maintain table size.
3. Add an explicit leave-table path that calls `CashOut()` and transitions through results cleanly.
4. Surface cash-game profit in both `internal/tui/results.go` and persisted stats.
5. Keep this mode sandbox-only. It prepares lifecycle expectations and UX, but it must not create an off-table balance system or reuse future `BalanceChain` state.

##### Deliverables

- [ ] Update `internal/session/session.go` so cash games continue until the player leaves rather than ending on the first bust condition.
- [ ] Wire rebuy and bot continuity behavior through `internal/engine/cashgame.go`.
- [ ] Add a leave-table flow and expose it clearly in the TUI.
- [ ] Persist cash-game profit into stats and show it in results.
- [ ] Add coverage for rebuy, cash-out, and long-running cash-game sessions.

##### Acceptance Criteria

- Cash games match the help text: fixed-blind, open-ended play with voluntary exit.
- The player can bust, rebuy, continue, and cash out without corrupting stats or history.
- No off-table wallet, gifting, or transferable balance is introduced.

#### 2.7 Setup Screen Completeness

##### What exists today

- `internal/tui/setup.go` exposes mode, seats, difficulty, starting stack, and player name.
- `session.Config` already supports additional fields such as `BlindSpeed`, `Seed`, `CashGameBuyIn`, and `CashGameBlinds`.

##### What is missing

- Blind speed is not exposed in the setup UI.
- Cash-game blinds and buy-in are not configurable from the TUI.
- Deterministic seeds are supported in the backend but not exposed to advanced users.
- The setup flow does not yet model the configuration distinctions that later free-table policy screens will need.

##### Completion requires

1. Add blind-speed controls for tournament-style modes.
2. Add cash-game-specific fields for blinds and buy-in, conditionally displayed when cash mode is selected.
3. Add an advanced section or equivalent for deterministic seeds and other expert options.
4. Keep backend validation in `session.Config` authoritative so the UI cannot drift away from supported values.

##### Deliverables

- [ ] Extend `internal/tui/setup.go` to expose `BlindSpeed`, `CashGameBlinds`, `CashGameBuyIn`, and an advanced seed option.
- [ ] Add validation and mode-conditional rendering so unsupported field combinations cannot be selected.
- [ ] Keep setup state shape aligned with future table-policy configuration objects.
- [ ] Add tests covering the new field combinations and start-game payload generation.

##### Acceptance Criteria

- The setup screen exposes the capabilities already supported by `session.Config`.
- Invalid mode and field combinations are prevented before session creation.
- Advanced options are deliberate and do not clutter the default flow.

#### 2.8 Settings Effectiveness

##### What exists today

- `internal/tui/settings.go` supports editing and persisting player name, sound toggle, volume, pot odds toggle, animation speed, default mode and difficulty and seats and stack, and theme.
- Sound preview and pot-odds behavior already have live effects.

##### What is missing

- `AnimationSpeed` is persisted but has no gameplay effect. `advanceAnimation()` in `internal/tui/game.go` is effectively a stub.
- Theme values are persisted but do not alter the visual system in `internal/tui/theme.go`.
- The settings screen promises configurable presentation without a renderer architecture that actually responds to those settings.

##### Completion requires

1. Define what animation means in the TUI and route it through `tea.Tick` or equivalent scheduling rather than placeholder state.
2. Add theme tokens and palette selection so theme changes are reflected in rendered output.
3. Keep theme and animation implementation compatible with the reducer architecture from Track 1 and the seat-layout work in Track 2.9.
4. Treat these as sandbox-facing presentation features. They are valuable but do not override product integrity work.

##### Deliverables

- [ ] Implement meaningful animation control in `internal/tui/game.go` and related rendering paths.
- [ ] Refactor `internal/tui/theme.go` so styles derive from the active theme rather than fixed global color constants only.
- [ ] Ensure settings changes are visible during gameplay, not just persisted.
- [ ] Add tests for settings persistence and runtime effect.

##### Acceptance Criteria

- Changing animation speed has an observable effect, including a true off mode.
- Changing theme changes the rendered visual system.
- Settings no longer contain inert options.

#### 2.9 Table Visual Clarity

##### What exists today

- `internal/tui/game.go` renders the table through `renderTable()`, `splitPlayers()`, `renderOpponentRow()`, `renderSeat()`, `renderBoardArea()`, `renderHumanArea()`, and `renderActionBar()`.
- Opponents are currently composed as multiline string blocks in a flat horizontal row.
- `internal/tui/theme.go` defines `SeatStyle()`, but opponents currently call it as `SeatStyle(true, false, ...)`, which gives all live opponents similar visual weight.
- The action area reuses `needsAction`, `lastAction`, and transient message state inside the same model.

##### What is missing

- Active player highlighting is structurally wrong. The current actor is not clearly distinguished from other non-folded players.
- The current multiline seat composition is fragile. Variable-height seat content in `renderSeat()` and string concatenation in `renderOpponentRow()` cause lateral association drift and poor scaling from 6-max to heads-up.
- Eliminated players do not have a stable premium representation. They collapse into low-information `(out)` markers instead of remaining visually anchored or being cleanly compacted.
- Prompt and log sequencing is not authoritative from the user's point of view. `waitForSession()`, `handleActionReq()`, and `renderActionBar()` can still yield stale-action-button behavior if state is not invalidated atomically.
- Invalid-action copy still reflects raw engine errors from `internal/engine/betting.go` rather than human-readable guidance.
- Street progression is minimal and heads-up layout remains a special case handled only by game mode labels, not by actual table geometry.

##### Completion requires

1. Build fixed-height seat cards and anchored seat geometry on top of the Track 1 reducer rather than raw multiline concatenation.
2. Render the current actor, recent action label, status pill, stack, and cards as one visual unit per seat.
3. Decide and implement the late-stage table policy: either compact the table as player count drops or retain muted ghost seats. In either case, eliminated players must remain visually deliberate.
4. Separate the event rail, prompt dock, and transient error banner so buttons and logs cannot visually disagree.
5. Translate invalid-action errors into player-facing guidance such as "Check is available here" or "Minimum raise is X".
6. Add dedicated heads-up geometry instead of reusing the same opponent row, and keep it compatible with the future heads-up-only P2P table layout.

##### Deliverables

- [ ] Replace fragile row concatenation in `internal/tui/game.go` with anchored seat layout and fixed-height seat cards.
- [ ] Pass the authoritative current actor into rendering and apply a distinct active-state style.
- [ ] Add per-seat recent-action labels and a clearer street progression indicator.
- [ ] Rework action and prompt rendering so log updates and legal controls change atomically.
- [ ] Add human-readable invalid-action mapping instead of exposing raw engine strings.
- [ ] Add render and sequencing tests for 2-max, 6-max, and 9-max layouts.

##### Acceptance Criteria

- A player's name, stack, status, and cards remain visually anchored as a single unit across state changes.
- Action buttons disappear or change exactly when authoritative prompt state changes.
- Eliminated-player rendering is visually deliberate rather than pooled or detached.
- Heads-up and short-handed layouts do not leave arbitrary empty scars on the table.

#### 2.10 Tournament UX

##### What exists today

- Tournament mode already works. Blinds increase, elimination works, and the session ends with ranked results.
- The header shows current blinds and blind increase events are displayed as transient messages.

##### What is missing

- Players cannot see upcoming blind levels or hands remaining until the next increase.
- The field state is opaque: players remaining, average stack, and relative standing are not surfaced.
- Milestone moments such as bubble, final table, and heads-up are not communicated as part of the UX.

##### Completion requires

1. Add a tournament HUD in the table header or side panel.
2. Surface current level, next level, hands until increase, players remaining, and average stack.
3. Add milestone messages for major tournament transitions.
4. Keep these values artifact-compatible so tournament sessions and transcripts can expose the same metadata later in tooling.

##### Deliverables

- [ ] Add tournament HUD information to the gameplay screen.
- [ ] Surface next blind level and countdown to increase.
- [ ] Show field status and milestone events.
- [ ] Add tests covering the tournament metadata calculations.

##### Acceptance Criteria

- Tournament progression is visible without leaving the game screen.
- Milestone moments are surfaced consistently.
- The HUD values match authoritative tournament state.

#### 2.11 Bot Reasoning Display

##### What exists today

- `internal/session/session.go` emits `bot_thinking` events with bot name and think time.
- Bot decisions in `internal/ai/bot.go` already contain a `Reason` string.
- The TUI currently shows only a generic thinking message.

##### What is missing

- The reasoning string is discarded in the current UX.
- Think-time variation exists but has little presentation value because it is not tied to a structured display surface.

##### Completion requires

1. Display a concise post-action reason string in the message rail or action log.
2. Optionally gate this behind a setting for users who prefer a cleaner competitive UI.
3. Keep this feature sandbox-only. It is valuable for personality and learning, but it is not part of the free-table multiplayer scope.

##### Deliverables

- [ ] Surface bot reasoning strings in the TUI after bot actions.
- [ ] Add an optional setting to suppress reasoning if needed.
- [ ] Add tests for reasoning propagation from AI decision to rendered message.

##### Acceptance Criteria

- Bot decisions can be displayed with short rationale text.
- The feature can be disabled cleanly if it proves noisy.

#### 2.12 AI Strategic Depth

##### What exists today

- `internal/ai/bot.go` already evaluates hand strength, draws, pot odds, and pressure.
- Profiles in `internal/ai/characters.go` already include many strategic traits such as `VPIP`, `PFR`, `Trap`, `DrawBias`, `HeroCallBias`, `ThreeBetBias`, and `PositionBias`.
- The current bot loop already handles tilt accumulation and decay.

##### What is missing

- Several profile traits are defined but either unused or only weakly represented in action selection.
- Pre-flop and post-flop logic are still too uniform.
- There is no meaningful board-texture model, multiway adjustment, stack-depth planning, or tournament-pressure logic.

##### Completion requires

1. Phase 1: use `VPIP` and `PFR` to gate pre-flop participation and opening ranges.
2. Phase 2: implement trapping, hero-calling, draw bias, and three-bet behavior from the profile model.
3. Phase 3: add board-texture and multiway awareness plus more varied sizing.
4. Phase 4: add stack-aware play, tournament pressure, and inter-street planning.
5. Keep this work sandbox-only unless and until a separate bot-in-multiplayer plan is approved.

##### Deliverables

- [ ] Implement phased strategic upgrades in `internal/ai/bot.go` using the existing profile fields in `internal/ai/characters.go`.
- [ ] Add targeted tests for pre-flop ranges, bias-driven decisions, board texture handling, and stack-aware logic.
- [ ] Preserve deterministic behavior under seeded runs for testability and replay.

##### Acceptance Criteria

- Existing profile fields materially influence decisions.
- Seeded bot behavior remains testable.
- AI upgrades improve the sandbox without expanding multiplayer scope.

#### 2.13 Test Coverage

##### What exists today

- Engine tests are already strong across betting, pots, showdowns, decks, tables, and tournaments.
- Session, AI, TUI, and storage packages have tests, but they concentrate on narrower unit behavior.

##### What is missing

- There are no artifact-backed save/load round-trip tests.
- There are no session resume tests from saved artifacts.
- There are no end-to-end stats recording tests.
- There are no durable transcript tests that connect history to replay.
- There are not enough reducer and sequencing tests around prompt state and TUI visibility.

##### Completion requires

1. Treat test work as a requirement attached to every Track 2 item above.
2. Add artifact migration and round-trip tests for saves, stats, transcripts, and replay.
3. Add sequencing tests for prompt invalidation, stale message rejection, and action-bar correctness.
4. Add golden or structural render tests for the table layout states introduced by Track 2.9.

##### Deliverables

- [ ] Add save/load artifact round-trip tests.
- [ ] Add session resume tests from persisted artifacts.
- [ ] Add stats persistence and results and highlight tests.
- [ ] Add transcript-to-history-to-replay tests.
- [ ] Add reducer and rendering tests for prompt, layout, and error-state behavior.

##### Acceptance Criteria

- No Track 2 item is considered complete without its corresponding tests.
- The main product flows from save, resume, stats, history, replay, and TUI prompt handling are covered by automated tests.

### Acceptance Criteria

- The current backlog from the existing roadmap is closed in the order above, with Track 2.13 treated as a cross-cutting requirement rather than a terminal phase.
- All user-visible promises currently made by the sandbox help text, setup screen, save/load menu, statistics view, history view, and replay surfaces are either implemented or explicitly removed.
- The TUI behaves as a coherent premium terminal product instead of a collection of partially wired screens.

## Track 3. Sandbox as Finished Product

### Objective

Freeze the sandbox as a finished, shippable product before any multiplayer track begins, with truthful UX, clear storage boundaries, and no hidden dependency on identity, networking, or persistent economy code.

### Dependencies

Track 2 must be complete.

### Deliverables

- [ ] Confirm all sandbox modes are mature and truthful:
  - tournament remains sandbox-only
  - cash game continuity remains sandbox-only
  - heads-up duel remains sandbox-only
  - no menu, help text, or settings screen implies unavailable multiplayer features
- [ ] Confirm sandbox chips are ephemeral and cannot bridge into free-play persistent balances.
- [ ] Namespace sandbox saves, stats, transcripts, and related history under `~/.ante/sandbox/` and route them through `ArtifactStore`.
- [ ] Confirm the sandbox functions with zero dependency on identity, networking, P2P discovery, relay infrastructure, or `BalanceChain` verification modules.
- [ ] Perform a final polish pass across copy, help text, defaults, screenshots, and menu labels so the sandbox no longer contains visible scaffolding or misleading claims.
- [ ] Lock a final sandbox acceptance test matrix covering save/load, replay, results, long cash sessions, tournament progression, settings effects, and TUI prompt integrity.

### Acceptance Criteria

- The sandbox can be built, launched, and completed without loading identity, networking, or persistent multiplayer economy modules.
- All sandbox data lives under the sandbox namespace and remains isolated from future free-play state.
- No sandbox chips or sandbox artifacts are reusable as multiplayer balance proofs.
- The sandbox remains a coherent product even if multiplayer work is delayed.

## Track 4. Cryptographic Identity and Signed-State Foundations

### Objective

Introduce the concrete identity, key hierarchy, signing model, and canonical encoding rules required for multiplayer authentication and later economy verification.

### Dependencies

Track 3 must be complete.

### Deliverables

- [ ] Generate root identity entropy using 256 bits from `crypto/rand`.
- [ ] Derive a 24-word BIP-39 mnemonic and 512-bit master seed from that entropy.
- [ ] Derive and persist the long-lived key hierarchy:
  - `Ed25519` keypair for libp2p peer identity
  - `Secp256k1` keypair for EVM wallet compatibility, derived now for stable identity and wallet mapping even though paid tables are out of scope
- [ ] Encrypt the seed at rest using `Argon2id` with `time=3`, `memory=64MB`, and `threads=4`, plus `AES-256-GCM` for authenticated encryption.
- [ ] Define and implement `IdentityGenesis` with the following fields:
  - version
  - public key
  - wallet address
  - timestamp
  - proof-of-work difficulty
  - proof-of-work nonce
  - signature
- [ ] Implement proof-of-work on identity creation using:
  - `H = SHA-256(version || public_key || wallet_addr || timestamp || difficulty || nonce)`
  - the first `difficulty` bits must be zero
  - calibration target of approximately 20-30 seconds on median 2024 hardware, with an initial planning value around difficulty 22
- [ ] Define canonical deterministic encoding for every signed state:
  - raw JSON signatures are forbidden
  - protobuf payloads must use deterministic serialization
  - non-protobuf signed artifacts must use an equally canonical documented binary encoding
- [ ] Define signed payload classes and verification rules for at least:
  - `IdentityGenesis`
  - `SessionKeyDelegation`
  - join, resume, and action envelopes
  - checkpoint acknowledgements
  - settlement proofs and dispute references introduced later
- [ ] Implement per-session ephemeral `Ed25519` session keys plus `SessionKeyDelegation` signed by the parent key, scoped to table identifier and validity window.
- [ ] Add `ZKCredential` hooks with the following structure:
  - claim type
  - proof bytes, maximum 4KB
  - issuer DID
  - expiration
  - maximum 3 credentials per request
  - strictly defensive parsing and storage
- [ ] Add TUI identity management flows for:
  - create identity
  - restore identity from mnemonic
  - export mnemonic
  - view public keys and wallet address
  - show proof-of-work progress during identity creation
- [ ] Add tests covering deterministic encoding, seed encryption and decryption, proof-of-work validation, session-key delegation validation, and mnemonic recovery.

### Acceptance Criteria

- Identity creation, restore, export, and public-key inspection work end to end.
- The root seed is encrypted at rest and recoverable only through the documented unlock path.
- All signed states have a deterministic canonical encoding, and none rely on raw JSON bytes.
- Session keys are ephemeral, delegated correctly, and scoped to a single table and validity window.

## Track 5. P2P Network, Discovery, Lobby, and Session Protocol

### Objective

Build the concrete heads-up P2P network stack and session protocol for free-play tables, carrying forward `Session Authority`, `Prompt Envelope`, and transcript sequencing into a real network runtime.

### Dependencies

Track 4 must be complete.

### Deliverables

- [ ] Build the libp2p host stack using:
  - QUIC transport
  - Noise security protocol
  - Yamux multiplexer
  - the `Ed25519` peer identity key from the identity module
- [ ] Add mDNS for local and LAN discovery.
- [ ] Add Kademlia DHT for global discovery with hardcoded but replaceable bootstrap nodes.
- [ ] Add GossipSub v1.1 for lobby functionality using:
  - `ante/lobby/free/v1` for table announcements
  - reserved later topics `ante/free-heads/v1` and `ante/free-locks/v1`, which become active in Track 6
- [ ] Add DCUtR and AutoRelay for NAT traversal and relay fallback. Relay nodes are initially project-maintained, but the list must be configurable and community-replaceable. The protocol must never depend on any specific relay.
- [ ] Enforce heads-up cash as the sole P2P table format. No tournaments, 6-max, or 9-max P2P offers may be created or joined.
- [ ] Move the current `Session Authority`, `Prompt Envelope`, transcript checkpointing, and byte-stable reducer-facing envelopes into the network session protocol layer.
- [ ] Keep the loopback sandbox capable of using the same envelope shapes for testing, but make the protocol definition itself belong to this track.
- [ ] Define protobuf wire messages for `TableOffer` with the following fields:
  - peer ID
  - session public key
  - context
  - blinds
  - min buy-in and max buy-in
  - action timeout
  - session timeout
  - required zero-knowledge claims, empty for now
  - protocol version
  - creation timestamp
  - signature
- [ ] Define protobuf wire messages for `JoinRequest` with the following fields:
  - peer ID
  - session key delegation
  - requested buy-in
  - `BalanceChain` head proof placeholder, accepted as a stub until Track 6
  - zero-knowledge credentials, empty for now
  - protocol version
  - signature
- [ ] Implement a host-authoritative placeholder `DeckSource` using commit-reveal with `crypto/rand`, explicitly marked as insecure and temporary. Every action must still be signed and transcripted exactly as it will be with the final fairness backend. This harness is for network bringup only and is not valid for persistent-value production play.
- [ ] Implement a reconnection protocol with:
  - 60-second timeout
  - nonce-based state synchronization
  - explicit handling for in-hand disconnects versus between-hand disconnects
- [ ] Integrate peer scoring with GossipSub v1.1 scoring. Negative score sources must include:
  - invalid messages
  - excessive rate
  - failed handshakes
  - short or low-quality sessions
- [ ] Add rate limiting at the transport layer.
- [ ] Standardize free table tiers:
  - Micro: blinds 1/2, buy-in 40-200
  - Low: blinds 2/5, buy-in 100-500
  - Mid: blinds 5/10, buy-in 200-1000
- [ ] Add a TUI lobby view for:
  - listing tables
  - creating a table
  - joining a table
  - filtering by tier
- [ ] Add integration tests proving that two clients on different machines can discover each other, complete handshake, exchange identical transcripts, and recover from reconnect under the defined timeout rules.

### Acceptance Criteria

- Two clients on different machines discover each other via the DHT, one creates a table, the other joins, they play a complete heads-up session, and both transcripts are identical byte for byte.
- The network stack works across LAN and relay-assisted paths without depending on any single relay.
- Heads-up cash is enforced everywhere in offer creation, join validation, and gameplay routing.
- The temporary host-authoritative bringup harness is explicitly limited to non-economic development use.

## Track 6. BalanceChain, Free-Play Economy, and Free Tables Alpha

### Objective

Implement the persistent free-play chip economy, symmetric settlement model, and publicly usable free-table alpha, while preserving economy isolation and the lazy-applied refill model.

### Dependencies

Track 5 must be complete. `TimeAnchorProvider` remains a mandatory dependency throughout this track.

### Deliverables

- [ ] Implement `BalanceChain` as a hash-linked chain of signed balance entries for one identity.
- [ ] Define every `BalanceChain` entry with the following common fields:
  - type
  - monotonically increasing sequence number with no gaps
  - amount
  - resulting balance
  - timestamp
  - time anchor
  - previous entry hash using `SHA-256`
  - conditional fields by type
  - owner signature
  - entry hash
- [ ] Implement the following entry types:
  - `GenesisGrant`
  - `SessionLock`
  - `SessionUnlock`
  - `SessionSettlement`
  - `DailyRefillClaim`
  - `EmergencyRefillClaim`
  - `SessionBonusClaim`
  - `CheckpointEntry`
  - `AbandonmentEvidence`
- [ ] Define `GenesisGrant` as exactly one grant per identity, linked to `IdentityGenesis`, with a starting balance of 2500 free chips.
- [ ] Implement lazy-applied refill policies using `TimeAnchorProvider`:
  - Daily refill: if balance is below 1000, credit to 1000, with a 24-hour cooldown
  - Emergency refill: if balance is below 250, credit to 250, with a 10-minute cooldown
  - Session bonus: plus 100 if balance is below 500 after completing a session normally, maximum one claim per session
- [ ] Implement `SessionLock` so a buy-in is reserved before joining a table. Maximum one active lock may exist per identity at a time.
- [ ] Implement `SessionUnlock` for failed joins, canceled sessions, or other cases where a lock is released without settlement.
- [ ] Implement `SessionSettlement` as a co-signed result. Each player must add their own settlement entry containing the counterparty's signature confirming the result.
- [ ] Implement `CheckpointEntry` as a co-signed compaction point so a player can present checkpoint plus subsequent entries rather than the entire chain.
- [ ] Implement `VerifyBalanceChain` to validate, during join and settlement flows:
  - genesis validity
  - proof-of-work validity
  - chain integrity
  - signature validity
  - counterparty signatures on settlements and checkpoints
  - cooldown compliance for refills
  - threshold compliance for refill claims
  - balance never negative
  - no active unresolved lock
  - target verification cost under 100ms for 1000 entries
- [ ] Add head gossip on `ante/free-heads/v1` using `HeadAnnouncement` with the following fields:
  - identity public key
  - sequence
  - head hash
  - timestamp
  - signature
  - peers cache known heads and detect equivocation when the same sequence is advertised with different hashes
- [ ] Add lock gossip on `ante/free-locks/v1`.
- [ ] Implement abandonment handling so that:
  - the pot goes to the remaining player according to the protocol rules
  - `AbandonmentEvidence` is recorded in the chain
  - a lock without settlement remains visible to future peers and verification logic
- [ ] Persist the free-play economy under the following paths:
  - `~/.ante/free/balance_chain.bin`
  - `~/.ante/free/checkpoints/`
  - `~/.ante/free/locks/`
  - `~/.ante/cache/free_heads/`
- [ ] Add symmetric settlement and dispute preservation for every balance-affecting table. Both peers must retain the latest mutually signed checkpoint or settlement plus transcript tail. Persistent-value free tables may not rely on host-only evidence.
- [ ] Remove the Track 5 bringup harness from any balance-affecting production path. Free-play alpha tables must not run on a host-authoritative trust model.
- [ ] Deliver the free-table alpha product surface with the following scope:
  - heads-up cash only
  - no bots
  - no tournaments
  - no spectators
  - no chat
  - no leaderboard
  - no gifts or off-table transfers
- [ ] Add TUI surfaces for buy-in lock state, chain-verification failure messaging, settlement confirmation, reconnect recovery, and free-table join feedback.
- [ ] Add integration tests covering join verification, lock lifecycle, refill claims, settlement signing, abandonment evidence, reconnect without duplicate effects, and head equivocation detection.

### Acceptance Criteria

- A join handshake verifies `BalanceChain` state within the target time budget and rejects unresolved active locks or invalid chains.
- Two players can complete a free heads-up session, produce co-signed settlements, update their chains, and advertise matching heads to the network.
- Reconnect, abandonment, and failed joins do not mint, burn, or duplicate free-play chips incorrectly.
- Free Tables multiplayer is publicly usable in alpha after this track, which is the point at which `README.md` may be rewritten to reflect the new project scope.

## Track 7. Verifiable Fairness, Auditability, and Stable Public Claims

### Objective

Provide the evidence, tooling, migration guarantees, and operational runbooks required for stable public integrity claims across gameplay, transcripts, and free-play balance effects.

### Dependencies

Track 6 must be complete and operationally proven. `TimeAnchorProvider` remains a hard requirement for stable sign-off.

### Deliverables

- [ ] Replace the temporary network bringup deck source with a documented shuffle or deck-integrity scheme suitable for external verification.
- [ ] Add transcript export and a standalone verification tool that can reproduce a hand outcome from exported artifacts alone.
- [ ] Extend the verifier so it validates:
  - transcript integrity
  - canonical signed-state bytes and signatures
  - mutually signed settlements and checkpoints
  - `BalanceChain` effects for the session
  - dispute bundles and abandonment evidence where applicable
- [ ] Formalize symmetric dispute procedures so either side can submit the latest mutually signed state plus transcript evidence and reach the same verification outcome.
- [ ] Version and migrate artifacts across sandbox, identity, network cache, transcript, and free-play economy state with explicit compatibility guarantees.
- [ ] Publish internal runbooks for:
  - key compromise
  - transcript corruption
  - time-anchor outage
  - head equivocation and settlement mismatch
  - bootstrap or relay outage
  - artifact migration rollback
- [ ] Only after the above is complete, update public-facing language to describe fairness or auditability characteristics. No "provably fair" claim may appear earlier.

### Acceptance Criteria

- An independent verifier can reproduce a hand outcome and its balance effects from exported artifacts using documented procedures.
- Artifact migrations across supported versions are tested and reversible within documented limits.
- Stable release sign-off includes transcript verification, identity validation, settlement validation, `BalanceChain` correctness, and time-anchor failure handling.

## Appendix A. Dependency Graph

```text
Track 0 -> Track 1 -> Track 2 -> Track 3 -> Track 4 -> Track 5 -> Track 6 -> Track 7

Track 2 internal ordering:
  2.1 Save / Load / Continue
    -> 2.2 Statistics Persistence
    -> 2.3 Hand History Browser
    -> 2.4 Hand Replay

  2.2 Statistics Persistence
    -> 2.5 Results Screen

  2.6 Cash Game Continuity
    -> sandbox lifecycle maturity required by Track 3

  2.7 Setup Screen Completeness
    -> truthful sandbox configuration surface required by Track 3

  2.8 Settings Effectiveness
    -> presentation maturity required by Track 3

  2.9 Table Visual Clarity
    -> premium sandbox quality and later heads-up P2P table clarity

  2.10 Tournament UX
    -> sandbox-only tournament maturity required by Track 3

  2.11 Bot Reasoning Display
    -> sandbox-only UX

  2.12 AI Strategic Depth
    -> sandbox-only gameplay depth

  2.13 Test Coverage
    -> required alongside every Track 2 item above
```

## Appendix B. Storage Layout

The exact on-disk encoding may evolve under `ArtifactStore`, but the intended logical layout after Track 3 is:

```text
~/.ante/
  config.json
  manifest.json
  migrations/
    *.json
  sandbox/
    saves/
      slot-1.*
      slot-2.*
      ...
    stats/
      sessions/
        <session-id>.*
      aggregates/
        current.*
    transcripts/
      <session-id>/
        chunk-<n>.*
        checkpoint-<seq>.*
    history/
      indexes/
        sessions.*
      hands/
        <session-id>/
          <hand-id>.*
  identity/
    seed.enc
    identity_genesis.*
    profiles/
      <identity-id>.*
    session_keys/
      <table-id>/
        <session-key-id>.*
  free/
    balance_chain.bin
    checkpoints/
      <checkpoint-id>.*
    locks/
      <lock-id>.*
    settlements/
      <session-id>.*
  cache/
    free_heads/
      <identity-id>.*
    network/
      peers/
        <peer-id>.*
      offers/
        <offer-id>.*
  disputes/
    <session-id>/
      bundle.*
```

Notes:

- The current `stats.gob` and `saves/slot_*.gob` files are migration sources, not the target architecture.
- `ArtifactStore` owns these namespaces even though they are not grouped under a single `artifacts/` directory.
- Sandbox artifacts and free-play persistent economy artifacts must remain physically and logically separated.
- External tools must not depend on physical filenames or encodings; they must depend on documented artifact semantics and versioning rules instead.

## Appendix C. Sandbox Backlog

These items remain sandbox-only unless explicitly promoted by a later roadmap revision:

- advanced AI phases beyond the baseline needed to keep the local game interesting
- bot reasoning overlays and educational UI
- local-only save-slot ergonomics beyond the core resumable flow
- deterministic seed exposure and other expert sandbox setup controls
- premium theme variants and animation experimentation beyond the default production presentation
- local tournament milestone storytelling and cosmetic presentation flourishes
- offline replay enhancements aimed at analysis rather than multiplayer operations
- any use of bots outside sandbox play; bots do not graduate into multiplayer under this roadmap

No item in this appendix should silently expand into multiplayer scope, economy scope, or social scope.

## Appendix D. Documentation Updates

- `README.md` will be rewritten to reflect the new project scope only after Free Tables multiplayer reaches a publicly usable alpha, which occurs after Track 6.
- `CONTRIBUTING.md` will be updated after Track 1 to reflect the new architecture boundaries, module structure, and artifact ownership rules.
- `SECURITY.md` will be updated after Track 4 to reflect the cryptographic identity system and the responsible disclosure scope for protocol-level vulnerabilities.
- None of these documentation updates blocks any track.
