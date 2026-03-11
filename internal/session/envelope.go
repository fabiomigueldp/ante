package session

import "github.com/fabiomigueldp/ante/internal/engine"

type PromptKind uint8

const (
	PromptKindAction PromptKind = iota
	PromptKindBetweenHands
)

type ControlIntentKind uint8

const (
	ControlIntentUnknown ControlIntentKind = iota
	ControlIntentReadyNextHand
	ControlIntentLeaveTable
)

type ControlIntent struct {
	Kind ControlIntentKind
}

type Prompt struct {
	Seq          uint64
	HandID       int
	Kind         PromptKind
	PlayerID     engine.PlayerID
	View         engine.PlayerView
	LegalActions []engine.LegalAction
}

type Notice struct {
	Type      string
	Message   string
	Event     engine.Event
	PlayerID  engine.PlayerID
	BotName   string
	ThinkTime int
	Reason    string
}

type SessionError struct {
	Code    string
	Message string
}

type Envelope struct {
	Seq       uint64
	SessionID string
	HandID    int
	Snapshot  TableState
	Prompt    *Prompt
	Notice    *Notice
	Error     *SessionError
}

type PlayerActionIntent struct {
	PromptSeq uint64
	HandID    int
	Action    engine.Action
	Control   ControlIntent
}

func (e Envelope) IsTerminal() bool {
	if e.Error != nil && e.Error.Code == "session_error" {
		return true
	}
	if e.Notice == nil {
		return false
	}
	switch e.Notice.Type {
	case "tournament_finished", "session_ended":
		return true
	default:
		return false
	}
}
