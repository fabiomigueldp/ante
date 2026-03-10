package storage

import (
	"encoding/json"
	"testing"
)

func TestDefaultConfigIncludesSoundVolume(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.SoundVolume != 70 {
		t.Fatalf("DefaultConfig sound volume = %d, want 70", cfg.SoundVolume)
	}
}

func TestConfigJSONRoundTripPreservesSoundVolume(t *testing.T) {
	base := DefaultConfig()
	base.SoundVolume = 40

	data, err := json.Marshal(base)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}

	decoded := DefaultConfig()
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}

	if decoded.SoundVolume != 40 {
		t.Fatalf("decoded sound volume = %d, want 40", decoded.SoundVolume)
	}
}
