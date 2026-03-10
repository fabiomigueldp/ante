package sim

import (
	"math/rand"

	"github.com/fabiomigueldp/ante/internal/engine"
)

type Report struct {
	HandsSimulated int
	SeedsTested    int
	Panics         int
}

func RunRandomHands(hands int, seed int64) Report {
	rng := rand.New(rand.NewSource(seed))
	report := Report{SeedsTested: hands}
	for i := 0; i < hands; i++ {
		players := []*engine.Player{
			{ID: 1, Name: "A", Stack: 200, SeatIndex: 0},
			{ID: 2, Name: "B", Stack: 200, SeatIndex: 1},
			{ID: 3, Name: "C", Stack: 200, SeatIndex: 2},
			{ID: 4, Name: "D", Stack: 200, SeatIndex: 3},
		}
		table, _ := engine.NewTable(engine.ModeTournament, 4, engine.TournamentBlinds("normal"), int64(i)+seed, players)
		hand := table.NextHand()
		if hand == nil {
			continue
		}
		safeSimulate(hand, rng, &report)
		report.HandsSimulated++
	}
	return report
}

func safeSimulate(hand *engine.Hand, rng *rand.Rand, report *Report) {
	defer func() {
		if recover() != nil {
			report.Panics++
		}
	}()
	for {
		step := hand.NextStep()
		switch step.Type {
		case engine.StepAutoAdvance:
			if hand.Phase == engine.PhaseShowdown {
				hand.ResolveShowdown()
			} else {
				hand.AdvanceStreet()
			}
		case engine.StepNeedAction:
			legal := hand.LegalActions(step.PlayerID)
			if len(legal) == 0 {
				return
			}
			choice := legal[rng.Intn(len(legal))]
			action := engine.Action{PlayerID: step.PlayerID, Type: choice.Type, Amount: choice.MinAmount}
			if choice.MaxAmount > choice.MinAmount && (choice.Type == engine.ActionBet || choice.Type == engine.ActionRaise) {
				action.Amount = choice.MinAmount + rng.Intn(choice.MaxAmount-choice.MinAmount+1)
			}
			if choice.Type == engine.ActionAllIn {
				action.Amount = 0
			}
			_, _ = hand.ApplyAction(step.PlayerID, action)
		case engine.StepComplete:
			return
		}
	}
}
