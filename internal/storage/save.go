package storage

import (
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fabiomigueldp/ante/internal/engine"
)

// SaveSlot holds a serializable game state.
type SaveSlot struct {
	Name          string
	Timestamp     time.Time
	Mode          string
	HandNumber    int
	PlayerName    string
	PlayerStack   int
	TotalPlayers  int
	ActivePlayers int
	BlindLevel    int

	// Serializable game state
	TableData TableSaveData
	Players   []PlayerSaveData
	BotSeeds  map[engine.PlayerID]int64
	Config    GameConfig
	History   []HandRecordSave
}

type TableSaveData struct {
	Mode         engine.GameMode
	Seats        int
	DealerSeat   int
	HandNumber   int
	CurrentLevel int
	MasterSeed   int64
	BlindsConfig engine.BlindStructure
}

type PlayerSaveData struct {
	ID        engine.PlayerID
	Name      string
	Stack     int
	Status    engine.PlayerStatus
	SeatIndex int
	IsHuman   bool
	BotID     string // character ID for bots
}

type GameConfig struct {
	Mode           engine.GameMode
	Difficulty     int
	Seats          int
	StartingStack  int
	BlindSpeed     string
	PlayerName     string
	Seed           int64
	CashGameBuyIn  int
	CashGameBlinds [2]int
}

type HandRecordSave struct {
	HandID     int
	DealerSeat int
	Board      []engine.Card
	Actions    []engine.Action
	Timestamp  time.Time
}

func savesDir() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	d := filepath.Join(dir, "saves")
	return d, os.MkdirAll(d, 0o755)
}

func SaveGame(slot int, save *SaveSlot) error {
	dir, err := savesDir()
	if err != nil {
		return err
	}
	path := filepath.Join(dir, fmt.Sprintf("slot_%d.gob", slot))
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return gob.NewEncoder(f).Encode(save)
}

func LoadGame(slot int) (*SaveSlot, error) {
	dir, err := savesDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(dir, fmt.Sprintf("slot_%d.gob", slot))
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var save SaveSlot
	if err := gob.NewDecoder(f).Decode(&save); err != nil {
		return nil, err
	}
	return &save, nil
}

func DeleteSave(slot int) error {
	dir, err := savesDir()
	if err != nil {
		return err
	}
	path := filepath.Join(dir, fmt.Sprintf("slot_%d.gob", slot))
	return os.Remove(path)
}

func ListSaves() []SaveInfo {
	dir, err := savesDir()
	if err != nil {
		return nil
	}
	var saves []SaveInfo
	for i := 1; i <= 5; i++ {
		path := filepath.Join(dir, fmt.Sprintf("slot_%d.gob", i))
		info, err := os.Stat(path)
		if err != nil {
			saves = append(saves, SaveInfo{Slot: i, Empty: true})
			continue
		}
		slot, err := LoadGame(i)
		if err != nil {
			saves = append(saves, SaveInfo{Slot: i, Empty: true})
			continue
		}
		saves = append(saves, SaveInfo{
			Slot:      i,
			Empty:     false,
			Name:      slot.Name,
			Mode:      slot.Mode,
			HandNum:   slot.HandNumber,
			Stack:     slot.PlayerStack,
			Timestamp: info.ModTime(),
		})
	}
	return saves
}

type SaveInfo struct {
	Slot      int
	Empty     bool
	Name      string
	Mode      string
	HandNum   int
	Stack     int
	Timestamp time.Time
}
