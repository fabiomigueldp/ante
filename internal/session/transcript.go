package session

import (
	"fmt"

	"github.com/fabiomigueldp/ante/internal/engine"
	"github.com/fabiomigueldp/ante/internal/storage"
)

type transcriptRefs struct {
	chunkID        string
	checkpointID   string
	snapshotID     string
	checkpointHash storage.TranscriptHash
}

type TranscriptWriter struct {
	store              storage.ArtifactStore
	timeAnchorProvider storage.TimeAnchorProvider
	sessionID          string
	transcriptID       string
	playerName         string
	mode               string
	nextChunkIndex     int
	lastCheckpointHash storage.TranscriptHash
	head               storage.TranscriptHead
}

func newTranscriptWriter(store storage.ArtifactStore, provider storage.TimeAnchorProvider, sessionID, playerName, mode string, startedAt storage.TimeAnchor) (*TranscriptWriter, error) {
	transcriptID, err := storage.TranscriptIDFromSessionID(sessionID)
	if err != nil {
		return nil, err
	}
	writer := &TranscriptWriter{
		store:              store,
		timeAnchorProvider: provider,
		sessionID:          sessionID,
		transcriptID:       transcriptID,
		playerName:         playerName,
		mode:               mode,
		nextChunkIndex:     1,
		head: storage.TranscriptHead{
			Version:      1,
			SessionID:    sessionID,
			TranscriptID: transcriptID,
			PlayerName:   playerName,
			Mode:         mode,
			StartedAt:    startedAt,
			UpdatedAt:    startedAt,
		},
	}
	if artifact, loadErr := store.LoadTranscriptHeadArtifact(transcriptID); loadErr == nil {
		writer.head = artifact.Payload
		writer.nextChunkIndex = artifact.Payload.HandsPlayed + 1
		writer.lastCheckpointHash = artifact.Payload.LatestChunkHash
	} else if loadErr != storage.ErrArtifactNotFound {
		return nil, loadErr
	}
	return writer, nil
}

func (w *TranscriptWriter) Head() storage.TranscriptHead {
	if w == nil {
		return storage.TranscriptHead{}
	}
	return w.head
}

func (w *TranscriptWriter) CommitHand(s *Session, hand *engine.Hand, record engine.HandRecord) (transcriptRefs, error) {
	if w == nil || hand == nil {
		return transcriptRefs{}, nil
	}
	chunkID, err := storage.ChunkIDFromSessionID(w.sessionID, w.nextChunkIndex)
	if err != nil {
		return transcriptRefs{}, err
	}
	checkpointID, err := storage.CheckpointIDFromSessionID(w.sessionID, hand.ID)
	if err != nil {
		return transcriptRefs{}, err
	}
	snapshotID, err := storage.SnapshotIDFromSessionID(w.sessionID, hand.ID, s.seq)
	if err != nil {
		return transcriptRefs{}, err
	}
	anchor, err := w.timeAnchorProvider.Now()
	if err != nil {
		return transcriptRefs{}, err
	}
	records, err := w.buildRecords(hand, chunkID, snapshotID)
	if err != nil {
		return transcriptRefs{}, err
	}
	chunk := storage.TranscriptChunk{
		Version:          1,
		ID:               chunkID,
		SessionID:        w.sessionID,
		TranscriptID:     w.transcriptID,
		ChunkIndex:       w.nextChunkIndex,
		HandID:           hand.ID,
		SnapshotID:       snapshotID,
		CheckpointID:     checkpointID,
		PreviousHash:     w.lastCheckpointHash,
		Players:          append([]engine.PlayerSnapshot(nil), record.Players...),
		DealerSeat:       hand.DealerSeat,
		Blinds:           hand.Blinds,
		FinalBoard:       append([]engine.Card(nil), hand.Board...),
		HumanStack:       currentHumanStack(s),
		ResultLabel:      handResultLabel(s, hand),
		WinningPlayerIDs: collectWinningPlayerIDs(hand.Winners),
		Records:          records,
		CommittedAt:      anchor,
	}
	sum, err := storage.CanonicalSHA256(struct {
		SessionID        string
		TranscriptID     string
		ChunkID          string
		ChunkIndex       int
		HandID           int
		SnapshotID       string
		CheckpointID     string
		PreviousHash     storage.TranscriptHash
		Players          []engine.PlayerSnapshot
		DealerSeat       int
		Blinds           engine.BlindLevel
		FinalBoard       []engine.Card
		HumanStack       int
		ResultLabel      string
		WinningPlayerIDs []engine.PlayerID
		Records          []storage.TranscriptRecord
	}{
		SessionID:        chunk.SessionID,
		TranscriptID:     chunk.TranscriptID,
		ChunkID:          chunk.ID,
		ChunkIndex:       chunk.ChunkIndex,
		HandID:           chunk.HandID,
		SnapshotID:       chunk.SnapshotID,
		CheckpointID:     chunk.CheckpointID,
		PreviousHash:     chunk.PreviousHash,
		Players:          chunk.Players,
		DealerSeat:       chunk.DealerSeat,
		Blinds:           chunk.Blinds,
		FinalBoard:       chunk.FinalBoard,
		HumanStack:       chunk.HumanStack,
		ResultLabel:      chunk.ResultLabel,
		WinningPlayerIDs: chunk.WinningPlayerIDs,
		Records:          chunk.Records,
	})
	if err != nil {
		return transcriptRefs{}, err
	}
	chunk.CheckpointHash = storage.TranscriptHash{Algorithm: "sha256", Sum: sum}
	if _, err := w.store.SaveTranscriptChunkArtifact(chunk); err != nil {
		return transcriptRefs{}, err
	}
	w.lastCheckpointHash = chunk.CheckpointHash
	w.head = storage.TranscriptHead{
		Version:            1,
		SessionID:          w.sessionID,
		TranscriptID:       w.transcriptID,
		PlayerName:         w.playerName,
		Mode:               w.mode,
		LatestChunkID:      chunk.ID,
		LatestSnapshotID:   snapshotID,
		LatestCheckpointID: checkpointID,
		LatestChunkHash:    chunk.CheckpointHash,
		LatestSeq:          s.seq,
		HandsPlayed:        chunk.ChunkIndex,
		StartedAt:          w.head.StartedAt,
		UpdatedAt:          anchor,
	}
	if _, err := w.store.SaveTranscriptHeadArtifact(w.head); err != nil {
		return transcriptRefs{}, err
	}
	w.nextChunkIndex++
	return transcriptRefs{chunkID: chunk.ID, checkpointID: checkpointID, snapshotID: snapshotID, checkpointHash: chunk.CheckpointHash}, nil
}

func (w *TranscriptWriter) buildRecords(hand *engine.Hand, chunkID, snapshotID string) ([]storage.TranscriptRecord, error) {
	records := make([]storage.TranscriptRecord, 0, len(hand.Events))
	for idx, event := range hand.Events {
		anchor, err := w.timeAnchorProvider.Now()
		if err != nil {
			return nil, err
		}
		payload, err := storage.EncodeCanonical(event)
		if err != nil {
			return nil, err
		}
		record := storage.TranscriptRecord{
			Version:      1,
			SessionID:    w.sessionID,
			TranscriptID: w.transcriptID,
			ChunkID:      chunkID,
			Sequence:     uint64((w.nextChunkIndex-1)<<32) + uint64(idx+1),
			HandID:       hand.ID,
			Kind:         event.EventType(),
			SnapshotID:   snapshotID,
			Message:      describeTranscriptEvent(event),
			Payload:      payload,
			TimeAnchor:   anchor,
		}
		populateTranscriptRecord(&record, hand, event)
		records = append(records, record)
	}
	return records, nil
}

func populateTranscriptRecord(record *storage.TranscriptRecord, hand *engine.Hand, event engine.Event) {
	if record == nil {
		return
	}
	switch e := event.(type) {
	case engine.HandStartedEvent:
		record.Street = engine.StreetPreflop
	case engine.BlindsPostedEvent:
		record.PlayerID = e.PlayerID
		record.AwardAmount = e.Amount
	case engine.HoleCardsDealtEvent:
		record.PlayerID = e.PlayerID
		record.ShownCards = e.Cards
	case engine.ActionTakenEvent:
		record.PlayerID = e.PlayerID
		action := e.Action
		record.Action = &action
		record.PotTotal = e.PotTotal
	case engine.StreetAdvancedEvent:
		record.Street = e.Street
		record.NewCards = append([]engine.Card(nil), e.NewCards...)
	case engine.HandRevealedEvent:
		record.PlayerID = e.PlayerID
		record.ShownCards = e.Cards
		record.EvalName = e.Eval.Name
	case engine.PotAwardedEvent:
		record.Winners = append([]engine.PlayerID(nil), e.Winners...)
		record.AwardAmount = e.Amount
	case engine.PlayerEliminatedEvent:
		record.PlayerID = e.PlayerID
	case engine.BlindLevelChangedEvent:
		record.Street = engine.StreetPreflop
	}
	if hand != nil {
		record.NewCards = append([]engine.Card(nil), record.NewCards...)
	}
}

func describeTranscriptEvent(event engine.Event) string {
	switch e := event.(type) {
	case engine.HandStartedEvent:
		return fmt.Sprintf("Hand #%d begins", e.HandID)
	case engine.BlindsPostedEvent:
		return "blind posted"
	case engine.HoleCardsDealtEvent:
		return "hole cards dealt"
	case engine.ActionTakenEvent:
		return fmt.Sprintf("action %v", e.Action.Type)
	case engine.StreetAdvancedEvent:
		return streetAdvanceMessage(e)
	case engine.ShowdownStartedEvent:
		return "Showdown"
	case engine.HandRevealedEvent:
		return e.Eval.Name
	case engine.PotAwardedEvent:
		return fmt.Sprintf("pot awarded %d", e.Amount)
	case engine.PlayerEliminatedEvent:
		return fmt.Sprintf("player %d eliminated", e.PlayerID)
	case engine.BlindLevelChangedEvent:
		return fmt.Sprintf("blinds %d/%d", e.SB, e.BB)
	case engine.TournamentFinishedEvent:
		return "Tournament finished"
	default:
		return event.EventType()
	}
}

func handResultLabel(s *Session, hand *engine.Hand) string {
	humanWon := 0
	for _, event := range hand.Events {
		if awarded, ok := event.(engine.PotAwardedEvent); ok {
			for _, winner := range awarded.Winners {
				if winner == s.HumanID {
					share := awarded.Amount
					if len(awarded.Winners) > 0 {
						share = awarded.Amount / len(awarded.Winners)
						if awarded.OddChip == s.HumanID {
							share += awarded.Amount - share*len(awarded.Winners)
						}
					}
					humanWon += share
				}
			}
		}
	}
	if humanWon > 0 {
		return fmt.Sprintf("Won %d", humanWon)
	}
	return fmt.Sprintf("Stack %d", currentHumanStack(s))
}

func collectWinningPlayerIDs(winners map[int][]engine.PlayerID) []engine.PlayerID {
	seen := map[engine.PlayerID]bool{}
	result := make([]engine.PlayerID, 0)
	for _, ids := range winners {
		for _, id := range ids {
			if !seen[id] {
				seen[id] = true
				result = append(result, id)
			}
		}
	}
	return result
}
