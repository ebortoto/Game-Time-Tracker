package history

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	historydomain "game-time-tracker/internal/domain/history"
)

type MySQLRepository struct {
	db *sql.DB
}

func NewMySQLRepository(db *sql.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

func (r *MySQLRepository) Load() ([]historydomain.Entry, error) {
	const query = `
		SELECT game_name, play_date, total_time_secs, last_played_date
		FROM daily_history
		ORDER BY game_name, play_date
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query history: %w", err)
	}
	defer rows.Close()

	entries := make([]historydomain.Entry, 0)
	for rows.Next() {
		var gameName string
		var playDate time.Time
		var totalSecs int64
		var lastPlayed time.Time
		if err := rows.Scan(&gameName, &playDate, &totalSecs, &lastPlayed); err != nil {
			return nil, fmt.Errorf("scan history row: %w", err)
		}
		entries = append(entries, historydomain.Entry{
			GameName:       gameName,
			Date:           playDate.Format("2006-01-02"),
			TotalTimeSecs:  totalSecs,
			LastPlayedDate: lastPlayed,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate history rows: %w", err)
	}
	return entries, nil
}

func (r *MySQLRepository) Save(entries []historydomain.Entry) error {
	if len(entries) == 0 {
		return nil
	}
	ctx := context.Background()
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	const upsert = `
		INSERT INTO daily_history (game_name, play_date, total_time_secs, last_played_date)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			total_time_secs = total_time_secs + VALUES(total_time_secs),
			last_played_date = GREATEST(last_played_date, VALUES(last_played_date))
	`
	stmt, err := tx.PrepareContext(ctx, upsert)
	if err != nil {
		return fmt.Errorf("prepare upsert: %w", err)
	}
	defer stmt.Close()

	now := time.Now()
	for _, entry := range entries {
		date := entry.Date
		if date == "" {
			if !entry.LastPlayedDate.IsZero() {
				date = entry.LastPlayedDate.Format("2006-01-02")
			} else {
				date = now.Format("2006-01-02")
			}
		}
		lastPlayed := entry.LastPlayedDate
		if lastPlayed.IsZero() {
			lastPlayed = now
		}

		if _, err := stmt.ExecContext(ctx, entry.GameName, date, entry.TotalTimeSecs, lastPlayed); err != nil {
			return fmt.Errorf("upsert entry %s/%s: %w", entry.GameName, date, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	tx = nil
	return nil
}
