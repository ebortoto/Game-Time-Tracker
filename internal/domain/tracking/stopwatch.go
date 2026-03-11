package tracking

import "time"

// Stopwatch is a domain entity that models elapsed playtime for a game session.
type Stopwatch struct {
	startedAt time.Time
	accum     time.Duration
	running   bool
}

func (s *Stopwatch) Start() {
	if s.running {
		return
	}
	s.startedAt = time.Now()
	s.running = true
}

func (s *Stopwatch) Pause() {
	if !s.running {
		return
	}
	s.accum += time.Since(s.startedAt)
	s.running = false
}

func (s *Stopwatch) Elapsed() time.Duration {
	if s.running {
		return s.accum + time.Since(s.startedAt)
	}
	return s.accum
}

func (s *Stopwatch) IsRunning() bool {
	return s.running
}
