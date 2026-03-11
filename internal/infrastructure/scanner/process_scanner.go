package scanner

import "game-time-tracker/internal/detector"

// ProcessScanner adapts the current detector package to the application Scanner port.
type ProcessScanner struct {
	config *detector.Config
}

func NewProcessScanner(games []string) *ProcessScanner {
	return &ProcessScanner{config: detector.New(games)}
}

func (s *ProcessScanner) Scan() (bool, string, bool) {
	return s.config.Scan()
}
