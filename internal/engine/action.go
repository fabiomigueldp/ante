package engine

type ActionType uint8

const (
	ActionFold ActionType = iota
	ActionCheck
	ActionCall
	ActionBet
	ActionRaise
	ActionAllIn
	ActionPostBlind
	ActionPostAnte
)

type BlindType uint8

const (
	BlindSmall BlindType = iota
	BlindBig
	BlindAnte
)

type Action struct {
	PlayerID PlayerID
	Type     ActionType
	Amount   int
}

type LegalAction struct {
	Type      ActionType
	MinAmount int
	MaxAmount int
}
