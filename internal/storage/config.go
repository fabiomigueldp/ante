package storage

import "path/filepath"

// Config holds all user-configurable settings.
type Config struct {
	PlayerName     string `json:"player_name"`
	SoundEnabled   bool   `json:"sound_enabled"`
	SoundVolume    int    `json:"sound_volume"`
	ShowPotOdds    bool   `json:"show_pot_odds"`
	AnimationSpeed string `json:"animation_speed"`
	DefaultMode    string `json:"default_mode"`
	DefaultSeats   int    `json:"default_seats"`
	DefaultDiff    string `json:"default_difficulty"`
	StartingStack  int    `json:"starting_stack"`
	Theme          string `json:"theme"`
}

func DefaultConfig() Config {
	return Config{
		PlayerName:     "Player",
		SoundEnabled:   true,
		SoundVolume:    70,
		ShowPotOdds:    true,
		AnimationSpeed: "normal",
		DefaultMode:    "tournament",
		DefaultSeats:   6,
		DefaultDiff:    "medium",
		StartingStack:  100,
		Theme:          "classic",
	}
}

func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func LoadConfigResult() (Config, error) {
	artifact, err := DefaultArtifactStore().LoadConfigArtifact()
	if err == nil {
		artifact.Payload.SoundVolume = clampInt(artifact.Payload.SoundVolume, 0, 100)
		return artifact.Payload, nil
	}
	if err == ErrArtifactNotFound {
		return DefaultConfig(), nil
	}
	return DefaultConfig(), err
}

func LoadConfig() Config {
	cfg, err := LoadConfigResult()
	if err != nil {
		return DefaultConfig()
	}
	return cfg
}

func SaveConfig(cfg Config) error {
	cfg.SoundVolume = clampInt(cfg.SoundVolume, 0, 100)
	_, err := DefaultArtifactStore().SaveConfigArtifact(cfg)
	return err
}

func clampInt(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
