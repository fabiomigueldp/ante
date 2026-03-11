package session

import (
	"bytes"
	"errors"
	"fmt"
	"sort"

	"github.com/fabiomigueldp/ante/internal/ai"
	"github.com/fabiomigueldp/ante/internal/engine"
	"github.com/fabiomigueldp/ante/internal/storage"
)

var (
	ErrSaveMidHandNotSupported = errors.New("cannot save during an active hand")
	ErrSaveSlotUnavailable     = errors.New("save slot unavailable")
	ErrSaveIntegrityMismatch   = errors.New("save integrity mismatch")
)

func (s *Session) CanSave() bool {
	if s == nil {
		return false
	}
	return s.Phase == PhaseWaitingReady && s.readyState != nil && s.readyState.HumanPending
}

func (s *Session) BuildSaveArtifact() (*storage.SaveSlot, error) {
	if s == nil {
		return nil, fmt.Errorf("session is nil")
	}
	if !s.CanSave() {
		return nil, ErrSaveMidHandNotSupported
	}
	deps := s.deps
	if deps.ArtifactStore == nil || deps.TimeAnchorProvider == nil {
		deps = sessionDependenciesProvider()
	}
	anchor, err := deps.TimeAnchorProvider.Now()
	if err != nil {
		return nil, err
	}
	slot := &storage.SaveSlot{
		SchemaVersion:  1,
		SessionID:      s.SessionID,
		LastSeq:        s.seq,
		LifecyclePhase: saveLifecyclePhase(s.Phase),
		Name:           defaultSaveName(s),
		Timestamp:      anchor.Timestamp,
		Mode:           modeString(s.Config.Mode),
		HandNumber:     s.HandCount,
		PlayerName:     s.Config.PlayerName,
		PlayerStack:    currentHumanStack(s),
		TotalPlayers:   len(s.Table.Players),
		ActivePlayers:  len(s.Table.ActivePlayers()),
		BlindLevel:     s.Table.CurrentBlinds().Level,
		TableData: storage.TableSaveData{
			Mode:         s.Table.Mode,
			Seats:        s.Table.Seats,
			DealerSeat:   s.Table.DealerSeat,
			HandNumber:   s.Table.HandNumber,
			CurrentLevel: s.Table.CurrentLevel,
			MasterSeed:   s.Table.MasterSeed,
			BlindsConfig: s.Table.BlindsConfig,
		},
		Players:   buildPlayerSaves(s),
		BotSeeds:  map[engine.PlayerID]int64{},
		BotStates: map[engine.PlayerID]storage.BotStateSave{},
		Config: storage.GameConfig{
			Mode:           s.Config.Mode,
			Difficulty:     difficultyCode(s.Config.Difficulty),
			Seats:          s.Config.Seats,
			StartingStack:  s.Config.StartingStack,
			BlindSpeed:     s.Config.BlindSpeed,
			PlayerName:     s.Config.PlayerName,
			Seed:           s.Config.Seed,
			CashGameBuyIn:  s.Config.CashGameBuyIn,
			CashGameBlinds: s.Config.CashGameBlinds,
		},
		Metrics: storage.SessionMetricsSnapshot{},
		History: buildHistorySaves(s),
	}
	if s.metrics != nil {
		slot.Metrics = s.metrics.Snapshot()
	}
	if s.readyState != nil {
		slot.Boundary = storage.BoundaryStateSave{
			Active:          true,
			HandID:          s.readyState.HandID,
			Snapshot:        saveBoundarySnapshot(s.readyState.Snapshot),
			HumanPending:    s.readyState.HumanPending,
			HumanCanLeave:   s.readyState.HumanCanLeave,
			LastResultLabel: s.readyState.LastResultMessage,
		}
	}
	for playerID, bot := range s.Bots {
		state := bot.State()
		slot.BotSeeds[playerID] = state.Seed
		slot.BotStates[playerID] = storage.BotStateSave{Seed: state.Seed, DrawCount: state.DrawCount, TiltLevel: state.TiltLevel}
	}
	if s.Tournament != nil {
		slot.Tournament = storage.TournamentSaveData{
			HandsAtLevel: s.Tournament.HandsAtLevel,
			Eliminations: buildEliminationSaves(s.Tournament.Eliminations),
		}
	}
	if s.CashGame != nil {
		slot.CashGame = storage.CashGameSaveData{Profit: copyProfitMap(s.CashGame.Profit)}
	}
	slot.Integrity = storage.TranscriptHash{Algorithm: "sha256"}
	sum, err := canonicalSaveHash(*slot)
	if err != nil {
		return nil, err
	}
	slot.Integrity.Sum = sum
	return slot, nil
}

func (s *Session) SaveToSlot(slot int) error {
	artifact, err := s.BuildSaveArtifact()
	if err != nil {
		return err
	}
	return storage.SaveGame(slot, artifact)
}

func canonicalSaveHash(slot storage.SaveSlot) ([]byte, error) {
	slot.Integrity = storage.TranscriptHash{}
	return storage.CanonicalSHA256(slot)
}

func ValidateSaveArtifact(slot *storage.SaveSlot) error {
	if slot == nil {
		return fmt.Errorf("save slot is nil")
	}
	if slot.Integrity.Algorithm == "" || len(slot.Integrity.Sum) == 0 {
		return nil
	}
	if slot.Integrity.Algorithm != "sha256" {
		return fmt.Errorf("unsupported save integrity algorithm: %s", slot.Integrity.Algorithm)
	}
	sum, err := canonicalSaveHash(*slot)
	if err != nil {
		return err
	}
	if !bytes.Equal(sum, slot.Integrity.Sum) {
		return ErrSaveIntegrityMismatch
	}
	return nil
}

func buildPlayerSaves(s *Session) []storage.PlayerSaveData {
	players := make([]storage.PlayerSaveData, 0, len(s.Table.Players))
	for _, player := range s.Table.Players {
		if player == nil {
			continue
		}
		entry := storage.PlayerSaveData{
			ID:        player.ID,
			Name:      player.Name,
			Stack:     player.Stack,
			Status:    player.Status,
			SeatIndex: player.SeatIndex,
			IsHuman:   player.ID == s.HumanID,
		}
		if bot, ok := s.Bots[player.ID]; ok {
			entry.BotID = bot.Character.ID
		}
		players = append(players, entry)
	}
	sort.Slice(players, func(i, j int) bool { return players[i].SeatIndex < players[j].SeatIndex })
	return players
}

func buildHistorySaves(s *Session) []storage.HandRecordSave {
	if s.History == nil {
		return nil
	}
	out := make([]storage.HandRecordSave, 0, len(s.History.Records))
	for _, record := range s.History.Records {
		out = append(out, storage.HandRecordSave{
			HandID:     record.HandID,
			DealerSeat: record.DealerSeat,
			Blinds:     record.Blinds,
			Players:    append([]engine.PlayerSnapshot(nil), record.Players...),
			Board:      append([]engine.Card(nil), record.Board...),
			Actions:    append([]engine.Action(nil), record.Actions...),
			Timestamp:  record.Timestamp,
		})
	}
	return out
}

func buildEliminationSaves(elims []engine.Elimination) []storage.TournamentEliminationSave {
	out := make([]storage.TournamentEliminationSave, 0, len(elims))
	for _, elim := range elims {
		out = append(out, storage.TournamentEliminationSave{PlayerID: elim.PlayerID, Position: elim.Position, HandNum: elim.HandNum})
	}
	return out
}

func copyProfitMap(in map[engine.PlayerID]int) map[engine.PlayerID]int {
	if len(in) == 0 {
		return nil
	}
	out := make(map[engine.PlayerID]int, len(in))
	for playerID, profit := range in {
		out[playerID] = profit
	}
	return out
}

func currentHumanStack(s *Session) int {
	for _, player := range s.Table.Players {
		if player != nil && player.ID == s.HumanID {
			return player.Stack
		}
	}
	return 0
}

func defaultSaveName(s *Session) string {
	mode := modeString(s.Config.Mode)
	return fmt.Sprintf("%s - %s", s.Config.PlayerName, mode)
}

func modeString(mode engine.GameMode) string {
	switch mode {
	case engine.ModeCashGame:
		return "cash"
	case engine.ModeHeadsUpDuel:
		return "headsup"
	default:
		return "tournament"
	}
}

func difficultyCode(difficulty ai.Difficulty) int {
	switch difficulty {
	case ai.DifficultyEasy:
		return 0
	case ai.DifficultyHard:
		return 2
	default:
		return 1
	}
}

func saveLifecyclePhase(phase Phase) string {
	switch phase {
	case PhasePlayingHand:
		return "playing_hand"
	case PhaseWaitingReady:
		return "waiting_ready"
	case PhaseSessionOver:
		return "session_over"
	default:
		return "setup"
	}
}

func saveBoundarySnapshot(snapshot TableState) storage.TableSaveBoundaryState {
	return storage.TableSaveBoundaryState{
		HandNum:    snapshot.HandNum,
		HandID:     snapshot.HandID,
		Blinds:     snapshot.Blinds,
		Board:      append([]engine.Card(nil), snapshot.Board...),
		Pot:        snapshot.Pot,
		Street:     snapshot.Street,
		DealerSeat: snapshot.DealerSeat,
		HumanCards: snapshot.HumanCards,
		Players:    makeBoundaryPlayers(snapshot.Players),
		Showdown:   snapshot.Showdown,
		Revealed:   makeBoundaryRevealed(snapshot.Revealed),
		Payouts:    makeBoundaryPayouts(snapshot.ShowdownPayouts),
		PotAwards:  append([]string(nil), snapshot.PotAwards...),
	}
}

func makeBoundaryPlayers(players []PlayerInfo) []storage.PlayerSaveData {
	out := make([]storage.PlayerSaveData, 0, len(players))
	for _, player := range players {
		out = append(out, storage.PlayerSaveData{
			ID:        player.ID,
			Name:      player.Name,
			Stack:     player.Stack,
			Status:    player.Status,
			SeatIndex: player.Seat,
			IsHuman:   player.IsHuman,
		})
	}
	return out
}

func makeBoundaryRevealed(hands []RevealedHand) []storage.BoundaryRevealedSave {
	out := make([]storage.BoundaryRevealedSave, 0, len(hands))
	for _, hand := range hands {
		out = append(out, storage.BoundaryRevealedSave{
			PlayerID: hand.PlayerID,
			Name:     hand.Name,
			Cards:    hand.Cards,
			Eval:     hand.Eval,
		})
	}
	return out
}

func makeBoundaryPayouts(payouts []ShowdownPayout) []storage.BoundaryPayoutSave {
	out := make([]storage.BoundaryPayoutSave, 0, len(payouts))
	for _, payout := range payouts {
		out = append(out, storage.BoundaryPayoutSave{
			PotIndex: payout.PotIndex,
			Winners:  append([]engine.PlayerID(nil), payout.Winners...),
			Amount:   payout.Amount,
			OddChip:  payout.OddChip,
		})
	}
	return out
}
