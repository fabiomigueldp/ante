package session

import (
	"testing"

	"github.com/fabiomigueldp/ante/internal/ai"
	"github.com/fabiomigueldp/ante/internal/engine"
)

func TestDebugHeadsUpMultiHand(t *testing.T) {
	cfg := Config{
		Mode:          engine.ModeHeadsUpDuel,
		Difficulty:    ai.DifficultyEasy,
		Seats:         2,
		StartingStack: 20,
		PlayerName:    "Seed",
		Seed:          777,
	}

	sess, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	for handNum := 0; handNum < 200; handNum++ {
		hand := sess.Table.NextHand()
		if hand == nil {
			t.Logf("Table finished after %d hands", handNum)
			return
		}

		maxSteps := 200
		completed := false
		for i := 0; i < maxSteps; i++ {
			step := hand.NextStep()

			switch step.Type {
			case engine.StepComplete:
				completed = true
			case engine.StepAutoAdvance:
				hand.AdvanceStreet()
			case engine.StepNeedAction:
				view := hand.PlayerView(step.PlayerID)
				_ = view
				legal := hand.LegalActions(step.PlayerID)

				var action engine.Action
				action.PlayerID = step.PlayerID
				if step.PlayerID == sess.HumanID {
					for _, la := range legal {
						if la.Type == engine.ActionCheck {
							action.Type = engine.ActionCheck
							break
						}
						if la.Type == engine.ActionFold {
							action.Type = engine.ActionFold
							break
						}
					}
				} else {
					bot := sess.Bots[step.PlayerID]
					decision := bot.Decide(view)
					action = decision.Action
				}

				_, applyErr := hand.ApplyAction(step.PlayerID, action)
				if applyErr != nil {
					t.Fatalf("Hand %d: ApplyAction error: %v (action type=%d amount=%d legal=%v)", handNum+1, applyErr, action.Type, action.Amount, legal)
				}
			}
			if completed {
				break
			}
		}
		if !completed {
			t.Fatalf("Hand %d did not complete", handNum+1)
		}

		// Handle eliminations
		if sess.Tournament != nil {
			sess.Tournament.HandleEliminations(hand)
		}

		for _, p := range sess.Table.Players {
			if p != nil {
				t.Logf("Hand %d: %s stack=%d status=%d", handNum+1, p.Name, p.Stack, p.Status)
			}
		}

		if sess.Table.IsFinished() {
			t.Logf("Tournament done after %d hands!", handNum+1)
			return
		}
	}
	t.Fatal("Tournament did not finish after 200 hands")
}
