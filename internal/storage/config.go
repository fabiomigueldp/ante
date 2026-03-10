package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds all user-configurable settings.
type Config struct {
	PlayerName     string `json:"player_name"`
	SoundEnabled   bool   `json:"sound_enabled"`
	SoundVolume    int    `json:"sound_volume"`
	ShowPotOdds    bool   `json:"show_pot_odds"`
	AnimationSpeed string `json:"animation_speed"` // "slow", "normal", "fast", "off"
	DefaultMode    string `json:"default_mode"`    // "tournament", "cash", "headsup"
	DefaultSeats   int    `json:"default_seats"`
	DefaultDiff    string `json:"default_difficulty"` // "easy", "medium", "hard"
	StartingStack  int    `json:"starting_stack"`     // in BB
	Theme          string `json:"theme"`              // "classic", "dark", "green"
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

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".ante")
	return dir, os.MkdirAll(dir, 0o755)
}

func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func LoadConfig() Config {
	path, err := configPath()
	if err != nil {
		return DefaultConfig()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultConfig()
	}
	cfg := DefaultConfig()
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig()
	}
	cfg.SoundVolume = clampInt(cfg.SoundVolume, 0, 100)
	return cfg
}

func SaveConfig(cfg Config) error {
	cfg.SoundVolume = clampInt(cfg.SoundVolume, 0, 100)
	path, err := configPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
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
