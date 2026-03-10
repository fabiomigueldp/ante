package engine

type Tournament struct {
	Table         *Table
	StartingStack int
	Eliminations  []Elimination
	HandsAtLevel  int
}

type Elimination struct {
	PlayerID PlayerID
	Position int
	HandNum  int
}

func NewTournament(table *Table, startingStack int) *Tournament {
	return &Tournament{Table: table, StartingStack: startingStack}
}

func (t *Tournament) CheckBlindIncrease() *BlindLevelChangedEvent {
	if t.Table == nil || t.Table.BlindsConfig.HandsPerLevel <= 0 {
		return nil
	}
	t.HandsAtLevel++
	if t.HandsAtLevel < t.Table.BlindsConfig.HandsPerLevel {
		return nil
	}
	t.HandsAtLevel = 0
	if t.Table.CurrentLevel < len(t.Table.BlindsConfig.Levels)-1 {
		t.Table.CurrentLevel++
	}
	level := t.Table.CurrentBlinds()
	return &BlindLevelChangedEvent{Level: level.Level, SB: level.SB, BB: level.BB, Ante: level.Ante}
}

func (t *Tournament) HandleEliminations(hand *Hand) []PlayerEliminatedEvent {
	if hand == nil || t.Table == nil {
		return nil
	}
	t.Table.ApplyHandResults(hand)
	remaining := len(t.Table.ActivePlayers())
	var events []PlayerEliminatedEvent
	for _, player := range hand.Players {
		if player == nil || player.Status == StatusOut || player.Status == StatusSittingOut || player.Stack > 0 || t.wasEliminated(player.ID) {
			continue
		}
		tablePlayer := playerByID(t.Table.Players, player.ID)
		if tablePlayer == nil {
			continue
		}
		position := remaining
		if position == 0 {
			position = 1
		}
		tablePlayer.Status = StatusOut
		t.Eliminations = append(t.Eliminations, Elimination{PlayerID: player.ID, Position: position, HandNum: hand.ID})
		event := PlayerEliminatedEvent{PlayerID: player.ID, Position: position, OnHandID: hand.ID, FinalStack: 0}
		events = append(events, event)
		remaining--
	}
	return events
}

func (t *Tournament) wasEliminated(id PlayerID) bool {
	for _, elim := range t.Eliminations {
		if elim.PlayerID == id {
			return true
		}
	}
	return false
}

func (t *Tournament) ShouldTransitionToHeadsUp() bool {
	return t.Table != nil && len(t.Table.ActivePlayers()) == 2
}

func (t *Tournament) Results() []TournamentResult {
	if t.Table == nil {
		return nil
	}
	results := make([]TournamentResult, 0, len(t.Table.Players))
	for _, elim := range t.Eliminations {
		player := playerByID(t.Table.Players, elim.PlayerID)
		name := ""
		if player != nil {
			name = player.Name
		}
		results = append(results, TournamentResult{PlayerID: elim.PlayerID, Position: elim.Position, Name: name})
	}
	for _, player := range t.Table.ActivePlayers() {
		results = append(results, TournamentResult{PlayerID: player.ID, Position: 1, Name: player.Name})
	}
	return results
}
