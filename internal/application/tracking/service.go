package tracking

import (
	"fmt"
	"time"

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

// Service coordinates tracking use-cases without binding to infrastructure details.
type Service struct {
	scanner     Scanner
	overlay     OverlayWriter
	stopwatches map[string]*trackingdomain.Stopwatch
}

func NewService(scanner Scanner, overlay OverlayWriter) *Service {
	return &Service{
		scanner:     scanner,
		overlay:     overlay,
		stopwatches: make(map[string]*trackingdomain.Stopwatch),
	}
}

func (s *Service) Tick() {
	found, gameName, focused := s.scanner.Scan()

	if found {
		if _, exists := s.stopwatches[gameName]; !exists {
			s.stopwatches[gameName] = &trackingdomain.Stopwatch{}
		}
		watch := s.stopwatches[gameName]

		if focused {
			watch.Start()
			s.overlay.UpdateText(fmt.Sprintf("[JOGANDO]\n%s\n%s", gameName, formatDuration(watch.Elapsed())))
			return
		}

		watch.Pause()
		s.overlay.UpdateText(fmt.Sprintf("[PAUSA]\n%s\n%s", gameName, formatDuration(watch.Elapsed())))
		return
	}

	for _, watch := range s.stopwatches {
		if watch.IsRunning() {
			watch.Pause()
		}
	}
	s.overlay.UpdateText("Aguardando Jogo...")
}

func (s *Service) PauseAll() {
	for _, watch := range s.stopwatches {
		watch.Pause()
	}
}

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}
