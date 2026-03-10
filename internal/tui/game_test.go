package tui

import (
	"testing"

	"github.com/fabiomigueldp/ante/internal/ai"
	"github.com/fabiomigueldp/ante/internal/audio"
	"github.com/fabiomigueldp/ante/internal/engine"
	"github.com/fabiomigueldp/ante/internal/session"
)

func updateGameFromEvent(t *testing.T, m GameModel, ev session.SessionEvent) GameModel {
	t.Helper()
	model, _ := m.handleSessionEvent(ev)
	next, ok := model.(GameModel)
	if !ok {
		t.Fatalf("expected GameModel, got %T", model)
	}
	return next
}

func TestGameHandleActionReqDoesNotReturnPointerModel(t *testing.T) {
	sess, err := session.New(session.Config{
		Mode:          engine.ModeTournament,
		Difficulty:    ai.DifficultyMedium,
		Seats:         6,
		StartingStack: 100,
		PlayerName:    "Hero",
		Seed:          42,
	})
	if err != nil {
		t.Fatalf("session.New error: %v", err)
	}

	m := NewGameModel(sess, true)
	req := session.ActionRequest{
		View: engine.PlayerView{
			MyCards:    [2]engine.Card{engine.NewCard(engine.Ace, engine.Spades), engine.NewCard(engine.King, engine.Hearts)},
			MyStack:    200,
			MyBet:      0,
			Pot:        3,
			CurrentBet: 2,
			Street:     engine.StreetPreflop,
		},
		LegalActions: []engine.LegalAction{{Type: engine.ActionFold}, {Type: engine.ActionCall, MinAmount: 2, MaxAmount: 2}},
		Snapshot: session.TableState{
			HandNum: 1,
			Blinds:  engine.BlindLevel{SB: 1, BB: 2},
			Players: []session.PlayerInfo{{ID: 1, Name: "Hero", Stack: 200, IsHuman: true}},
		},
	}

	model, _ := m.handleActionReq(req)
	if _, ok := model.(GameModel); !ok {
		t.Fatalf("expected GameModel, got %T", model)
	}
}

func TestGameActionTakenSoundsRespectMapping(t *testing.T) {
	sess, err := session.New(session.Config{
		Mode:          engine.ModeTournament,
		Difficulty:    ai.DifficultyMedium,
		Seats:         6,
		StartingStack: 100,
		PlayerName:    "Hero",
		Seed:          42,
	})
	if err != nil {
		t.Fatalf("session.New error: %v", err)
	}

	plays := make([]audio.SoundType, 0, 3)
	prev := playGameSound
	playGameSound = func(sound audio.SoundType) { plays = append(plays, sound) }
	defer func() { playGameSound = prev }()

	m := NewGameModel(sess, true)
	m = updateGameFromEvent(t, m, session.SessionEvent{Type: "action_taken", Event: engine.ActionTakenEvent{PlayerID: sess.HumanID, Action: engine.Action{Type: engine.ActionCheck}, PotTotal: 3}})
	m = updateGameFromEvent(t, m, session.SessionEvent{Type: "action_taken", Event: engine.ActionTakenEvent{PlayerID: 2, Action: engine.Action{Type: engine.ActionRaise}, PotTotal: 6}})
	m = updateGameFromEvent(t, m, session.SessionEvent{Type: "action_taken", Event: engine.ActionTakenEvent{PlayerID: 2, Action: engine.Action{Type: engine.ActionCall}, PotTotal: 6}})
	_ = updateGameFromEvent(t, m, session.SessionEvent{Type: "action_taken", Event: engine.ActionTakenEvent{PlayerID: 2, Action: engine.Action{Type: engine.ActionAllIn}, PotTotal: 20}})

	if len(plays) != 3 {
		t.Fatalf("len(plays) = %d, want 3", len(plays))
	}
	if plays[0] != audio.SoundCheck {
		t.Fatalf("first sound = %v, want SoundCheck", plays[0])
	}
	if plays[1] != audio.SoundOpponentPressure {
		t.Fatalf("second sound = %v, want SoundOpponentPressure", plays[1])
	}
	if plays[2] != audio.SoundAllIn {
		t.Fatalf("third sound = %v, want SoundAllIn", plays[2])
	}
}

func TestGameWaitingForHumanPlaysOnce(t *testing.T) {
	sess, err := session.New(session.Config{
		Mode:          engine.ModeTournament,
		Difficulty:    ai.DifficultyMedium,
		Seats:         6,
		StartingStack: 100,
		PlayerName:    "Hero",
		Seed:          42,
	})
	if err != nil {
		t.Fatalf("session.New error: %v", err)
	}

	count := 0
	prev := playGameSound
	playGameSound = func(sound audio.SoundType) {
		if sound == audio.SoundYourTurn {
			count++
		}
	}
	defer func() { playGameSound = prev }()

	m := NewGameModel(sess, true)
	m = updateGameFromEvent(t, m, session.SessionEvent{Type: "waiting_for_human", Snapshot: session.TableState{Players: []session.PlayerInfo{{ID: 1, Name: "Hero", Stack: 100, IsHuman: true}}}})
	_, _ = m.handleActionReq(session.ActionRequest{Snapshot: session.TableState{Players: []session.PlayerInfo{{ID: 1, Name: "Hero", Stack: 100, IsHuman: true}}}})

	if count != 1 {
		t.Fatalf("your turn sound count = %d, want 1", count)
	}
}

func TestGameHoleCardsDealPlaysOnceForHuman(t *testing.T) {
	sess, err := session.New(session.Config{
		Mode:          engine.ModeTournament,
		Difficulty:    ai.DifficultyMedium,
		Seats:         6,
		StartingStack: 100,
		PlayerName:    "Hero",
		Seed:          42,
	})
	if err != nil {
		t.Fatalf("session.New error: %v", err)
	}

	count := 0
	prev := playGameSound
	playGameSound = func(sound audio.SoundType) {
		if sound == audio.SoundHoleCards {
			count++
		}
	}
	defer func() { playGameSound = prev }()

	m := NewGameModel(sess, true)
	m = updateGameFromEvent(t, m, session.SessionEvent{Type: "hand_started", Event: engine.HandStartedEvent{}, Snapshot: session.TableState{Players: []session.PlayerInfo{{ID: sess.HumanID, Name: "Hero", Stack: 100, IsHuman: true}}}})
	m = updateGameFromEvent(t, m, session.SessionEvent{Type: "hole_cards_dealt", Event: engine.HoleCardsDealtEvent{PlayerID: sess.HumanID}})
	_ = updateGameFromEvent(t, m, session.SessionEvent{Type: "hole_cards_dealt", Event: engine.HoleCardsDealtEvent{PlayerID: sess.HumanID}})

	if count != 1 {
		t.Fatalf("hole cards sound count = %d, want 1", count)
	}
}

func TestGameStreetAdvanceDifferentiatesFlopAndTurnRiver(t *testing.T) {
	sess, err := session.New(session.Config{
		Mode:          engine.ModeTournament,
		Difficulty:    ai.DifficultyMedium,
		Seats:         6,
		StartingStack: 100,
		PlayerName:    "Hero",
		Seed:          42,
	})
	if err != nil {
		t.Fatalf("session.New error: %v", err)
	}

	plays := []audio.SoundType{}
	prev := playGameSound
	playGameSound = func(sound audio.SoundType) { plays = append(plays, sound) }
	defer func() { playGameSound = prev }()

	m := NewGameModel(sess, true)
	m = updateGameFromEvent(t, m, session.SessionEvent{Type: "street_advanced", Event: engine.StreetAdvancedEvent{Street: engine.StreetFlop, NewCards: []engine.Card{engine.NewCard(engine.Ace, engine.Spades), engine.NewCard(engine.King, engine.Hearts), engine.NewCard(engine.Queen, engine.Diamonds)}}})
	_ = updateGameFromEvent(t, m, session.SessionEvent{Type: "street_advanced", Event: engine.StreetAdvancedEvent{Street: engine.StreetTurn, NewCards: []engine.Card{engine.NewCard(engine.Jack, engine.Clubs)}}})

	if len(plays) != 2 {
		t.Fatalf("len(plays) = %d, want 2", len(plays))
	}
	if plays[0] != audio.SoundFlop {
		t.Fatalf("first street sound = %v, want SoundFlop", plays[0])
	}
	if plays[1] != audio.SoundTurnRiver {
		t.Fatalf("second street sound = %v, want SoundTurnRiver", plays[1])
	}
}

func TestGameShowdownAndBustoutAndEndUsePremiumCues(t *testing.T) {
	sess, err := session.New(session.Config{
		Mode:          engine.ModeTournament,
		Difficulty:    ai.DifficultyMedium,
		Seats:         6,
		StartingStack: 100,
		PlayerName:    "Hero",
		Seed:          42,
	})
	if err != nil {
		t.Fatalf("session.New error: %v", err)
	}

	plays := []audio.SoundType{}
	prev := playGameSound
	playGameSound = func(sound audio.SoundType) { plays = append(plays, sound) }
	defer func() { playGameSound = prev }()

	m := NewGameModel(sess, true)
	m = updateGameFromEvent(t, m, session.SessionEvent{Type: "showdown_started"})
	m = updateGameFromEvent(t, m, session.SessionEvent{Type: "player_eliminated", Event: engine.PlayerEliminatedEvent{PlayerID: sess.HumanID, Position: 2}})
	_ = updateGameFromEvent(t, m, session.SessionEvent{Type: "tournament_finished", Event: engine.TournamentFinishedEvent{Results: []engine.TournamentResult{{PlayerID: sess.HumanID, Position: 1, Name: "Hero"}}}})

	if len(plays) != 3 {
		t.Fatalf("len(plays) = %d, want 3", len(plays))
	}
	if plays[0] != audio.SoundShowdown {
		t.Fatalf("first sound = %v, want SoundShowdown", plays[0])
	}
	if plays[1] != audio.SoundBustout {
		t.Fatalf("second sound = %v, want SoundBustout", plays[1])
	}
	if plays[2] != audio.SoundVictory {
		t.Fatalf("third sound = %v, want SoundVictory", plays[2])
	}
}

func TestGameCashSessionEndCanUseDefeatCue(t *testing.T) {
	sess, err := session.New(session.Config{
		Mode:           engine.ModeCashGame,
		Difficulty:     ai.DifficultyMedium,
		Seats:          6,
		StartingStack:  100,
		CashGameBuyIn:  1000,
		CashGameBlinds: [2]int{5, 10},
		PlayerName:     "Hero",
		Seed:           42,
	})
	if err != nil {
		t.Fatalf("session.New error: %v", err)
	}

	plays := []audio.SoundType{}
	prev := playGameSound
	playGameSound = func(sound audio.SoundType) { plays = append(plays, sound) }
	defer func() { playGameSound = prev }()

	m := NewGameModel(sess, true)
	m.players = []session.PlayerInfo{{ID: sess.HumanID, Name: "Hero", Stack: 800, IsHuman: true}}
	_ = updateGameFromEvent(t, m, session.SessionEvent{Type: "session_ended", Message: "Session over."})

	if len(plays) != 1 {
		t.Fatalf("len(plays) = %d, want 1", len(plays))
	}
	if plays[0] != audio.SoundDefeat {
		t.Fatalf("sound = %v, want SoundDefeat", plays[0])
	}
}
