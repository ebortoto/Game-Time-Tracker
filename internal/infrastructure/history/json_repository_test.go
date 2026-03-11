package history

import (
	"path/filepath"
	"testing"
	"time"

	historydomain "game-time-tracker/internal/domain/history"
)

func TestJSONRepositorySaveMergesWithExistingTotals(t *testing.T) {
	path := filepath.Join(t.TempDir(), "playtime_history.json")
	repo := NewJSONRepository(path)

	baseTime := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	if err := repo.Save([]historydomain.Entry{{
		GameName:       "PapersPlease.exe",
		Date:           "2026-03-10",
		TotalTimeSecs:  120,
		LastPlayedDate: baseTime,
	}}); err != nil {
		t.Fatalf("seed save failed: %v", err)
	}

	if err := repo.Save([]historydomain.Entry{{
		GameName:       "PapersPlease.exe",
		Date:           "2026-03-10",
		TotalTimeSecs:  30,
		LastPlayedDate: baseTime.Add(10 * time.Minute),
	}, {
		GameName:       "CS2.exe",
		Date:           "2026-03-10",
		TotalTimeSecs:  45,
		LastPlayedDate: baseTime.Add(20 * time.Minute),
	}}); err != nil {
		t.Fatalf("merge save failed: %v", err)
	}

	entries, err := repo.Load()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 daily entries after merge, got %d", len(entries))
	}

	byGame := map[string]historydomain.Entry{}
	for _, entry := range entries {
		byGame[entry.GameName] = entry
	}

	if byGame["PapersPlease.exe"].TotalTimeSecs != 150 {
		t.Fatalf("expected PapersPlease total 150, got %d", byGame["PapersPlease.exe"].TotalTimeSecs)
	}
	if byGame["PapersPlease.exe"].Date != "2026-03-10" {
		t.Fatalf("expected PapersPlease date 2026-03-10, got %s", byGame["PapersPlease.exe"].Date)
	}
	if byGame["CS2.exe"].TotalTimeSecs != 45 {
		t.Fatalf("expected CS2 total 45, got %d", byGame["CS2.exe"].TotalTimeSecs)
	}
	if byGame["PapersPlease.exe"].LastPlayedDate.Before(baseTime.Add(10 * time.Minute)) {
		t.Fatal("expected PapersPlease last played date to be updated")
	}
}
