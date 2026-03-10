package engine

import "fmt"

type Street uint8

const (
	StreetPreflop Street = iota
	StreetFlop
	StreetTurn
	StreetRiver
)

type HandPhase uint8

const (
	PhaseInit HandPhase = iota
	PhasePostBlinds
	PhasePreflop
	PhaseFlop
	PhaseTurn
	PhaseRiver
	PhaseShowdown
	PhaseComplete
)

type HandStepType uint8

const (
	StepNeedAction HandStepType = iota
	StepAutoAdvance
	StepComplete
)

type HandStep struct {
	Type     HandStepType
	PlayerID PlayerID
}

type Hand struct {
	ID                int
	Deck              *Deck
	Players           []*Player
	Board             []Card
	Phase             HandPhase
	Street            Street
	Betting           *BettingRound
	DealerSeat        int
	SBSeat            int
	BBSeat            int
	ActionSeat        int
	Pots              []Pot
	Events            []Event
	Actions           []Action
	StreetActions     []Action
	Winners           map[int][]PlayerID
	Blinds            BlindLevel
	SeedValue         int64
	LastAggressor     PlayerID
	ShowdownStartSeat int
}

func NewHand(id int, players []*Player, dealerSeat int, blinds BlindLevel, seed int64) *Hand {
	prepared := make([]*Player, 0, len(players))
	for _, player := range players {
		if player == nil {
			continue
		}
		player.ResetForHand()
		prepared = append(prepared, player)
	}
	deck := NewDeck(seed)
	deck.Shuffle()
	dealer := resolveDealerSeat(prepared, dealerSeat)
	sbSeat, bbSeat := blindSeats(prepared, dealer)
	hand := &Hand{
		ID:                id,
		Deck:              deck,
		Players:           prepared,
		Phase:             PhaseInit,
		Street:            StreetPreflop,
		DealerSeat:        dealer,
		SBSeat:            sbSeat,
		BBSeat:            bbSeat,
		ActionSeat:        -1,
		Winners:           make(map[int][]PlayerID),
		Blinds:            blinds,
		SeedValue:         seed,
		ShowdownStartSeat: -1,
	}
	hand.recordEvent(HandStartedEvent{HandID: id, DealerSeat: dealer, SBSeat: sbSeat, BBSeat: bbSeat, Blinds: blinds})
	return hand
}

func (h *Hand) Seed() int64 {
	return h.SeedValue
}

func (h *Hand) PostBlinds() []Event {
	if h.Phase != PhaseInit {
		return nil
	}
	var emitted []Event
	if h.Blinds.Ante > 0 {
		for _, player := range h.Players {
			if !isEligibleForDeal(player) || player.Stack == 0 {
				continue
			}
			contrib := player.Contribute(h.Blinds.Ante)
			event := BlindsPostedEvent{PlayerID: player.ID, Amount: contrib, Type: BlindAnte}
			h.recordEvent(event)
			emitted = append(emitted, event)
		}
	}
	if sb := h.playerAtSeat(h.SBSeat); sb != nil {
		contrib := sb.Contribute(h.Blinds.SB)
		event := BlindsPostedEvent{PlayerID: sb.ID, Amount: contrib, Type: BlindSmall}
		h.recordEvent(event)
		emitted = append(emitted, event)
	}
	if bb := h.playerAtSeat(h.BBSeat); bb != nil {
		contrib := bb.Contribute(h.Blinds.BB)
		event := BlindsPostedEvent{PlayerID: bb.ID, Amount: contrib, Type: BlindBig}
		h.recordEvent(event)
		emitted = append(emitted, event)
	}
	h.Phase = PhasePostBlinds
	return emitted
}

func (h *Hand) DealHoleCards() []Event {
	if h.Phase != PhasePostBlinds {
		return nil
	}
	order := dealOrder(h.Players, h.DealerSeat)
	for round := 0; round < 2; round++ {
		for _, seat := range order {
			player := h.playerAtSeat(seat)
			if player == nil || !isEligibleForDeal(player) {
				continue
			}
			player.HoleCards[round] = h.Deck.Deal()
		}
	}
	var emitted []Event
	for _, player := range h.Players {
		if player == nil || !isEligibleForDeal(player) {
			continue
		}
		event := HoleCardsDealtEvent{PlayerID: player.ID, Cards: player.HoleCards}
		h.recordEvent(event)
		emitted = append(emitted, event)
	}
	h.Betting = NewBettingRound(StreetPreflop, h.Blinds.BB, h.Blinds.BB)
	h.Phase = PhasePreflop
	h.Street = StreetPreflop
	h.ActionSeat = h.preflopFirstToActSeat()
	h.StreetActions = nil
	return emitted
}

func (h *Hand) NextStep() HandStep {
	switch h.Phase {
	case PhaseComplete:
		return HandStep{Type: StepComplete}
	case PhaseInit, PhasePostBlinds, PhaseShowdown:
		return HandStep{Type: StepAutoAdvance}
	}

	if h.onlyOneContenderLeft() {
		h.Phase = PhaseShowdown
		return HandStep{Type: StepAutoAdvance}
	}

	if h.Betting != nil && h.Betting.IsComplete(h.Players) {
		if h.Street == StreetRiver {
			h.Phase = PhaseShowdown
		}
		return HandStep{Type: StepAutoAdvance}
	}

	if h.ActionSeat == -1 {
		h.ActionSeat = h.nextSeatToAct()
		if h.ActionSeat == -1 {
			if h.Street == StreetRiver {
				h.Phase = PhaseShowdown
			}
			return HandStep{Type: StepAutoAdvance}
		}
	}

	player := h.playerAtSeat(h.ActionSeat)
	if player == nil || !player.CanAct() {
		// Player at ActionSeat can't act (e.g., went all-in from posting blind).
		// Try to find next player who can act.
		h.ActionSeat = nextActingSeat(h.Players, h.ActionSeat)
		if h.ActionSeat == -1 {
			if h.Street == StreetRiver {
				h.Phase = PhaseShowdown
			}
			return HandStep{Type: StepAutoAdvance}
		}
		player = h.playerAtSeat(h.ActionSeat)
		if player == nil {
			return HandStep{Type: StepAutoAdvance}
		}
	}
	return HandStep{Type: StepNeedAction, PlayerID: player.ID}
}

func (h *Hand) LegalActions(playerID PlayerID) []LegalAction {
	if h.Betting == nil {
		return nil
	}
	return h.Betting.LegalActions(h.playerByID(playerID))
}

func (h *Hand) ApplyAction(playerID PlayerID, action Action) ([]Event, error) {
	if h.ActionSeat == -1 {
		return nil, fmt.Errorf("no current actor")
	}
	actor := h.playerAtSeat(h.ActionSeat)
	if actor == nil || actor.ID != playerID {
		return nil, fmt.Errorf("not player %d turn", playerID)
	}
	resolved, err := h.Betting.Apply(actor, action)
	if err != nil {
		return nil, err
	}
	resolved.PlayerID = actor.ID
	h.Actions = append(h.Actions, resolved)
	h.StreetActions = append(h.StreetActions, resolved)
	if resolved.Type == ActionBet || resolved.Type == ActionRaise || (resolved.Type == ActionAllIn && actor.Bet == h.Betting.CurrentBet) {
		h.LastAggressor = actor.ID
	}
	event := ActionTakenEvent{PlayerID: actor.ID, Action: resolved, PotTotal: h.totalCommitted()}
	h.recordEvent(event)

	if h.onlyOneContenderLeft() {
		h.ActionSeat = -1
		h.Phase = PhaseShowdown
		return []Event{event}, nil
	}

	if h.Betting.IsComplete(h.Players) {
		h.ActionSeat = -1
		if h.Street == StreetRiver {
			h.Phase = PhaseShowdown
		}
		return []Event{event}, nil
	}

	h.ActionSeat = h.nextSeatToActFrom(h.ActionSeat)
	return []Event{event}, nil
}

func (h *Hand) AdvanceStreet() []Event {
	switch h.Phase {
	case PhaseInit:
		return h.PostBlinds()
	case PhasePostBlinds:
		return h.DealHoleCards()
	case PhaseShowdown:
		return h.ResolveShowdown()
	}

	if h.Street == StreetRiver {
		h.Phase = PhaseShowdown
		return nil
	}

	h.resetStreetBets()
	h.ActionSeat = -1
	h.Betting.ResetForNextStreet(h.Street + 1)
	h.Street++
	var newCards []Card
	switch h.Street {
	case StreetFlop:
		h.Phase = PhaseFlop
		h.Deck.Burn()
		newCards = h.Deck.DealN(3)
	case StreetTurn:
		h.Phase = PhaseTurn
		h.Deck.Burn()
		newCards = h.Deck.DealN(1)
	case StreetRiver:
		h.Phase = PhaseRiver
		h.Deck.Burn()
		newCards = h.Deck.DealN(1)
	}
	h.Board = append(h.Board, newCards...)
	event := StreetAdvancedEvent{Street: h.Street, NewCards: cloneCards(newCards)}
	h.recordEvent(event)
	h.ActionSeat = h.firstPostflopToAct()
	// If nobody can act (all-in / folded), mark for auto-advance
	if h.ActionSeat == -1 && h.noPlayerCanAct() {
		if h.Street == StreetRiver {
			h.Phase = PhaseShowdown
		}
	}
	return []Event{event}
}

func (h *Hand) ResolveShowdown() []Event {
	if h.Phase == PhaseComplete {
		return nil
	}
	var emitted []Event
	start := ShowdownStartedEvent{}
	h.recordEvent(start)
	emitted = append(emitted, start)
	result := ResolveShowdown(h)
	for _, playerHand := range result.Revealed {
		event := HandRevealedEvent{PlayerID: playerHand.PlayerID, Cards: playerHand.Cards, Eval: playerHand.Eval}
		h.recordEvent(event)
		emitted = append(emitted, event)
	}
	for _, pot := range result.Pots {
		h.Winners[pot.PotIndex] = append([]PlayerID(nil), pot.Winners...)
		event := PotAwardedEvent{PotIndex: pot.PotIndex, Winners: append([]PlayerID(nil), pot.Winners...), Amount: pot.Amount, OddChip: pot.OddChip}
		h.recordEvent(event)
		emitted = append(emitted, event)
	}
	h.Pots = result.RawPots
	h.Phase = PhaseComplete
	return emitted
}

func (h *Hand) ActivePlayers() []*Player {
	out := make([]*Player, 0, len(h.Players))
	for _, player := range h.Players {
		if player != nil && player.Status != StatusFolded && player.Status != StatusOut && player.Status != StatusSittingOut {
			out = append(out, player)
		}
	}
	return out
}

func (h *Hand) PlayerView(playerID PlayerID) PlayerView {
	player := h.playerByID(playerID)
	view := PlayerView{}
	if player == nil {
		return view
	}
	view.MyID = player.ID
	view.MyCards = player.HoleCards
	view.MyStack = player.Stack
	view.MyBet = player.Bet
	view.MyPosition = h.positionForSeat(player.SeatIndex)
	view.Board = cloneCards(h.Board)
	view.Street = h.Street
	view.Pot = h.totalCommitted()
	if h.Betting != nil {
		view.CurrentBet = h.Betting.CurrentBet
		view.LegalActions = h.Betting.LegalActions(player)
	}
	for _, other := range h.Players {
		if other == nil || other.ID == player.ID || !isSeated(other) {
			continue
		}
		view.Players = append(view.Players, OpponentView{ID: other.ID, Name: other.Name, Stack: other.Stack, Bet: other.Bet, Status: other.Status, Seat: other.SeatIndex})
	}
	view.Actions = cloneActions(h.Actions)
	view.NumActivePlayers = len(h.ActivePlayers())
	return view
}

func (h *Hand) playerByID(id PlayerID) *Player {
	return playerByID(h.Players, id)
}

func (h *Hand) playerAtSeat(seat int) *Player {
	return playerBySeat(h.Players, seat)
}

func (h *Hand) totalCommitted() int {
	total := 0
	for _, player := range h.Players {
		if player != nil {
			total += player.TotalBet
		}
	}
	return total
}

func (h *Hand) resetStreetBets() {
	for _, player := range h.Players {
		if player != nil {
			player.Bet = 0
		}
	}
	h.LastAggressor = 0
	h.StreetActions = nil
}

func (h *Hand) onlyOneContenderLeft() bool {
	count := 0
	for _, player := range h.Players {
		if player == nil || player.Status == StatusFolded || player.Status == StatusOut || player.Status == StatusSittingOut {
			continue
		}
		count++
		if count > 1 {
			return false
		}
	}
	return count == 1
}

func (h *Hand) positionForSeat(seat int) Position {
	active := h.activeSeats()
	if len(active) == 2 {
		switch seat {
		case h.DealerSeat:
			return PositionDealer
		case h.SBSeat:
			return PositionSmallBlind
		case h.BBSeat:
			return PositionBigBlind
		}
	}
	if seat == h.SBSeat {
		return PositionSmallBlind
	}
	if seat == h.BBSeat {
		return PositionBigBlind
	}
	order := orderedActiveSeats(h.Players, h.DealerSeat)
	index := indexOfSeat(order, seat)
	if index == -1 {
		return PositionUnknown
	}
	switch {
	case index <= 1:
		return PositionEarly
	case index >= len(order)-2:
		return PositionLate
	default:
		return PositionMiddle
	}
}

func (h *Hand) preflopFirstToActSeat() int {
	if len(h.activeSeats()) == 2 {
		// Heads-up: SB acts first preflop, but only if they can act.
		sb := h.playerAtSeat(h.SBSeat)
		if sb != nil && sb.CanAct() {
			return h.SBSeat
		}
		// SB is all-in (e.g., from posting blind). Try BB.
		return nextActingSeat(h.Players, h.SBSeat)
	}
	return nextActingSeat(h.Players, h.BBSeat)
}

func (h *Hand) firstPostflopToAct() int {
	if len(h.activeSeats()) == 2 {
		return nextActingSeat(h.Players, h.DealerSeat)
	}
	return firstActiveLeftOfDealer(h.Players, h.DealerSeat)
}

func (h *Hand) nextSeatToAct() int {
	if h.ActionSeat != -1 {
		return h.ActionSeat
	}
	if h.Street == StreetPreflop {
		return h.preflopFirstToActSeat()
	}
	return h.firstPostflopToAct()
}

func (h *Hand) nextSeatToActFrom(current int) int {
	return nextActingSeat(h.Players, current)
}

func (h *Hand) activeSeats() []int {
	seats := make([]int, 0, len(h.Players))
	for _, player := range h.Players {
		if isSeated(player) && player.Status != StatusOut && player.Status != StatusSittingOut {
			seats = append(seats, player.SeatIndex)
		}
	}
	return seats
}

func (h *Hand) recordEvent(event Event) {
	h.Events = append(h.Events, event)
}

func resolveDealerSeat(players []*Player, desired int) int {
	if desired >= 0 {
		if player := playerBySeat(players, desired); player != nil && player.Status != StatusOut && player.Status != StatusSittingOut {
			return desired
		}
	}
	return nextOccupiedSeat(players, desired)
}

func blindSeats(players []*Player, dealerSeat int) (int, int) {
	active := activeSeatCount(players)
	if active == 2 {
		sb := dealerSeat
		bb := nextOccupiedSeat(players, dealerSeat)
		return sb, bb
	}
	sb := nextOccupiedSeat(players, dealerSeat)
	bb := nextOccupiedSeat(players, sb)
	return sb, bb
}

func dealOrder(players []*Player, dealerSeat int) []int {
	return orderedActiveSeats(players, dealerSeat)
}

func orderedActiveSeats(players []*Player, startSeat int) []int {
	order := make([]int, 0, len(players))
	if len(players) == 0 {
		return order
	}
	seat := startSeat
	for range len(players) {
		seat = nextOccupiedSeat(players, seat)
		if seat == -1 || indexOfSeat(order, seat) >= 0 {
			break
		}
		order = append(order, seat)
	}
	return order
}

func nextOccupiedSeat(players []*Player, current int) int {
	if len(players) == 0 {
		return -1
	}
	max := maxSeat(players)
	if max < 0 {
		return -1
	}
	for offset := 1; offset <= max+1; offset++ {
		seat := (current + offset) % (max + 1)
		player := playerBySeat(players, seat)
		if isSeated(player) && player.Status != StatusOut && player.Status != StatusSittingOut {
			return seat
		}
	}
	return -1
}

func nextActingSeat(players []*Player, current int) int {
	if len(players) == 0 {
		return -1
	}
	max := maxSeat(players)
	if max < 0 {
		return -1
	}
	for offset := 1; offset <= max+1; offset++ {
		seat := (current + offset) % (max + 1)
		player := playerBySeat(players, seat)
		if player != nil && player.CanAct() {
			return seat
		}
	}
	return -1
}

func firstActiveLeftOfDealer(players []*Player, dealerSeat int) int {
	return nextActingSeat(players, dealerSeat)
}

func playerBySeat(players []*Player, seat int) *Player {
	for _, player := range players {
		if player != nil && player.SeatIndex == seat {
			return player
		}
	}
	return nil
}

func maxSeat(players []*Player) int {
	max := -1
	for _, player := range players {
		if player != nil && player.SeatIndex > max {
			max = player.SeatIndex
		}
	}
	return max
}

func activeSeatCount(players []*Player) int {
	count := 0
	for _, player := range players {
		if isSeated(player) && player.Status != StatusOut && player.Status != StatusSittingOut {
			count++
		}
	}
	return count
}

func isSeated(player *Player) bool {
	return player != nil
}

func isEligibleForDeal(player *Player) bool {
	return player != nil && player.Status != StatusOut && player.Status != StatusSittingOut
}

func indexOfSeat(seats []int, seat int) int {
	for i, current := range seats {
		if current == seat {
			return i
		}
	}
	return -1
}

func (h *Hand) noPlayerCanAct() bool {
	for _, p := range h.Players {
		if p != nil && p.CanAct() {
			return false
		}
	}
	return true
}
