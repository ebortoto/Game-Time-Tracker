package tracking

import (
	"fmt"
	"time"

	historydomain "game-time-tracker/internal/domain/history"
	trackingdomain "game-time-tracker/internal/domain/tracking"
)

// Scanner is an application port for process/game detection.
type Scanner interface {
	Scan() (found bool, gameName string, focused bool)
}

// OverlayWriter is an application port for output rendering (RTSS/TUI/etc.).
type OverlayWriter interface {
	UpdateText(text string)
}

// HistoryRepository is an application port for playtime persistence.
type HistoryRepository interface {
	Save(entries []historydomain.Entry) error
}

// Service coordinates tracking use-cases without binding to infrastructure details.
type Service struct {
	scanner        Scanner
	overlay        OverlayWriter
	historyRepo    HistoryRepository
	stopwatches    map[string]*trackingdomain.Stopwatch
	hadGame        bool
	onHistorySaved func(entries []historydomain.Entry)

	lastScanAt   time.Time
	scanInterval time.Duration

	currentFound   bool
	currentGame    string
	currentFocused bool
}

// StatusSnapshot represents the current tracking state for external consumers.
type StatusSnapshot struct {
	Found    bool
	GameName string
	Focused  bool
	Elapsed  time.Duration
}

func NewService(scanner Scanner, overlay OverlayWriter) *Service {
	return NewServiceWithHistory(scanner, overlay, nil)
}

func NewServiceWithHistory(scanner Scanner, overlay OverlayWriter, historyRepo HistoryRepository) *Service {
	return &Service{
		scanner:      scanner,
		overlay:      overlay,
		historyRepo:  historyRepo,
		stopwatches:  make(map[string]*trackingdomain.Stopwatch),
		scanInterval: 1 * time.Second,
	}
}

func (s *Service) Tick() {
	now := time.Now()
	if s.lastScanAt.IsZero() || now.Sub(s.lastScanAt) >= s.scanInterval {
		s.scanState()
		s.lastScanAt = now
	}
	s.renderOverlay()
}

func (s *Service) scanState() {
	found, gameName, focused := s.scanner.Scan()
	s.currentFound = found
	s.currentGame = gameName
	s.currentFocused = focused

	if found {
		s.hadGame = true
		if _, exists := s.stopwatches[gameName]; !exists {
			s.stopwatches[gameName] = &trackingdomain.Stopwatch{}
		}
		return
	}

	for _, watch := range s.stopwatches {
		if watch.IsRunning() {
			watch.Pause()
		}
	}
	if s.hadGame {
		if err := s.SaveHistorySnapshot(); err != nil {
			fmt.Println("Error saving history after game close:", err)
		}
		s.hadGame = false
	}
}

func (s *Service) renderOverlay() {
	if s.currentFound {
		watch := s.stopwatches[s.currentGame]
		if s.currentFocused {
			watch.Start()
			s.overlay.UpdateText(fmt.Sprintf("[PLAYING]\n%s\n%s", s.currentGame, formatDuration(watch.Elapsed())))
			return
		}

		watch.Pause()
		s.overlay.UpdateText(fmt.Sprintf("[PAUSED]\n%s\n%s", s.currentGame, formatDuration(watch.Elapsed())))
		return
	}

	s.overlay.UpdateText("Waiting for game...")
}

func (s *Service) PauseAll() {
	for _, watch := range s.stopwatches {
		watch.Pause()
	}
}

func (s *Service) SetHistorySavedHandler(handler func(entries []historydomain.Entry)) {
	s.onHistorySaved = handler
}

func (s *Service) CurrentStatus() StatusSnapshot {
	status := StatusSnapshot{
		Found:    s.currentFound,
		GameName: s.currentGame,
		Focused:  s.currentFocused,
	}
	if !s.currentFound {
		return status
	}

	watch, ok := s.stopwatches[s.currentGame]
	if !ok {
		return status
	}
	status.Elapsed = watch.Elapsed()
	return status
}

func (s *Service) SaveHistorySnapshot() error {
	if s.historyRepo == nil {
		return nil
	}

	entries := s.buildHistoryEntries(time.Now())
	if err := s.historyRepo.Save(entries); err != nil {
		return err
	}
	if s.onHistorySaved != nil {
		s.onHistorySaved(append([]historydomain.Entry(nil), entries...))
	}
	return nil
}

func (s *Service) buildHistoryEntries(now time.Time) []historydomain.Entry {
	entries := make([]historydomain.Entry, 0, len(s.stopwatches))
	for gameName, watch := range s.stopwatches {
		entries = append(entries, historydomain.Entry{
			GameName:       gameName,
			TotalTimeSecs:  int64(watch.Elapsed() / time.Second),
			LastPlayedDate: now,
		})
	}
	return entries
}

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}
