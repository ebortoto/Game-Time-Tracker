package history

import (
	"time"

	historydomain "game-time-tracker/internal/domain/history"
	"game-time-tracker/internal/storage"
)

type JSONRepository struct {
	path string
}

func NewJSONRepository(path string) *JSONRepository {
	return &JSONRepository{path: path}
}

func (r *JSONRepository) Load() ([]historydomain.Entry, error) {
	records, err := storage.LoadHistory(r.path)
	if err != nil {
		return nil, err
	}

	entries := make([]historydomain.Entry, 0, len(records))
	for _, rec := range records {
		var lastPlayed time.Time
		if rec.LastPlayedDate != "" {
			parsed, parseErr := time.Parse(time.RFC3339, rec.LastPlayedDate)
			if parseErr == nil {
				lastPlayed = parsed
			}
		}
		entries = append(entries, historydomain.Entry{
			GameName:       rec.GameName,
			TotalTimeSecs:  rec.TotalTimeSecs,
			LastPlayedDate: lastPlayed,
		})
	}
	return entries, nil
}

func (r *JSONRepository) Save(entries []historydomain.Entry) error {
	existingRecords, err := storage.LoadHistory(r.path)
	if err != nil {
		return err
	}

	merged := make(map[string]storage.GameHistoryRecord, len(existingRecords)+len(entries))
	for _, rec := range existingRecords {
		merged[rec.GameName] = rec
	}

	for _, entry := range entries {
		current := merged[entry.GameName]
		current.GameName = entry.GameName
		current.TotalTimeSecs += entry.TotalTimeSecs
		if entry.LastPlayedDate.IsZero() {
			entry.LastPlayedDate = time.Now()
		}
		current.LastPlayedDate = entry.LastPlayedDate.Format(time.RFC3339)
		merged[entry.GameName] = current
	}

	records := make([]storage.GameHistoryRecord, 0, len(merged))
	for _, rec := range merged {
		records = append(records, rec)
	}

	return storage.SaveHistory(r.path, records)
}
