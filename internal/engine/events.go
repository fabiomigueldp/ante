package engine

type Event interface {
	EventType() string
}

type HandStartedEvent struct {
	HandID     int
	DealerSeat int
	SBSeat     int
	BBSeat     int
	Blinds     BlindLevel
}

func (HandStartedEvent) EventType() string { return "hand_started" }

type BlindsPostedEvent struct {
	PlayerID PlayerID
	Amount   int
	Type     BlindType
}

func (BlindsPostedEvent) EventType() string { return "blind_posted" }

type HoleCardsDealtEvent struct {
	PlayerID PlayerID
	Cards    [2]Card
}

func (HoleCardsDealtEvent) EventType() string { return "hole_cards_dealt" }

type ActionTakenEvent struct {
	PlayerID PlayerID
	Action   Action
	PotTotal int
}

func (ActionTakenEvent) EventType() string { return "action_taken" }

type StreetAdvancedEvent struct {
	Street   Street
	NewCards []Card
}

func (StreetAdvancedEvent) EventType() string { return "street_advanced" }

type ShowdownStartedEvent struct{}

func (ShowdownStartedEvent) EventType() string { return "showdown_started" }

type HandRevealedEvent struct {
	PlayerID PlayerID
	Cards    [2]Card
	Eval     EvalResult
}

func (HandRevealedEvent) EventType() string { return "hand_revealed" }

type PotAwardedEvent struct {
	PotIndex int
	Winners  []PlayerID
	Amount   int
	OddChip  PlayerID
}

func (PotAwardedEvent) EventType() string { return "pot_awarded" }

type PlayerEliminatedEvent struct {
	PlayerID   PlayerID
	Position   int
	ByPlayer   PlayerID
	OnHandID   int
	FinalStack int
}

func (PlayerEliminatedEvent) EventType() string { return "player_eliminated" }

type BlindLevelChangedEvent struct {
	Level int
	SB    int
	BB    int
	Ante  int
}

func (BlindLevelChangedEvent) EventType() string { return "blind_level_changed" }

type TournamentResult struct {
	PlayerID PlayerID
	Position int
	Name     string
}

type TournamentFinishedEvent struct {
	Results []TournamentResult
}

func (TournamentFinishedEvent) EventType() string { return "tournament_finished" }
