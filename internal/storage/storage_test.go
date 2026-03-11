package storage

import (
	"path/filepath"
	"testing"
)

func TestLoadHistoryMissingFileReturnsEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.json")
	records, err := LoadHistory(path)
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if len(records) != 0 {
		t.Fatalf("expected empty records for missing file, got: %d", len(records))
	}
}

func TestSaveAndLoadHistoryRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.json")
	input := []GameHistoryRecord{
		{GameName: "PapersPlease.exe", Date: "2026-03-10", TotalTimeSecs: 120, LastPlayedDate: "2026-03-10T10:00:00Z"},
		{GameName: "CS2.exe", Date: "2026-03-10", TotalTimeSecs: 300, LastPlayedDate: "2026-03-10T11:00:00Z"},
	}

	if err := SaveHistory(path, input); err != nil {
		t.Fatalf("expected save to succeed, got: %v", err)
	}

	got, err := LoadHistory(path)
	if err != nil {
		t.Fatalf("expected load to succeed, got: %v", err)
	}
	if len(got) != len(input) {
		t.Fatalf("expected %d records, got %d", len(input), len(got))
	}

	gotMap := map[string]GameHistoryRecord{}
	for _, rec := range got {
		gotMap[rec.GameName] = rec
	}
	for _, rec := range input {
		loaded, ok := gotMap[rec.GameName]
		if !ok {
			t.Fatalf("expected record for %s", rec.GameName)
		}
		if loaded.TotalTimeSecs != rec.TotalTimeSecs {
			t.Fatalf("expected total %d for %s, got %d", rec.TotalTimeSecs, rec.GameName, loaded.TotalTimeSecs)
		}
		if loaded.LastPlayedDate != rec.LastPlayedDate {
			t.Fatalf("expected last played %s for %s, got %s", rec.LastPlayedDate, rec.GameName, loaded.LastPlayedDate)
		}
		if loaded.Date != rec.Date {
			t.Fatalf("expected date %s for %s, got %s", rec.Date, rec.GameName, loaded.Date)
		}
	}
}
