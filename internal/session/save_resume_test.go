package session

import (
	"bytes"
	"testing"
	"time"

	"github.com/fabiomigueldp/ante/internal/ai"
	"github.com/fabiomigueldp/ante/internal/engine"
	"github.com/fabiomigueldp/ante/internal/storage"
)

func TestSaveResumeRoundTripAfterFirstHandBoundary(t *testing.T) {
	store, _, _ := newSessionTestStore(t)
	useSessionDependenciesForTest(t, Dependencies{ArtifactStore: store, TimeAnchorProvider: store.TimeAnchorProvider()})
	sess, err := New(Config{
		Mode:          engine.ModeTournament,
		Difficulty:    ai.DifficultyMedium,
		Seats:         3,
		StartingStack: 50,
		PlayerName:    "Hero",
		Seed:          1234,
	})
	if err != nil {
		t.Fatalf("New error: %v", err)
	}

	playExactlyOneHand(t, sess)
	if !sess.CanSave() {
		t.Fatal("expected session to be saveable at hand boundary")
	}
	if sess.Phase != PhaseWaitingReady {
		t.Fatalf("phase = %d, want PhaseWaitingReady", sess.Phase)
	}
	sess.readyState.Snapshot.Boundary = true
	sess.readyState.Snapshot.Showdown = true
	sess.readyState.Snapshot.Revealed = []RevealedHand{{
		PlayerID: 2,
		Name:     "Bot",
		Cards:    [2]engine.Card{engine.NewCard(engine.Ace, engine.Spades), engine.NewCard(engine.Ace, engine.Hearts)},
		Eval:     "One Pair",
	}}
	sess.readyState.Snapshot.ShowdownPayouts = []ShowdownPayout{{
		PotIndex: 0,
		Winners:  []engine.PlayerID{2},
		Amount:   22,
	}}
	sess.readyState.Snapshot.PotAwards = []string{"Bot wins 22"}

	save, err := sess.BuildSaveArtifact()
	if err != nil {
		t.Fatalf("BuildSaveArtifact error: %v", err)
	}
	originalHash, err := canonicalSaveHash(*save)
	if err != nil {
		t.Fatalf("canonicalSaveHash error: %v", err)
	}
	if _, err := store.SaveSaveArtifact(1, *save); err != nil {
		t.Fatalf("SaveSaveArtifact error: %v", err)
	}

	loadedArtifact, err := store.LoadSaveArtifact(1)
	if err != nil {
		t.Fatalf("LoadSaveArtifact error: %v", err)
	}
	loadedHash, err := canonicalSaveHash(loadedArtifact.Payload)
	if err != nil {
		t.Fatalf("canonicalSaveHash loaded error: %v", err)
	}
	if !bytes.Equal(originalHash, loadedHash) {
		t.Fatal("expected original and loaded save artifact hashes to match")
	}

	resumed, err := ResumeFromSave(&loadedArtifact.Payload)
	if err != nil {
		t.Fatalf("ResumeFromSave error: %v", err)
	}

	if resumed.SessionID != sess.SessionID {
		t.Fatalf("session id = %q, want %q", resumed.SessionID, sess.SessionID)
	}
	if resumed.Table.CurrentBlinds() != sess.Table.CurrentBlinds() {
		t.Fatalf("blinds = %+v, want %+v", resumed.Table.CurrentBlinds(), sess.Table.CurrentBlinds())
	}
	if resumed.Table.HandNumber != sess.Table.HandNumber {
		t.Fatalf("hand number = %d, want %d", resumed.Table.HandNumber, sess.Table.HandNumber)
	}
	if resumed.seq != sess.seq {
		t.Fatalf("seq = %d, want %d", resumed.seq, sess.seq)
	}
	if resumed.Phase != PhaseWaitingReady {
		t.Fatalf("resumed phase = %d, want PhaseWaitingReady", resumed.Phase)
	}
	if resumed.readyState == nil || !resumed.readyState.HumanPending {
		t.Fatal("expected resumed session to restore waiting-ready boundary state")
	}
	if resumed.readyState.Snapshot.HandID != sess.readyState.Snapshot.HandID {
		t.Fatalf("ready snapshot hand id = %d, want %d", resumed.readyState.Snapshot.HandID, sess.readyState.Snapshot.HandID)
	}
	if ok := resumed.emitResumedBoundaryPrompt(); !ok {
		t.Fatal("expected resumed boundary prompt to be emitted")
	}
	env := <-resumed.Updates
	if env.Prompt == nil || env.Prompt.Kind != PromptKindBetweenHands {
		t.Fatalf("expected between-hands prompt on resume, got %+v", env.Prompt)
	}
	if env.Snapshot.HandID != resumed.readyState.Snapshot.HandID {
		t.Fatalf("resumed envelope hand id = %d, want %d", env.Snapshot.HandID, resumed.readyState.Snapshot.HandID)
	}
	if env.Snapshot.Pot != resumed.readyState.Snapshot.Pot {
		t.Fatalf("resumed envelope pot = %d, want %d", env.Snapshot.Pot, resumed.readyState.Snapshot.Pot)
	}
	if !equalCards(env.Snapshot.Board, resumed.readyState.Snapshot.Board) {
		t.Fatalf("resumed envelope board = %+v, want %+v", env.Snapshot.Board, resumed.readyState.Snapshot.Board)
	}
	if env.Snapshot.HumanCards != resumed.readyState.Snapshot.HumanCards {
		t.Fatalf("resumed envelope human cards = %+v, want %+v", env.Snapshot.HumanCards, resumed.readyState.Snapshot.HumanCards)
	}
	if !env.Snapshot.Showdown {
		t.Fatal("expected resumed envelope to preserve showdown visibility")
	}
	if len(env.Snapshot.Revealed) != 1 || env.Snapshot.Revealed[0].Name != "Bot" {
		t.Fatalf("expected resumed envelope to preserve revealed hands, got %+v", env.Snapshot.Revealed)
	}
	if len(env.Snapshot.ShowdownPayouts) != 1 || env.Snapshot.ShowdownPayouts[0].Amount != 22 {
		t.Fatalf("expected resumed envelope to preserve structured payouts, got %+v", env.Snapshot.ShowdownPayouts)
	}
	if len(env.Snapshot.PotAwards) != 1 || env.Snapshot.PotAwards[0] != "Bot wins 22" {
		t.Fatalf("expected resumed envelope to preserve pot awards, got %+v", env.Snapshot.PotAwards)
	}
	if len(resumed.Table.Players) != len(sess.Table.Players) {
		t.Fatalf("len(players) = %d, want %d", len(resumed.Table.Players), len(sess.Table.Players))
	}
	for i := range sess.Table.Players {
		got := resumed.Table.Players[i]
		want := sess.Table.Players[i]
		if got.ID != want.ID || got.Stack != want.Stack || got.Status != want.Status || got.SeatIndex != want.SeatIndex {
			t.Fatalf("player[%d] = %+v, want %+v", i, *got, *want)
		}
	}
	for playerID, bot := range sess.Bots {
		resumedBot := resumed.Bots[playerID]
		if resumedBot == nil {
			t.Fatalf("missing resumed bot %d", playerID)
		}
		if resumedBot.State() != bot.State() {
			t.Fatalf("bot state mismatch for %d: got %+v want %+v", playerID, resumedBot.State(), bot.State())
		}
	}
}

func TestBuildSaveArtifactRejectsMidHand(t *testing.T) {
	store, _, _ := newSessionTestStore(t)
	useSessionDependenciesForTest(t, Dependencies{ArtifactStore: store, TimeAnchorProvider: store.TimeAnchorProvider()})
	sess, err := New(Config{
		Mode:          engine.ModeTournament,
		Difficulty:    ai.DifficultyMedium,
		Seats:         3,
		StartingStack: 50,
		PlayerName:    "Hero",
		Seed:          42,
	})
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	sess.currentHand = sess.Table.NextHand()
	sess.HandCount = 1

	if sess.CanSave() {
		t.Fatal("expected CanSave to be false during an active hand")
	}
	if _, err := sess.BuildSaveArtifact(); err != ErrSaveMidHandNotSupported {
		t.Fatalf("BuildSaveArtifact error = %v, want ErrSaveMidHandNotSupported", err)
	}
}

func newSessionTestStore(t *testing.T) (*storage.FileSystemStore, string, storage.TimeAnchor) {
	t.Helper()
	root := t.TempDir()
	anchor := storage.TimeAnchor{Timestamp: mustTimeUTC(2026, 3, 11, 12, 0, 0), Source: "test_clock"}
	store, err := storage.NewFileSystemStore(root, staticAnchorProvider{anchor: anchor})
	if err != nil {
		t.Fatalf("NewFileSystemStore error: %v", err)
	}
	return store, root, anchor
}

func useSessionDependenciesForTest(t *testing.T, deps Dependencies) {
	t.Helper()
	old := sessionDependenciesProvider
	sessionDependenciesProvider = func() Dependencies { return deps }
	t.Cleanup(func() {
		sessionDependenciesProvider = old
	})
}

type staticAnchorProvider struct {
	anchor storage.TimeAnchor
}

func (p staticAnchorProvider) Now() (storage.TimeAnchor, error) {
	return p.anchor, nil
}

func mustTimeUTC(year int, month int, day int, hour int, minute int, second int) (ts time.Time) {
	return time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC)
}

func playExactlyOneHand(t *testing.T, sess *Session) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-done:
				return
			case env := <-sess.Updates:
				if env.Prompt != nil {
					if env.Prompt.Kind == PromptKindBetweenHands {
						sess.ActionResp <- PlayerActionIntent{PromptSeq: env.Prompt.Seq, HandID: env.Prompt.HandID, Control: ControlIntent{Kind: ControlIntentReadyNextHand}}
						continue
					}
					sess.ActionResp <- PlayerActionIntent{PromptSeq: env.Prompt.Seq, HandID: env.Prompt.HandID, Action: defaultActionForPrompt(env.Prompt)}
				}
			}
		}
	}()
	defer close(done)
	hand := sess.Table.NextHand()
	if hand == nil {
		t.Fatal("expected first hand")
	}
	sess.currentHand = hand
	sess.HandCount++
	summary, ok := sess.playHand(hand)
	if !ok {
		t.Fatal("expected hand to complete")
	}
	if sess.Tournament == nil {
		sess.Table.ApplyHandResults(hand)
	}
	if err := sess.recordHand(hand); err != nil {
		t.Fatalf("recordHand error: %v", err)
	}
	if !sess.beginBetweenHands(summary) {
		t.Fatal("expected session to enter between-hands state")
	}
	if summary.HandID != hand.ID {
		t.Fatalf("summary hand id = %d, want %d", summary.HandID, hand.ID)
	}
}
