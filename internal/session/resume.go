package session

import (
	"fmt"
	"math/rand"

	"github.com/fabiomigueldp/ante/internal/ai"
	"github.com/fabiomigueldp/ante/internal/engine"
	"github.com/fabiomigueldp/ante/internal/storage"
)

func ResumeFromSlot(slot int) (*Session, error) {
	save, err := storage.LoadGame(slot)
	if err != nil {
		return nil, err
	}
	return ResumeFromSave(save)
}

func ResumeFromSave(save *storage.SaveSlot) (*Session, error) {
	if err := ValidateSaveArtifact(save); err != nil {
		return nil, err
	}
	if save == nil {
		return nil, fmt.Errorf("save slot is nil")
	}
	players := make([]*engine.Player, 0, len(save.Players))
	bots := make(map[engine.PlayerID]*ai.Bot)
	botOrder := make([]engine.PlayerID, 0)
	humanID := engine.PlayerID(1)
	for _, savedPlayer := range save.Players {
		player := &engine.Player{
			ID:        savedPlayer.ID,
			Name:      savedPlayer.Name,
			Stack:     savedPlayer.Stack,
			Status:    savedPlayer.Status,
			SeatIndex: savedPlayer.SeatIndex,
		}
		players = append(players, player)
		if savedPlayer.IsHuman {
			humanID = savedPlayer.ID
			continue
		}
		character, ok := ai.CharacterByID(savedPlayer.BotID)
		if !ok {
			return nil, fmt.Errorf("unknown bot character id %q", savedPlayer.BotID)
		}
		botState, hasState := save.BotStates[savedPlayer.ID]
		seed := save.BotSeeds[savedPlayer.ID]
		if hasState {
			seed = botState.Seed
		}
		if seed == 0 {
			return nil, fmt.Errorf("missing bot seed for player %d", savedPlayer.ID)
		}
		if hasState {
			bots[savedPlayer.ID] = ai.NewBotFromState(character, ai.BotState{Seed: botState.Seed, DrawCount: botState.DrawCount, TiltLevel: botState.TiltLevel})
		} else {
			bots[savedPlayer.ID] = ai.NewBot(character, seed)
		}
		botOrder = append(botOrder, savedPlayer.ID)
	}

	table, err := engine.NewTable(save.TableData.Mode, save.TableData.Seats, save.TableData.BlindsConfig, save.TableData.MasterSeed, players)
	if err != nil {
		return nil, err
	}
	table.DealerSeat = save.TableData.DealerSeat
	table.HandNumber = save.TableData.HandNumber
	table.CurrentLevel = save.TableData.CurrentLevel

	cfg := Config{
		Mode:           save.Config.Mode,
		Difficulty:     difficultyFromCode(save.Config.Difficulty),
		Seats:          save.Config.Seats,
		StartingStack:  save.Config.StartingStack,
		BlindSpeed:     save.Config.BlindSpeed,
		PlayerName:     save.Config.PlayerName,
		Seed:           save.Config.Seed,
		CashGameBuyIn:  save.Config.CashGameBuyIn,
		CashGameBlinds: save.Config.CashGameBlinds,
	}
	deps := sessionDependenciesProvider()
	sess := &Session{
		Config:     cfg,
		SessionID:  save.SessionID,
		Table:      table,
		History:    restoreHistory(save.History),
		Bots:       bots,
		HumanID:    humanID,
		Phase:      phaseFromSave(save),
		HandCount:  save.HandNumber,
		Updates:    make(chan Envelope, 1024),
		ActionResp: make(chan PlayerActionIntent),
		rng:        rand.New(rand.NewSource(cfg.Seed)),
		botOrder:   botOrder,
		stop:       make(chan struct{}),
		seq:        save.LastSeq,
		resumed:    true,
		deps:       deps,
	}
	metrics := metricsFromSnapshot(save.Metrics)
	if metrics.startTime.Timestamp.IsZero() {
		anchor, anchorErr := deps.TimeAnchorProvider.Now()
		if anchorErr != nil {
			return nil, anchorErr
		}
		metrics = newMetricsAccumulator(anchor)
	}
	transcript, err := newTranscriptWriter(deps.ArtifactStore, deps.TimeAnchorProvider, sess.SessionID, cfg.PlayerName, modeString(cfg.Mode), metrics.startTime)
	if err != nil {
		return nil, err
	}
	sess.metrics = metrics
	sess.transcript = transcript

	switch cfg.Mode {
	case engine.ModeTournament, engine.ModeHeadsUpDuel:
		sess.Tournament = engine.NewTournament(table, sess.startingChips())
		sess.Tournament.HandsAtLevel = save.Tournament.HandsAtLevel
		for _, elim := range save.Tournament.Eliminations {
			sess.Tournament.Eliminations = append(sess.Tournament.Eliminations, engine.Elimination{PlayerID: elim.PlayerID, Position: elim.Position, HandNum: elim.HandNum})
		}
	case engine.ModeCashGame:
		sess.CashGame = engine.NewCashGame(table, cfg.CashGameBuyIn)
		sess.CashGame.Profit = copyProfitMap(save.CashGame.Profit)
	}

	if sess.SessionID == "" {
		sess.SessionID = mustNewSessionID()
	}
	if save.Boundary.Active {
		sess.Phase = PhaseWaitingReady
		sess.readyState = &ReadyState{
			HandID:   save.Boundary.HandID,
			Snapshot: restoreBoundarySnapshot(save.Boundary.Snapshot),
			Ready: map[engine.PlayerID]bool{
				sess.HumanID: !save.Boundary.HumanPending,
			},
			HumanPending:      save.Boundary.HumanPending,
			HumanCanLeave:     save.Boundary.HumanCanLeave,
			LastResultMessage: save.Boundary.LastResultLabel,
		}
	}
	return sess, nil
}

func restoreHistory(records []storage.HandRecordSave) *engine.SessionHistory {
	history := &engine.SessionHistory{}
	for _, record := range records {
		history.Add(engine.HandRecord{
			HandID:     record.HandID,
			Players:    append([]engine.PlayerSnapshot(nil), record.Players...),
			DealerSeat: record.DealerSeat,
			Blinds:     record.Blinds,
			Board:      append([]engine.Card(nil), record.Board...),
			Actions:    append([]engine.Action(nil), record.Actions...),
			Timestamp:  record.Timestamp,
		})
	}
	return history
}

func difficultyFromCode(code int) ai.Difficulty {
	switch code {
	case 0:
		return ai.DifficultyEasy
	case 2:
		return ai.DifficultyHard
	default:
		return ai.DifficultyMedium
	}
}

func phaseFromSave(save *storage.SaveSlot) Phase {
	if save == nil {
		return PhaseSetup
	}
	switch save.LifecyclePhase {
	case "playing_hand":
		return PhasePlayingHand
	case "waiting_ready":
		return PhaseWaitingReady
	case "session_over":
		return PhaseSessionOver
	}
	if save.Boundary.Active {
		return PhaseWaitingReady
	}
	if save.HandNumber > 0 {
		return PhasePlayingHand
	}
	return PhaseSetup
}

func restoreBoundarySnapshot(saved storage.TableSaveBoundaryState) TableState {
	state := TableState{
		HandNum:    saved.HandNum,
		HandID:     saved.HandID,
		Blinds:     saved.Blinds,
		Board:      append([]engine.Card(nil), saved.Board...),
		Pot:        saved.Pot,
		Street:     saved.Street,
		DealerSeat: saved.DealerSeat,
		HumanCards: saved.HumanCards,
		Boundary:   true,
		Showdown:   saved.Showdown,
		PotAwards:  append([]string(nil), saved.PotAwards...),
	}
	for _, revealed := range saved.Revealed {
		state.Revealed = append(state.Revealed, RevealedHand{
			PlayerID: revealed.PlayerID,
			Name:     revealed.Name,
			Cards:    revealed.Cards,
			Eval:     revealed.Eval,
		})
	}
	for _, payout := range saved.Payouts {
		state.ShowdownPayouts = append(state.ShowdownPayouts, ShowdownPayout{
			PotIndex: payout.PotIndex,
			Winners:  append([]engine.PlayerID(nil), payout.Winners...),
			Amount:   payout.Amount,
			OddChip:  payout.OddChip,
		})
	}
	for _, player := range saved.Players {
		state.Players = append(state.Players, PlayerInfo{
			ID:      player.ID,
			Name:    player.Name,
			Stack:   player.Stack,
			Status:  player.Status,
			Seat:    player.SeatIndex,
			IsHuman: player.IsHuman,
		})
	}
	return state
}
