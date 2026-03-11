package tracking

import (
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

type fakeHistoryRepo struct{}

func (f fakeHistoryRepo) Save(_ []historydomain.Entry) error {
	return nil
}

func TestServiceTickWhenFocusedShowsPlaying(t *testing.T) {
	scanner := fakeScanner{found: true, gameName: "PapersPlease.exe", focused: true}
	overlay := &fakeOverlay{}
	svc := NewServiceWithHistory(scanner, overlay, fakeHistoryRepo{})

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

	svc.Tick()

	if !strings.Contains(overlay.lastText, "[PAUSED]") {
		t.Fatalf("expected paused status, got: %q", overlay.lastText)
	}
}

func TestServiceTickWhenNoGameShowsWaiting(t *testing.T) {
	scanner := fakeScanner{found: false}
	overlay := &fakeOverlay{}
	svc := NewService(scanner, overlay)

	svc.Tick()

	if overlay.lastText != "Waiting for game..." {
		t.Fatalf("expected waiting text, got: %q", overlay.lastText)
	}
}
