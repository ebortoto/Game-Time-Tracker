package tracking

import (
	"errors"
	"strings"
	"testing"
	"time"

	historydomain "game-time-tracker/internal/domain/history"
	trackingdomain "game-time-tracker/internal/domain/tracking"
)

type fakeScanner struct {
	found    bool
	gameName string
	focused  bool
}

func (f fakeScanner) Scan() (bool, string, bool) {
	return f.found, f.gameName, f.focused
}

type fakeOverlay struct {
	lastText string
}

func (f *fakeOverlay) UpdateText(text string) {
	f.lastText = text
}

type fakeHistoryRepo struct {
	saved      []historydomain.Entry
	shouldFail bool
	saveCalls  int
}

func (f *fakeHistoryRepo) Save(entries []historydomain.Entry) error {
	if f.shouldFail {
		return errors.New("save failed")
	}
	f.saveCalls++
	f.saved = append([]historydomain.Entry(nil), entries...)
	return nil
}

type sequenceScanner struct {
	results []fakeScanner
	index   int
}

func (s *sequenceScanner) Scan() (bool, string, bool) {
	if len(s.results) == 0 {
		return false, "", false
	}
	if s.index >= len(s.results) {
		last := s.results[len(s.results)-1]
		return last.found, last.gameName, last.focused
	}
	r := s.results[s.index]
	s.index++
	return r.found, r.gameName, r.focused
}

func TestServiceTickWhenFocusedShowsPlaying(t *testing.T) {
	scanner := fakeScanner{found: true, gameName: "PapersPlease.exe", focused: true}
	overlay := &fakeOverlay{}
	repo := &fakeHistoryRepo{}
	svc := NewServiceWithHistory(scanner, overlay, repo)
	svc.scanInterval = 0

	svc.Tick()

	if !strings.Contains(overlay.lastText, "[PLAYING]") {
		t.Fatalf("expected playing status, got: %q", overlay.lastText)
	}
	if !strings.Contains(overlay.lastText, "PapersPlease.exe") {
		t.Fatalf("expected game name in overlay, got: %q", overlay.lastText)
	}
}

func TestServiceTickWhenFoundButNotFocusedShowsPaused(t *testing.T) {
	scanner := fakeScanner{found: true, gameName: "PapersPlease.exe", focused: false}
	overlay := &fakeOverlay{}
	svc := NewServiceWithHistory(scanner, overlay, nil)
	svc.scanInterval = 0

	svc.Tick()

	if !strings.Contains(overlay.lastText, "[PAUSED]") {
		t.Fatalf("expected paused status, got: %q", overlay.lastText)
	}
}

func TestServiceTickWhenNoGameShowsWaiting(t *testing.T) {
	scanner := fakeScanner{found: false}
	overlay := &fakeOverlay{}
	svc := NewService(scanner, overlay)
	svc.scanInterval = 0

	svc.Tick()

	if overlay.lastText != "Waiting for game..." {
		t.Fatalf("expected waiting text, got: %q", overlay.lastText)
	}
}

func TestSaveHistorySnapshotSavesEntries(t *testing.T) {
	scanner := fakeScanner{found: true, gameName: "PapersPlease.exe", focused: true}
	overlay := &fakeOverlay{}
	repo := &fakeHistoryRepo{}
	svc := NewServiceWithHistory(scanner, overlay, repo)
	svc.scanInterval = 0

	svc.Tick()
	time.Sleep(1100 * time.Millisecond)
	svc.PauseAll()

	if err := svc.SaveHistorySnapshot(); err != nil {
		t.Fatalf("expected save to succeed, got error: %v", err)
	}

	if len(repo.saved) != 1 {
		t.Fatalf("expected 1 saved entry, got: %d", len(repo.saved))
	}
	if repo.saved[0].GameName != "PapersPlease.exe" {
		t.Fatalf("expected saved game name to be PapersPlease.exe, got: %q", repo.saved[0].GameName)
	}
	if repo.saved[0].Date == "" {
		t.Fatal("expected saved entry to include date")
	}
}

func TestSaveHistorySnapshotPropagatesSaveError(t *testing.T) {
	scanner := fakeScanner{found: true, gameName: "PapersPlease.exe", focused: true}
	overlay := &fakeOverlay{}
	repo := &fakeHistoryRepo{shouldFail: true}
	svc := NewServiceWithHistory(scanner, overlay, repo)
	svc.scanInterval = 0

	svc.Tick()
	time.Sleep(1100 * time.Millisecond)
	svc.PauseAll()

	err := svc.SaveHistorySnapshot()
	if err == nil {
		t.Fatal("expected save error, got nil")
	}
}

func TestServiceTickSavesOnceWhenGameCloses(t *testing.T) {
	scanner := &sequenceScanner{results: []fakeScanner{
		{found: true, gameName: "PapersPlease.exe", focused: true},
		{found: false, gameName: "", focused: false},
		{found: false, gameName: "", focused: false},
	}}
	overlay := &fakeOverlay{}
	repo := &fakeHistoryRepo{}
	svc := NewServiceWithHistory(scanner, overlay, repo)
	svc.scanInterval = 0

	svc.Tick()
	time.Sleep(1100 * time.Millisecond)
	svc.Tick()
	svc.Tick()

	if repo.saveCalls != 1 {
		t.Fatalf("expected one save call after game close, got: %d", repo.saveCalls)
	}
}

func TestServiceUsesTodayBaselineInOverlay(t *testing.T) {
	now := time.Now()
	today := now.Format("2006-01-02")
	scanner := fakeScanner{found: true, gameName: "PapersPlease.exe", focused: false}
	overlay := &fakeOverlay{}
	svc := NewService(scanner, overlay)
	svc.scanInterval = 0
	svc.SetInitialHistory([]historydomain.Entry{{
		GameName:      "PapersPlease.exe",
		Date:          today,
		TotalTimeSecs: 120,
	}}, now)

	svc.Tick()

	if !strings.Contains(overlay.lastText, "00:02:00") {
		t.Fatalf("expected overlay to include today's baseline 00:02:00, got: %q", overlay.lastText)
	}
}

func TestBuildHistoryEntriesUsesDeltaOnly(t *testing.T) {
	svc := NewService(fakeScanner{}, &fakeOverlay{})
	watch := svc.stopwatches["PapersPlease.exe"]
	if watch == nil {
		watch = &trackingdomain.Stopwatch{}
		svc.stopwatches["PapersPlease.exe"] = watch
	}

	watch.Start()
	time.Sleep(1200 * time.Millisecond)
	watch.Pause()

	first := svc.buildHistoryEntries(time.Now())
	second := svc.buildHistoryEntries(time.Now())

	if len(first) != 1 {
		t.Fatalf("expected first build to return one entry, got: %d", len(first))
	}
	if first[0].TotalTimeSecs <= 0 {
		t.Fatalf("expected first delta to be positive, got: %d", first[0].TotalTimeSecs)
	}
	if len(second) != 0 {
		t.Fatalf("expected second build to return no new deltas, got: %d", len(second))
	}
}

func TestSetInitialHistoryIgnoresOtherDates(t *testing.T) {
	now := time.Now()
	today := now.Format("2006-01-02")
	yesterday := now.Add(-24 * time.Hour).Format("2006-01-02")
	svc := NewService(fakeScanner{}, &fakeOverlay{})
	svc.SetInitialHistory([]historydomain.Entry{
		{GameName: "PapersPlease.exe", Date: today, TotalTimeSecs: 90},
		{GameName: "PapersPlease.exe", Date: yesterday, TotalTimeSecs: 500},
	}, now)

	if got := svc.dailyBaseSecs["PapersPlease.exe"]; got != 90 {
		t.Fatalf("expected same-day baseline 90, got: %d", got)
	}
}

func TestHandleDayRolloverPersistsOldDateAndResetsSession(t *testing.T) {
	overlay := &fakeOverlay{}
	repo := &fakeHistoryRepo{}
	svc := NewServiceWithHistory(fakeScanner{}, overlay, repo)
	oldDate := "2026-03-10"
	newDate := "2026-03-11"
	svc.activeDate = oldDate
	svc.stopwatches["PapersPlease.exe"] = &trackingdomain.Stopwatch{}

	watch := svc.stopwatches["PapersPlease.exe"]
	watch.Start()
	time.Sleep(1100 * time.Millisecond)
	watch.Pause()

	svc.handleDayRollover(time.Date(2026, 3, 11, 0, 0, 1, 0, time.Local), newDate)

	if repo.saveCalls != 1 {
		t.Fatalf("expected one save call on rollover, got: %d", repo.saveCalls)
	}
	if len(repo.saved) != 1 {
		t.Fatalf("expected one persisted rollover entry, got: %d", len(repo.saved))
	}
	if repo.saved[0].Date != oldDate {
		t.Fatalf("expected rollover save date %s, got: %s", oldDate, repo.saved[0].Date)
	}
	if svc.activeDate != newDate {
		t.Fatalf("expected activeDate switched to %s, got: %s", newDate, svc.activeDate)
	}
	if svc.persistedSecs["PapersPlease.exe"] != 0 {
		t.Fatalf("expected persisted seconds reset, got: %d", svc.persistedSecs["PapersPlease.exe"])
	}
}
