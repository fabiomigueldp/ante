package session

import (
	"fmt"
	"strings"

	"github.com/fabiomigueldp/ante/internal/engine"
)

func ReduceGameVM(current GameVM, env Envelope) GameVM {
	if env.Seq <= current.Seq {
		return current
	}

	next := current
	next.Seq = env.Seq
	next.SessionID = env.SessionID
	next.HandID = env.HandID
	next.applySnapshot(env.Snapshot)
	next.Prompt = clonePrompt(env.Prompt)
	next.PromptKind = PromptKindAction
	next.BetweenHands = false
	next.CanSave = false
	if env.Prompt != nil {
		next.PromptKind = env.Prompt.Kind
		next.BetweenHands = env.Prompt.Kind == PromptKindBetweenHands
		next.CanSave = env.Prompt.Kind == PromptKindBetweenHands
	}
	next.LastError = ""
	if env.Prompt == nil {
		next.BotReasoning = ""
	}
	if env.Error == nil {
		next.clearMessageIfError()
	}

	if env.Notice != nil {
		next.applyNotice(*env.Notice)
	}
	if env.Error != nil {
		next.applyError(*env.Error)
	}

	return next
}

func (vm *GameVM) applySnapshot(snapshot TableState) {
	vm.Snapshot = cloneTableState(snapshot)
	vm.Players = clonePlayers(snapshot.Players)
	vm.Board = append([]engine.Card(nil), snapshot.Board...)
	vm.Pot = snapshot.Pot
	vm.Street = snapshot.Street
	vm.HandNum = snapshot.HandNum
	vm.Blinds = snapshot.Blinds
	vm.DealerSeat = snapshot.DealerSeat
	vm.HumanCards = snapshot.HumanCards
	vm.MyStack = 0
	vm.MyBet = 0
	for _, player := range snapshot.Players {
		if player.IsHuman {
			vm.MyStack = player.Stack
			vm.MyBet = player.Bet
			break
		}
	}
	if snapshot.Boundary {
		vm.Showdown = snapshot.Showdown
		vm.Revealed = cloneRevealedHands(snapshot.Revealed)
		vm.ShowdownPayouts = cloneShowdownPayouts(snapshot.ShowdownPayouts)
		vm.PotAwards = append([]string(nil), snapshot.PotAwards...)
	}
}

func (vm *GameVM) applyNotice(notice Notice) {
	switch notice.Type {
	case "session_started":
		vm.StatusLine = notice.Message
		vm.Message = ""
		vm.MessageKind = MessageKindNone
	case "hand_started":
		vm.Showdown = false
		vm.Revealed = nil
		vm.ShowdownPayouts = nil
		vm.PotAwards = nil
		vm.BetweenHands = false
		vm.PromptKind = PromptKindAction
		vm.StatusLine = ""
		vm.Message = ""
		vm.MessageKind = MessageKindNone
		vm.Finished = false
		vm.Result = ""
	case "blind_posted", "action_taken":
		vm.StatusLine = notice.Message
		vm.Message = ""
		vm.MessageKind = MessageKindNone
	case "street_advanced":
		vm.StatusLine = notice.Message
		vm.Message = ""
		vm.MessageKind = MessageKindNone
	case "showdown_started":
		vm.Showdown = true
		vm.StatusLine = notice.Message
	case "hand_revealed":
		if revealed, ok := notice.Event.(engine.HandRevealedEvent); ok {
			vm.Revealed = append(vm.Revealed, RevealedHand{
				PlayerID: revealed.PlayerID,
				Name:     playerNameFromSnapshot(vm.Snapshot, revealed.PlayerID),
				Cards:    revealed.Cards,
				Eval:     revealed.Eval.Name,
			})
		}
	case "pot_awarded":
		if awarded, ok := notice.Event.(engine.PotAwardedEvent); ok {
			vm.ShowdownPayouts = append(vm.ShowdownPayouts, ShowdownPayout{
				PotIndex: awarded.PotIndex,
				Winners:  append([]engine.PlayerID(nil), awarded.Winners...),
				Amount:   awarded.Amount,
				OddChip:  awarded.OddChip,
			})
			vm.PotAwards = append(vm.PotAwards, notice.Message)
			if len(awarded.Winners) > 0 {
				vm.StatusLine = notice.Message
			}
		}
	case "player_eliminated", "blind_level_changed":
		vm.Message = notice.Message
		vm.MessageKind = MessageKindInfo
	case "hand_complete", "hand_summary":
		vm.StatusLine = notice.Message
	case "bot_thinking":
		vm.StatusLine = notice.Message
		vm.BotReasoning = notice.Reason
	case "waiting_for_human":
		vm.StatusLine = notice.Message
	case "waiting_for_ready":
		vm.BetweenHands = true
		vm.PromptKind = PromptKindBetweenHands
		vm.StatusLine = notice.Message
	case "tournament_finished", "session_ended":
		vm.Finished = true
		vm.Result = notice.Message
		vm.Message = ""
		vm.MessageKind = MessageKindNone
		vm.StatusLine = notice.Message
	}
}

func (vm *GameVM) applyError(err SessionError) {
	vm.LastError = err.Message
	vm.Message = err.Message
	vm.MessageKind = MessageKindError
	if err.Code == "session_error" {
		vm.Finished = true
		vm.Result = err.Message
		vm.StatusLine = err.Message
	}
}

func (vm *GameVM) clearMessageIfError() {
	if vm.MessageKind == MessageKindError {
		vm.Message = ""
		vm.MessageKind = MessageKindNone
	}
}

func clonePrompt(prompt *Prompt) *Prompt {
	if prompt == nil {
		return nil
	}
	copyPrompt := *prompt
	copyPrompt.View.Board = append([]engine.Card(nil), prompt.View.Board...)
	copyPrompt.View.Players = append([]engine.OpponentView(nil), prompt.View.Players...)
	copyPrompt.View.Actions = append([]engine.Action(nil), prompt.View.Actions...)
	copyPrompt.View.LegalActions = append([]engine.LegalAction(nil), prompt.View.LegalActions...)
	copyPrompt.LegalActions = append([]engine.LegalAction(nil), prompt.LegalActions...)
	return &copyPrompt
}

func cloneTableState(snapshot TableState) TableState {
	copyState := snapshot
	copyState.Players = clonePlayers(snapshot.Players)
	copyState.Board = append([]engine.Card(nil), snapshot.Board...)
	copyState.Revealed = cloneRevealedHands(snapshot.Revealed)
	copyState.ShowdownPayouts = cloneShowdownPayouts(snapshot.ShowdownPayouts)
	copyState.PotAwards = append([]string(nil), snapshot.PotAwards...)
	return copyState
}

func clonePlayers(players []PlayerInfo) []PlayerInfo {
	cloned := make([]PlayerInfo, len(players))
	copy(cloned, players)
	return cloned
}

func cloneRevealedHands(hands []RevealedHand) []RevealedHand {
	cloned := make([]RevealedHand, len(hands))
	copy(cloned, hands)
	return cloned
}

func cloneShowdownPayouts(payouts []ShowdownPayout) []ShowdownPayout {
	cloned := make([]ShowdownPayout, len(payouts))
	for i, payout := range payouts {
		cloned[i] = payout
		cloned[i].Winners = append([]engine.PlayerID(nil), payout.Winners...)
	}
	return cloned
}

func playerNameFromSnapshot(snapshot TableState, playerID engine.PlayerID) string {
	for _, player := range snapshot.Players {
		if player.ID == playerID {
			return player.Name
		}
	}
	return fmt.Sprintf("Player %d", playerID)
}

func StatusLineForNotice(notice Notice) string {
	return strings.TrimSpace(notice.Message)
}
