package ai

import (
	"math/rand"
	"strings"

	"github.com/fabiomigueldp/ante/internal/engine"
)

type SkillLevel uint8

const (
	SkillBeginner SkillLevel = iota
	SkillIntermediate
	SkillAdvanced
	SkillExpert
)

type Style string

const (
	StyleNit        Style = "nit"
	StyleTAG        Style = "tag"
	StyleLAG        Style = "lag"
	StyleStation    Style = "station"
	StyleStraight   Style = "straightforward"
	StyleManiac     Style = "maniac"
	StyleBalanced   Style = "balanced"
	StyleHeroCaller Style = "hero_caller"
	StyleTrappy     Style = "trappy"
)

type Decision struct {
	Action engine.Action
	Think  int
	Reason string
}

type Profile struct {
	Name         string
	Nickname     string
	Flavor       string
	Style        Style
	Skill        SkillLevel
	VPIP         float64
	PFR          float64
	Aggression   float64
	Bluff        float64
	CallDown     float64
	Trap         float64
	Tilt         float64
	ThinkMinMS   int
	ThinkMaxMS   int
	PositionBias float64
	DrawBias     float64
	LargeBetBias float64
	HeroCallBias float64
	ThreeBetBias float64
}

type Character struct {
	ID      string
	Profile Profile
}

type Context struct {
	View         engine.PlayerView
	RNG          *rand.Rand
	HandStrength float64
	DrawStrength float64
	Pressure     float64
	PotOdds      float64
	ToCall       int
	PotBefore    int
}

type Bot struct {
	Character Character
	RNG       *rand.Rand
	TiltLevel float64
}

func NewBot(character Character, seed int64) *Bot {
	return &Bot{Character: character, RNG: rand.New(rand.NewSource(seed))}
}

func (b *Bot) Decide(view engine.PlayerView) Decision {
	ctx := b.buildContext(view)
	legal := view.LegalActions
	if len(legal) == 0 {
		return Decision{}
	}

	decision := b.pickAction(ctx, legal)
	decision.Action.PlayerID = view.MyID
	decision.Think = b.thinkTime(decision, ctx)
	if decision.Reason == "" {
		decision.Reason = strings.ReplaceAll(string(b.Character.Profile.Style), "_", " ")
	}
	return decision
}

func (b *Bot) buildContext(view engine.PlayerView) Context {
	potBefore := view.Pot
	toCall := view.CurrentBet - view.MyBet
	if toCall < 0 {
		toCall = 0
	}
	ctx := Context{
		View:         view,
		RNG:          b.RNG,
		HandStrength: estimateStrength(view),
		DrawStrength: estimateDraws(view),
		Pressure:     tablePressure(view),
		ToCall:       toCall,
		PotBefore:    potBefore,
	}
	if toCall > 0 {
		ctx.PotOdds = float64(toCall) / float64(potBefore+toCall)
	}
	return ctx
}

func (b *Bot) pickAction(ctx Context, legal []engine.LegalAction) Decision {
	profile := b.Character.Profile
	holdings := ctx.HandStrength + ctx.DrawStrength*0.35 + b.TiltLevel*0.1
	callThreshold := 0.28 - profile.CallDown*0.12 - b.TiltLevel*0.08
	raiseThreshold := 0.62 - profile.Aggression*0.18 - profile.Bluff*0.12 - b.TiltLevel*0.15
	bluffThreshold := 0.18 + profile.Bluff*0.24 + profile.PositionBias*0.08

	if hasAction(legal, engine.ActionRaise) || hasAction(legal, engine.ActionBet) {
		if holdings >= raiseThreshold {
			return b.raiseDecision(ctx, legal, "value pressure")
		}
		if ctx.View.CurrentBet == 0 && holdings < 0.2 && ctx.RNG.Float64() < bluffThreshold {
			return b.raiseDecision(ctx, legal, "stab")
		}
		if ctx.View.CurrentBet > 0 && holdings < callThreshold && ctx.RNG.Float64() < profile.Bluff*0.12 {
			return b.raiseDecision(ctx, legal, "bluff raise")
		}
	}

	if hasAction(legal, engine.ActionCall) {
		if holdings+ctx.DrawStrength*0.25 >= maxFloat(callThreshold, ctx.PotOdds-0.04) {
			return Decision{Action: engine.Action{Type: engine.ActionCall}, Reason: "continue"}
		}
	}

	if hasAction(legal, engine.ActionCheck) {
		return Decision{Action: engine.Action{Type: engine.ActionCheck}, Reason: "check back"}
	}

	if hasAction(legal, engine.ActionAllIn) {
		if holdings >= 0.78 || (b.TiltLevel > 0.65 && ctx.RNG.Float64() < 0.4) {
			return Decision{Action: engine.Action{Type: engine.ActionAllIn}, Reason: "all in"}
		}
	}

	if hasAction(legal, engine.ActionFold) {
		return Decision{Action: engine.Action{Type: engine.ActionFold}, Reason: "fold"}
	}

	return Decision{Action: engine.Action{Type: legal[0].Type, Amount: legal[0].MinAmount}, Reason: "default"}
}

func (b *Bot) raiseDecision(ctx Context, legal []engine.LegalAction, reason string) Decision {
	var candidate *engine.LegalAction
	for _, action := range legal {
		if action.Type == engine.ActionRaise || action.Type == engine.ActionBet {
			copy := action
			candidate = &copy
			break
		}
	}
	if candidate == nil {
		if hasAction(legal, engine.ActionAllIn) {
			return Decision{Action: engine.Action{Type: engine.ActionAllIn}, Reason: reason}
		}
		if hasAction(legal, engine.ActionCall) {
			return Decision{Action: engine.Action{Type: engine.ActionCall}, Reason: reason}
		}
		return Decision{Action: engine.Action{Type: engine.ActionCheck}, Reason: reason}
	}

	profile := b.Character.Profile
	potBase := maxInt(ctx.PotBefore, ctx.View.CurrentBet)
	target := candidate.MinAmount
	if ctx.View.CurrentBet == 0 {
		fraction := 0.45 + profile.Aggression*0.45 + profile.LargeBetBias*0.2
		bet := int(float64(potBase+ctx.View.CurrentBet) * fraction)
		if bet < candidate.MinAmount {
			bet = candidate.MinAmount
		}
		if bet > candidate.MaxAmount {
			bet = candidate.MaxAmount
		}
		target = bet
	} else {
		multiplier := 2.2 + profile.Aggression*1.4 + profile.LargeBetBias*0.8
		raiseTo := ctx.View.CurrentBet + int(float64(ctx.ToCall)*multiplier)
		if raiseTo < candidate.MinAmount {
			raiseTo = candidate.MinAmount
		}
		if raiseTo > candidate.MaxAmount {
			raiseTo = candidate.MaxAmount
		}
		target = raiseTo
	}
	return Decision{Action: engine.Action{Type: candidate.Type, Amount: target}, Reason: reason}
}

func (b *Bot) thinkTime(decision Decision, ctx Context) int {
	profile := b.Character.Profile
	base := profile.ThinkMinMS
	span := profile.ThinkMaxMS - profile.ThinkMinMS
	if span < 0 {
		span = 0
	}
	complexity := 0.25
	if decision.Action.Type == engine.ActionRaise || decision.Action.Type == engine.ActionBet || decision.Action.Type == engine.ActionAllIn {
		complexity += 0.45
	}
	if ctx.View.Street >= engine.StreetTurn {
		complexity += 0.15
	}
	if ctx.ToCall > ctx.View.MyStack/3 {
		complexity += 0.15
	}
	return base + int(float64(span)*complexity)
}

func (b *Bot) ObserveBigLoss(fraction float64) {
	b.TiltLevel += fraction * b.Character.Profile.Tilt
	if b.TiltLevel > 1 {
		b.TiltLevel = 1
	}
}

func (b *Bot) CoolDown() {
	b.TiltLevel *= 0.82
}

func hasAction(legal []engine.LegalAction, actionType engine.ActionType) bool {
	for _, action := range legal {
		if action.Type == actionType {
			return true
		}
	}
	return false
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
