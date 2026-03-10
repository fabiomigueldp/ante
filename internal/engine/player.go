package engine

type PlayerID int

type PlayerStatus uint8

const (
	StatusActive PlayerStatus = iota
	StatusFolded
	StatusAllIn
	StatusOut
	StatusSittingOut
)

type Player struct {
	ID        PlayerID
	Name      string
	Stack     int
	Bet       int
	TotalBet  int
	HoleCards [2]Card
	Status    PlayerStatus
	SeatIndex int
}

func (p *Player) Clone() *Player {
	if p == nil {
		return nil
	}
	cp := *p
	return &cp
}

func (p *Player) ResetForHand() {
	p.Bet = 0
	p.TotalBet = 0
	p.HoleCards = [2]Card{}
	if p.Status != StatusOut && p.Status != StatusSittingOut {
		p.Status = StatusActive
	}
}

func (p *Player) InHand() bool {
	return p.Status == StatusActive || p.Status == StatusAllIn
}

func (p *Player) CanAct() bool {
	return p.Status == StatusActive && p.Stack > 0
}

func (p *Player) Contribute(amount int) int {
	if amount <= 0 {
		return 0
	}
	if amount >= p.Stack {
		amount = p.Stack
	}
	p.Stack -= amount
	p.Bet += amount
	p.TotalBet += amount
	if p.Stack == 0 {
		p.Status = StatusAllIn
	}
	return amount
}
