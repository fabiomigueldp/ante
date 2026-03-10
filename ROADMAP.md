# Roadmap

This document tracks known gaps between the current implementation and the intended feature set. Each section describes what exists today, what is missing, and what completing it would involve. Items are grouped by functional area and roughly ordered by priority.

---

## 1. Save / Load / Continue

### Current State

- `storage.SaveGame()` and `storage.LoadGame()` serialize and deserialize a `SaveSlot` using gob encoding. `storage.ListSaves()` enumerates five slot files.
- The `LoadGameModel` screen lists saved slots and allows deletion.
- The pause menu displays a "Save" option bound to `S`.

### What Is Missing

- **Save is not wired.** Pressing `S` in the pause menu enters a `TODO` path that calls `setMessage("Game saved!")` without actually calling `storage.SaveGame()`. No `SaveSlot` is constructed from the live session state. (See `internal/tui/game.go`, `handlePauseKey`.)
- **Load does not resume.** Pressing `Enter` on a saved slot in `LoadGameModel` hits a `TODO` and redirects back to the main menu. There is no code path that reconstructs a `session.Session` from a `SaveSlot`. (See `internal/tui/loadgame.go`, `Update`.)
- **SaveSlot schema is incomplete for mid-hand state.** The current `SaveSlot` stores table metadata, player stacks, bot seeds, config, and a simplified hand history, but does not capture the live hand state: current street, action seat, per-round bet amounts, pot and side-pot breakdown, remaining deck order, or transient flags needed to resume a hand in progress without inconsistency.

### What Completion Requires

1. Add a `buildSaveSlot()` method on `Session` (or on the TUI side using `Session` state) that snapshots all necessary data, including either a full in-hand state or a marker indicating the session should resume from the next hand boundary.
2. Extend `SaveSlot` with fields for mid-hand state if the decision is to support saving during a hand (not just between hands).
3. Wire the pause menu `S` key to call `storage.SaveGame()` with the constructed slot.
4. Implement a `resumeSession()` function that reconstructs a `Session` from a `SaveSlot`, including table, players, bot instances with correct seeds, tournament/cash-game state, and blind level.
5. Wire the load game `Enter` action to call `resumeSession()` and transition to `ScreenGame`.
6. Add tests for round-trip save/load integrity.

---

## 2. Statistics Persistence

### Current State

- `storage.StatsStore` holds a list of `SessionStats` and provides aggregate queries (total sessions, hands played, win rate, profit, best hand, average finish, recent sessions).
- `storage.SaveStats()` and `storage.LoadStats()` handle gob serialization.
- `StatsViewModel` reads persisted stats and displays them.
- `SessionStats` has fields for rich per-session data: hands won, flops seen, showdown results, all-in outcomes, biggest pot, largest win, longest streak, and best hand.

### What Is Missing

- **No automatic recording.** The session end flow (`emitSessionEnd` in `internal/session/session.go`) does not construct a `SessionStats` record and does not call `SaveStats()`. The stats store will remain empty unless data is injected manually.
- **Session-level counters are not all tracked.** Some `SessionStats` fields (like `FlopsSeen`, `ShowdownsWon`, `AllInsWon`, `BestHand`, `LargestWin`, `LongestStreak`) are defined but the session loop does not currently accumulate these counters during play.

### What Completion Requires

1. Add accumulators to `Session` (or a helper struct) that track per-session stats as hands are played: increment `HandsWon` on pot awards to the human, track showdown participation, record biggest pot, etc.
2. At session end (in `emitSessionEnd` or called from there), construct a `SessionStats`, call `store.Add()`, and persist via `SaveStats()`.
3. Verify that the stats screen correctly reflects data after a normal game session.
4. Add test coverage for the recording and retrieval path.

---

## 3. Hand History Browser

### Current State

- The "Hand History" menu entry opens `HistoryViewModel`, which loads `StatsStore` and lists entries per session (not per hand). Each row shows an index, mode, total hands played in that session, and a result string.
- The "View Details" label appears in the footer but pressing `Enter` on a selected entry hits a `TODO` comment and does nothing.
- Internally, `engine.SessionHistory` records full `HandRecord` objects (actions, board, players, blinds, seed, events) during a session, but this data is in-memory only and is not persisted beyond the session lifetime.

### What Is Missing

- **No per-hand browsing.** The screen aggregates by session, using `HandsPlayed` as a count rather than listing individual hands. Users cannot navigate to or inspect a specific hand.
- **No persistence of hand records.** `SessionHistory` lives in memory. Once the program exits, all hand-level data is lost. There is no serialization of `HandRecord` objects to disk.
- **View Details is not implemented.** The `Enter` key handler is a placeholder.

### What Completion Requires

1. Design a persistent hand history format, either extending the stats store or creating a separate per-session file that stores `[]HandRecord`.
2. Record hand history to disk at session end (or incrementally after each hand).
3. Update `HistoryViewModel` to support two levels of navigation: session list and per-hand list within a selected session.
4. Wire the `Enter` action to either expand the session into its hands or navigate to the replay screen for the selected hand.
5. Consider a retention policy for long histories (max records, pruning old sessions).

---

## 4. Hand Replay

### Current State

- `ReplayModel` exists and can step forward and backward through the actions of a `HandRecord`. It displays players, blinds, dealer seat, board cards, and a progress bar.
- The model handles keyboard navigation (left/right to step, home/end to jump, esc to go back).

### What Is Missing

- **Not connected to the history browser.** The transition from `ScreenHistory` to `ScreenReplay` in `app.go` does not pass a valid `HandRecord`. The `ScreenReplay` case attempts a type assertion against `*session.Session` but does not extract a record from it. (See `internal/tui/app.go`, `switchTo`, `ScreenReplay` case.)
- **Board reconstruction is approximate.** `boardAtStep()` uses a comment-documented heuristic: it shows all board cards as soon as any action has been taken, rather than revealing them street by street based on action boundaries. The code has a `streets` counter that is never actually used.
- **No hole card display.** The replay does not show the human's hole cards or any revealed cards beyond what `HandRecord` stores.

### What Completion Requires

1. Store or reconstruct street boundaries in the action list (either tag actions with their street or detect street transitions by counting betting round completions).
2. Update `boardAtStep()` to progressively reveal flop (3 cards), turn (1 card), and river (1 card) at the correct action indices.
3. Wire the history browser's `Enter` key to pass the selected `HandRecord` to `ReplayModel` via the `switchScreenMsg` data field.
4. Update the `ScreenReplay` case in `app.go` to properly extract and pass the `HandRecord`.
5. Optionally show hole cards for the human player and any cards revealed at showdown.

---

## 5. Cash Game Continuity

### Current State

- The help screen describes cash games as "Play as long as you want. Fixed blinds. Leave anytime with your current stack."
- `engine.CashGame` supports rebuy offers (`OfferRebuy`), player replacement (`ReplacePlayer`), and cash-out (`CashOut`).
- The session loop runs cash games, but terminates when the human busts or the table hits a standard end condition.

### What Is Missing

- **No open-ended session.** The session does not offer rebuys to the human or replace busted bots with new ones. When the human is eliminated, the session ends.
- **No voluntary exit.** There is no "Stand Up" or "Leave Table" action available during play. The player can only quit via the pause menu, which aborts the session entirely. The help text's promise of "leave anytime with your current stack" is not reflected in the actual flow.
- **No profit tracking during session.** The `CashGame.Profit` map exists but is not surfaced in the results screen or stats.

### What Completion Requires

1. After each hand, check if the human busted and offer a rebuy prompt.
2. Replace eliminated bots with new characters (or rebuy them) to maintain table size.
3. Add a "Stand Up" action (perhaps via the pause menu or a dedicated key) that calls `CashOut`, records profit, and transitions to the results screen.
4. Surface cash game profit in the results screen and stats recording.
5. Update the session loop to continue indefinitely until the player chooses to leave.

---

## 6. Setup Screen Completeness

### Current State

- The setup screen exposes five configurable fields: game mode, seats, difficulty, starting stack, and player name.
- `session.Config` supports additional fields: `Seed`, `BlindSpeed`, `CashGameBuyIn`, and `CashGameBlinds`.

### What Is Missing

The backend already supports configuration options that the UI does not expose:

- **Blind speed** -- The `BlindSpeed` field allows controlling how quickly tournament blinds escalate, but the setup screen does not offer this choice.
- **Cash game blinds** -- `CashGameBlinds` allows custom small/big blind values for cash games, but the UI uses engine defaults.
- **Cash game buy-in** -- `CashGameBuyIn` controls the initial buy-in amount separately from starting stack.
- **Seed** -- Deterministic seeds for reproducible sessions are supported but not exposed.

### What Completion Requires

1. Add setup fields for blind speed (with presets like "Slow", "Normal", "Fast", "Turbo").
2. Show cash-game-specific fields (custom blinds, buy-in) conditionally when the Cash Game mode is selected.
3. Optionally expose an advanced section or a separate "Advanced Setup" screen for seed and other expert options.

---

## 7. Settings Effectiveness

### Current State

- The settings screen allows editing and persisting: player name, sound toggle, volume, pot odds toggle, animation speed, default mode/difficulty/seats/stack, and theme.
- Sound and pot odds settings are functional and take effect during gameplay.

### What Is Missing

- **Animation speed has no effect.** `advanceAnimation()` in `internal/tui/game.go` is a no-op that immediately returns. The `AnimationSpeed` config value is saved but never read during gameplay.
- **Theme has no visual effect.** The `Theme` field is persisted and editable ("classic", "dark", "green"), but the TUI rendering does not branch on theme values. The visual appearance is the same regardless of the selected theme.

### What Completion Requires

1. **Animation:** Define what animation means in the context of a TUI poker game (e.g., delay between card reveals, timed display of opponent actions, chip movement simulation). Implement `advanceAnimation()` to introduce configurable pauses using `tea.Tick`. Respect the "off" setting by skipping all delays.
2. **Theme:** Create color palettes for each theme option. Update the style constants in `internal/tui/theme.go` to be functions of the active theme rather than fixed values. Reload styles when the theme changes.

---

## 8. AI Strategic Depth

### Current State

- The bot decision engine evaluates hand strength, draw potential, pot odds, and table pressure. It uses the profile's `Aggression`, `Bluff`, `CallDown`, and `Tilt` parameters to modulate thresholds for folding, calling, and raising.
- Raise sizing considers the profile's `LargeBetBias` and `Aggression`.
- Tilt accumulates on big losses and decays between hands.

### What Is Missing

Several profile attributes are defined and assigned to characters but have minimal or no influence on the decision path:

- **VPIP (Voluntarily Put money In Pot)** -- Defined in the profile but not used to gate pre-flop entry decisions.
- **PFR (Pre-Flop Raise)** -- Defined but not used to differentiate pre-flop raising frequency from post-flop.
- **Trap** -- Characters like "Phantom" have high trap values, but the `pickAction` logic does not implement trapping behavior (e.g., checking strong hands to induce bluffs, then raising).
- **DrawBias** -- Assigned per profile but not used to weight draw-heavy decisions differently.
- **HeroCallBias** -- Assigned but not used to make call-down-heavy players more resistant to folds facing large bets.
- **ThreeBetBias** -- Assigned but not used to drive three-bet frequency.
- **PositionBias** -- Assigned but given only minor weight in the current formula.

Additionally, the decision engine currently lacks:

- **Pre-flop hand charts or ranges.** All streets use the same threshold logic.
- **Multiway pot adjustments.** The bot does not tighten or loosen based on the number of active opponents.
- **Board texture reads.** Decisions do not consider whether the board is wet, dry, paired, or connected.
- **Bet sizing variety.** Sizing is formulaic and does not adapt to stack-to-pot ratio, opponent tendencies, or street.
- **Inter-street line coherence.** Each street's decision is independent; there is no concept of planning a line across streets (e.g., "bet flop, barrel turn, check river").
- **Tournament/ICM pressure.** Bots do not adjust play for bubble, final table, or stack-relative situations.
- **Short stack play.** No push/fold mode for low effective stacks.

### What Completion Requires

This is the largest area of work. A phased approach is recommended:

1. **Phase 1:** Wire VPIP and PFR to gate pre-flop action selection. Use pre-flop hand groupings (premium, strong, marginal, speculative) to create distinct ranges per profile.
2. **Phase 2:** Implement trapping behavior for high-Trap profiles. Add hero-calling logic for high-HeroCallBias profiles. Use ThreeBetBias to drive re-raising frequency.
3. **Phase 3:** Add board texture evaluation (pair count, flush/straight potential, connectivity) and use it to modulate post-flop aggression.
4. **Phase 4:** Introduce stack-aware play (push/fold for short stacks, pot control for deep stacks) and tournament pressure adjustments.

---

## 9. Bot Reasoning in the UX

### Current State

- When a bot is deciding, the session emits a `bot_thinking` event with `ThinkTime` and the bot's name. The TUI displays "{BotName} is thinking..." in the action bar.
- After the decision, the bot's `Reason` string (e.g., "value pressure", "stab", "fold") is stored in the `SessionEvent` but is not displayed anywhere in the UI.

### What Is Missing

- The reasoning string is discarded at the display level. There is an opportunity to show a brief explanation of the bot's decision (e.g., "Blitz raises -- value pressure") that would add personality and educational value.
- Think time variations exist but are not visually differentiated (e.g., a longer think on a tough decision could be more visually apparent).

### What Completion Requires

1. After a bot action event, display the bot's reasoning string briefly in the message area or alongside the action description in the action bar.
2. Optionally gate this behind a setting ("Show Bot Reasoning") for players who prefer less information.

---

## 10. Table Visual Clarity

### Current State

- Opponents are rendered in horizontal rows of seat blocks, centered in the terminal width. Each seat shows name, stack, status (folded/all-in/bet), and face-down cards (or revealed cards at showdown).
- The active player highlight uses `SeatStyle(true, false, ...)` where the first argument is always `true` for opponents, meaning all non-folded opponents appear with the same visual weight.
- The human player area is rendered separately at the bottom with large card art.

### What Is Missing

- **Active player highlight.** The player whose turn it is to act is not visually distinguished from other active players. All opponent seats use the same "active" style regardless of whose turn it is.
- **Action recency.** The last action taken is shown in the action bar text, but there is no per-seat visual indicator of what each player did most recently (e.g., a small "raised" label on their seat).
- **Street progression indicator.** Beyond the street name shown below the board, there is no visual marker of progression through the hand (e.g., a line or dots showing Pre-Flop > Flop > Turn > River).
- **Table geometry.** Opponents are laid out in a flat row or wrapped rows. A more table-like arrangement (semi-circular or oval) would improve spatial awareness but is complex to implement in a TUI.

### What Completion Requires

1. Pass the current action seat ID to the rendering layer and apply a distinct style (e.g., brighter border, underline, or marker) to the seat whose turn it is.
2. Add a per-seat last-action label that briefly shows what the player did (fades or clears on the next action).
3. Consider adding a street progress bar or indicator in the board area.

---

## 11. Tournament UX

### Current State

- Tournament mode works correctly: blinds increase, players are eliminated with positional ranking, and the session ends when one player remains.
- The header shows the current blind level. Blind increase events are announced via a message.

### What Is Missing

- **Blind structure visibility.** Players cannot see the upcoming blind levels or how many hands remain until the next increase.
- **Field status.** There is no display of how many players remain, average stack, or the player's relative standing.
- **Progression feel.** The tournament does not communicate the narrative arc (early levels, bubble, final table) visually.

### What Completion Requires

1. Add a tournament HUD element (or expand the header) showing: current level, next blind level, hands until increase, players remaining out of started, and average stack.
2. Optionally show a chip leader / short stack indicator.
3. Add event messages for milestone moments (bubble, final table, heads-up).

---

## 12. Results Screen

### Current State

- The results screen shows a main result message, hands played count, and final standings sorted by stack. It attempts to find a biggest pot from hand history but the extraction is incomplete (the code has a comment acknowledging it cannot easily extract amounts from the event interface).

### What Is Missing

- **Biggest pot display.** The code structure is present but the value is never populated.
- **Session highlights.** No display of best hand, longest win streak, or memorable moments from the session.
- **Stat comparison.** No comparison against the player's historical averages.

### What Completion Requires

1. Track biggest pot, best hand, and win/loss streak during the session (same counters needed for stats persistence).
2. Display these highlights on the results screen.
3. Optionally show delta against career averages if stats are available.

---

## 13. Test Coverage for Product Features

### Current State

- The core engine has good test coverage: betting, pot calculation, showdowns, deck operations, table management, and tournaments all have dedicated test files.
- Session, AI, TUI, and audio layers have tests, but they focus on unit-level behavior rather than end-to-end product flows.

### What Is Missing

- **No save/load round-trip tests.** There are no tests that construct a `SaveSlot`, serialize it, deserialize it, and verify integrity.
- **No session resume tests.** No tests verify that a session can be reconstructed from saved state.
- **No stats recording tests.** No tests verify that completing a session writes correct statistics.
- **No end-to-end flow tests.** No tests simulate a full user journey (start game, play hands, save, quit, load, resume, finish, verify stats).

### What Completion Requires

Add integration tests that exercise:

1. `SaveGame` / `LoadGame` round-trip with field-level assertions.
2. Session reconstruction from a loaded `SaveSlot`.
3. Statistics recording after session completion.
4. Replay model stepping through a real `HandRecord`.

---

## Priority Guidance

The following order is suggested for maximum product impact:

1. **Save / Load / Continue** -- Removes the most visible gap between promise and delivery.
2. **Statistics Persistence** -- Low implementation cost, high perceived value.
3. **AI Strategic Depth (Phase 1)** -- Pre-flop ranges and better use of existing profile parameters would significantly improve gameplay feel.
4. **Cash Game Continuity** -- Fulfills the described cash game experience.
5. **Hand History + Replay wiring** -- Connects existing pieces into a working feature.
6. **Results Screen enrichment** -- Quick win using data that will be tracked for stats.
7. **Tournament UX** -- Adds competitive depth.
8. **Bot Reasoning display** -- Small change, significant personality payoff.
9. **Settings Effectiveness** -- Themes and animation will polish the experience.
10. **Setup Screen Completeness** -- Exposing backend capabilities to the user.
11. **Table Visual Clarity** -- Iterative UX refinements.
12. **AI Strategic Depth (Phases 2-4)** -- Longer-term strategic improvement.
13. **Test Coverage** -- Should grow alongside each feature implementation.

---

## Release Automation

Separately from the feature roadmap, the CI pipeline could be extended with:

- **Cross-platform release builds** via GoReleaser or equivalent, producing downloadable binaries for Linux, macOS, and Windows.
- **Benchmark suites** to track engine performance over time.
- **Profiling workflows** to identify hot paths in the simulation runner.
