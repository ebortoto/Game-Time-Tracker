package storage

import (
	"encoding/json"
	"fmt"
	"os"
)

// GameHistoryRecord is the persisted playtime view per game.
type GameHistoryRecord struct {
	GameName       string `json:"gameName"`
	TotalTimeSecs  int64  `json:"totalTimeSecs"`
	LastPlayedDate string `json:"lastPlayedDate"`
}

func LoadHistory(path string) ([]GameHistoryRecord, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []GameHistoryRecord{}, nil
		}
		return nil, fmt.Errorf("read history: %w", err)
	}
	if len(data) == 0 {
		return []GameHistoryRecord{}, nil
	}

	var records []GameHistoryRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("parse history: %w", err)
	}
	return records, nil
}

func SaveHistory(path string, records []GameHistoryRecord) error {
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal history: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write history: %w", err)
	}
	return nil
}
