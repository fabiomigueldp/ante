package engine

type Position uint8

const (
	PositionUnknown Position = iota
	PositionEarly
	PositionMiddle
	PositionLate
	PositionSmallBlind
	PositionBigBlind
	PositionDealer
)

type OpponentView struct {
	ID     PlayerID
	Name   string
	Stack  int
	Bet    int
	Status PlayerStatus
	Seat   int
}

type PlayerView struct {
	MyID             PlayerID
	MyCards          [2]Card
	MyStack          int
	MyBet            int
	MyPosition       Position
	Board            []Card
	Street           Street
	Pot              int
	CurrentBet       int
	NumActivePlayers int
	Players          []OpponentView
	Actions          []Action
	LegalActions     []LegalAction
}
