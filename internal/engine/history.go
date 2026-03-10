package engine

import "time"

type PlayerSnapshot struct {
	ID    PlayerID
	Name  string
	Seat  int
	Stack int
}

type HandRecord struct {
	HandID     int
	Seed       int64
	Players    []PlayerSnapshot
	DealerSeat int
	Blinds     BlindLevel
	Board      []Card
	Actions    []Action
	Events     []Event
	Timestamp  time.Time
}

type SessionHistory struct {
	Records []HandRecord
}

func (sh *SessionHistory) Add(record HandRecord) {
	sh.Records = append(sh.Records, record)
}

func (sh *SessionHistory) Get(handID int) *HandRecord {
	for i := range sh.Records {
		if sh.Records[i].HandID == handID {
			return &sh.Records[i]
		}
	}
	return nil
}

func (sh *SessionHistory) Last(n int) []HandRecord {
	if n >= len(sh.Records) {
		out := make([]HandRecord, len(sh.Records))
		copy(out, sh.Records)
		return out
	}
	out := make([]HandRecord, n)
	copy(out, sh.Records[len(sh.Records)-n:])
	return out
}
