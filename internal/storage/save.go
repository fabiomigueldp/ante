package storage

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/fabiomigueldp/ante/internal/engine"
)

// SaveSlot holds a serializable game state.
type SaveSlot struct {
	Name          string    `json:"name"`
	Timestamp     time.Time `json:"timestamp"`
	Mode          string    `json:"mode"`
	HandNumber    int       `json:"hand_number"`
	PlayerName    string    `json:"player_name"`
	PlayerStack   int       `json:"player_stack"`
	TotalPlayers  int       `json:"total_players"`
	ActivePlayers int       `json:"active_players"`
	BlindLevel    int       `json:"blind_level"`

	TableData TableSaveData             `json:"table_data"`
	Players   []PlayerSaveData          `json:"players"`
	BotSeeds  map[engine.PlayerID]int64 `json:"bot_seeds"`
	Config    GameConfig                `json:"config"`
	History   []HandRecordSave          `json:"history"`
}

type TableSaveData struct {
	Mode         engine.GameMode       `json:"mode"`
	Seats        int                   `json:"seats"`
	DealerSeat   int                   `json:"dealer_seat"`
	HandNumber   int                   `json:"hand_number"`
	CurrentLevel int                   `json:"current_level"`
	MasterSeed   int64                 `json:"master_seed"`
	BlindsConfig engine.BlindStructure `json:"blinds_config"`
}

type PlayerSaveData struct {
	ID        engine.PlayerID     `json:"id"`
	Name      string              `json:"name"`
	Stack     int                 `json:"stack"`
	Status    engine.PlayerStatus `json:"status"`
	SeatIndex int                 `json:"seat_index"`
	IsHuman   bool                `json:"is_human"`
	BotID     string              `json:"bot_id"`
}

type GameConfig struct {
	Mode           engine.GameMode `json:"mode"`
	Difficulty     int             `json:"difficulty"`
	Seats          int             `json:"seats"`
	StartingStack  int             `json:"starting_stack"`
	BlindSpeed     string          `json:"blind_speed"`
	PlayerName     string          `json:"player_name"`
	Seed           int64           `json:"seed"`
	CashGameBuyIn  int             `json:"cash_game_buy_in"`
	CashGameBlinds [2]int          `json:"cash_game_blinds"`
}

type HandRecordSave struct {
	HandID     int             `json:"hand_id"`
	DealerSeat int             `json:"dealer_seat"`
	Board      []engine.Card   `json:"board"`
	Actions    []engine.Action `json:"actions"`
	Timestamp  time.Time       `json:"timestamp"`
}

func savesDir() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "saves"), nil
}

func SaveGame(slot int, save *SaveSlot) error {
	if save == nil {
		return fmt.Errorf("save slot cannot be nil")
	}
	_, err := DefaultArtifactStore().SaveSaveArtifact(slot, *save)
	return err
}

func LoadGame(slot int) (*SaveSlot, error) {
	artifact, err := DefaultArtifactStore().LoadSaveArtifact(slot)
	if err != nil {
		return nil, err
	}
	save := artifact.Payload
	return &save, nil
}

func DeleteSave(slot int) error {
	return DefaultArtifactStore().DeleteSaveArtifact(slot)
}

func ListSavesResult() ([]SaveInfo, error) {
	return DefaultArtifactStore().ListSaveArtifacts(DefaultSaveSlotCount)
}

func ListSaves() []SaveInfo {
	saves, _ := ListSavesResult()
	return saves
}

type SaveInfo struct {
	Slot         int       `json:"slot"`
	Empty        bool      `json:"empty"`
	Name         string    `json:"name"`
	Mode         string    `json:"mode"`
	HandNum      int       `json:"hand_num"`
	Stack        int       `json:"stack"`
	Timestamp    time.Time `json:"timestamp"`
	Incompatible bool      `json:"incompatible,omitempty"`
	Error        string    `json:"error,omitempty"`
}
