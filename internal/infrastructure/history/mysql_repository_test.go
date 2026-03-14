package history

import (
	"regexp"
	"testing"
	"time"

	historydomain "game-time-tracker/internal/domain/history"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMySQLRepositorySaveUpsert(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New failed: %v", err)
	}
	defer db.Close()

	repo := NewMySQLRepository(db)
	now := time.Date(2026, 3, 13, 10, 30, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectPrepare(regexp.QuoteMeta("INSERT INTO daily_history (game_name, play_date, total_time_secs, last_played_date)"))
	execExp := mock.ExpectExec(regexp.QuoteMeta("INSERT INTO daily_history (game_name, play_date, total_time_secs, last_played_date)"))
	execExp.WithArgs("PapersPlease.exe", "2026-03-13", int64(120), now)
	execExp.WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = repo.Save([]historydomain.Entry{{
		GameName:       "PapersPlease.exe",
		Date:           "2026-03-13",
		TotalTimeSecs:  120,
		LastPlayedDate: now,
	}})
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestMySQLRepositoryLoad(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New failed: %v", err)
	}
	defer db.Close()

	repo := NewMySQLRepository(db)
	lastPlayed := time.Date(2026, 3, 13, 11, 0, 0, 0, time.UTC)
	playDate := time.Date(2026, 3, 13, 0, 0, 0, 0, time.UTC)

	rows := sqlmock.NewRows([]string{"game_name", "play_date", "total_time_secs", "last_played_date"}).
		AddRow("PapersPlease.exe", playDate, int64(300), lastPlayed)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT game_name, play_date, total_time_secs, last_played_date")).
		WillReturnRows(rows)

	entries, err := repo.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].GameName != "PapersPlease.exe" {
		t.Fatalf("unexpected game name: %s", entries[0].GameName)
	}
	if entries[0].Date != "2026-03-13" {
		t.Fatalf("unexpected date: %s", entries[0].Date)
	}
	if entries[0].TotalTimeSecs != 300 {
		t.Fatalf("unexpected total secs: %d", entries[0].TotalTimeSecs)
	}
	if !entries[0].LastPlayedDate.Equal(lastPlayed) {
		t.Fatalf("unexpected last played date: %v", entries[0].LastPlayedDate)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}
