package tracking

import (
	"errors"
	"strings"
	"testing"

	historydomain "game-time-tracker/internal/domain/history"
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
}

func TestSaveHistorySnapshotPropagatesSaveError(t *testing.T) {
	scanner := fakeScanner{}
	overlay := &fakeOverlay{}
	repo := &fakeHistoryRepo{shouldFail: true}
	svc := NewServiceWithHistory(scanner, overlay, repo)
	svc.scanInterval = 0

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
	svc.Tick()
	svc.Tick()

	if repo.saveCalls != 1 {
		t.Fatalf("expected one save call after game close, got: %d", repo.saveCalls)
	}
}
