package timer

import "time"

type Stopwatch struct {
	startTime time.Time
	accum     time.Duration
	Running   bool
}

// Start begins or resumes timing.
func (s *Stopwatch) Start() {
	if !s.Running {
		s.startTime = time.Now()
		s.Running = true
	}
}

// Pause stops timing and stores the elapsed duration.
func (s *Stopwatch) Pause() {
	if s.Running {
		s.accum += time.Since(s.startTime)
		s.Running = false
	}
}

// Elapsed returns total elapsed time (accumulated + current run if active).
func (s *Stopwatch) Elapsed() time.Duration {
	if s.Running {
		return s.accum + time.Since(s.startTime)
	}
	return s.accum
}
