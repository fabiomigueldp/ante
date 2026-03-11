package storage

import "time"

type TimeAnchor struct {
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
}

func (a TimeAnchor) normalized() TimeAnchor {
	return TimeAnchor{
		Timestamp: a.Timestamp.UTC().Round(0),
		Source:    a.Source,
	}
}

type TimeAnchorProvider interface {
	Now() (TimeAnchor, error)
}

type LocalTimeAnchorProvider struct {
	Clock  func() time.Time
	Source string
}

func NewLocalTimeAnchorProvider() LocalTimeAnchorProvider {
	return LocalTimeAnchorProvider{
		Clock:  time.Now,
		Source: "local_clock",
	}
}

func (p LocalTimeAnchorProvider) Now() (TimeAnchor, error) {
	clock := p.Clock
	if clock == nil {
		clock = time.Now
	}
	source := p.Source
	if source == "" {
		source = "local_clock"
	}
	return TimeAnchor{
		Timestamp: clock().UTC().Round(0),
		Source:    source,
	}.normalized(), nil
}
