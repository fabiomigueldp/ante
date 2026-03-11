package session

import "github.com/fabiomigueldp/ante/internal/engine"

type MessageKind string

const (
	MessageKindNone  MessageKind = ""
	MessageKindInfo  MessageKind = "info"
	MessageKindError MessageKind = "error"
)

type RevealedHand struct {
	PlayerID engine.PlayerID
	Name     string
	Cards    [2]engine.Card
	Eval     string
}

type GameVM struct {
	Seq          uint64
	SessionID    string
	HandID       int
	Snapshot     TableState
	Players      []PlayerInfo
	Board        []engine.Card
	Pot          int
	Street       engine.Street
	HandNum      int
	Blinds       engine.BlindLevel
	DealerSeat   int
	HumanCards   [2]engine.Card
	MyStack      int
	MyBet        int
	Prompt       *Prompt
	StatusLine   string
	Message      string
	MessageKind  MessageKind
	Showdown     bool
	Revealed     []RevealedHand
	PotAwards    []string
	Finished     bool
	Result       string
	LastError    string
	BotReasoning string
}

func (vm GameVM) HasPrompt() bool {
	return vm.Prompt != nil
}

func BootstrapGameVM(sessionID string, snapshot TableState) GameVM {
	vm := GameVM{SessionID: sessionID}
	vm.applySnapshot(snapshot)
	return vm
}
