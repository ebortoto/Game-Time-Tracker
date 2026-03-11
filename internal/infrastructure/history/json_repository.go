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
		date := rec.Date
		if rec.LastPlayedDate != "" {
			parsed, parseErr := time.Parse(time.RFC3339, rec.LastPlayedDate)
			if parseErr == nil {
				lastPlayed = parsed
				if date == "" {
					date = parsed.Format("2006-01-02")
				}
			}
		}
		if date == "" {
			date = time.Now().Format("2006-01-02")
		}
		entries = append(entries, historydomain.Entry{
			GameName:       rec.GameName,
			Date:           date,
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

	keyFor := func(gameName, date string) string {
		return gameName + "|" + date
	}

	merged := make(map[string]storage.GameHistoryRecord, len(existingRecords)+len(entries))
	for _, rec := range existingRecords {
		date := rec.Date
		if date == "" && rec.LastPlayedDate != "" {
			parsed, parseErr := time.Parse(time.RFC3339, rec.LastPlayedDate)
			if parseErr == nil {
				date = parsed.Format("2006-01-02")
			}
		}
		if date == "" {
			date = time.Now().Format("2006-01-02")
		}
		rec.Date = date
		merged[keyFor(rec.GameName, rec.Date)] = rec
	}

	for _, entry := range entries {
		date := entry.Date
		if date == "" {
			if !entry.LastPlayedDate.IsZero() {
				date = entry.LastPlayedDate.Format("2006-01-02")
			} else {
				date = time.Now().Format("2006-01-02")
			}
		}
		k := keyFor(entry.GameName, date)
		current := merged[k]
		current.GameName = entry.GameName
		current.Date = date
		current.TotalTimeSecs += entry.TotalTimeSecs
		if entry.LastPlayedDate.IsZero() {
			entry.LastPlayedDate = time.Now()
		}
		current.LastPlayedDate = entry.LastPlayedDate.Format(time.RFC3339)
		merged[k] = current
	}

	records := make([]storage.GameHistoryRecord, 0, len(merged))
	for _, rec := range merged {
		records = append(records, rec)
	}

	return storage.SaveHistory(r.path, records)
}
