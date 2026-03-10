package engine

func playerByID(players []*Player, id PlayerID) *Player {
	for _, player := range players {
		if player != nil && player.ID == id {
			return player
		}
	}
	return nil
}

func cloneActions(actions []Action) []Action {
	if len(actions) == 0 {
		return nil
	}
	out := make([]Action, len(actions))
	copy(out, actions)
	return out
}

func cloneCards(cards []Card) []Card {
	if len(cards) == 0 {
		return nil
	}
	out := make([]Card, len(cards))
	copy(out, cards)
	return out
}
