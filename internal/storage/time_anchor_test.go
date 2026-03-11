package storage

import (
	"testing"
	"time"
)

func TestLocalTimeAnchorProviderUsesClockAndSource(t *testing.T) {
	provider := LocalTimeAnchorProvider{
		Clock: func() time.Time {
			return time.Date(2026, time.March, 11, 12, 0, 0, 99, time.FixedZone("UTC+2", 2*60*60))
		},
		Source: "local_clock",
	}

	anchor, err := provider.Now()
	if err != nil {
		t.Fatalf("Now error: %v", err)
	}
	if anchor.Source != "local_clock" {
		t.Fatalf("source = %q, want local_clock", anchor.Source)
	}
	if anchor.Timestamp.Location() != time.UTC {
		t.Fatalf("timestamp location = %v, want UTC", anchor.Timestamp.Location())
	}
	if anchor.Timestamp.Nanosecond() != 99 {
		t.Fatalf("nanoseconds = %d, want 99", anchor.Timestamp.Nanosecond())
	}
}

func TestTimeAnchorNormalization(t *testing.T) {
	anchor := TimeAnchor{
		Timestamp: time.Date(2026, time.March, 11, 10, 0, 0, 5, time.FixedZone("X", -3*60*60)),
		Source:    "test",
	}.normalized()

	if anchor.Timestamp.Location() != time.UTC {
		t.Fatalf("location = %v, want UTC", anchor.Timestamp.Location())
	}
	if anchor.Source != "test" {
		t.Fatalf("source = %q, want test", anchor.Source)
	}
}
