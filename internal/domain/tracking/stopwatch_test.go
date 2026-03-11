package tracking

import (
	"testing"
	"time"
)

func TestStopwatchStartPauseElapsed(t *testing.T) {
	var s Stopwatch

	s.Start()
	time.Sleep(20 * time.Millisecond)
	if s.Elapsed() <= 0 {
		t.Fatal("expected elapsed time to increase after Start")
	}

	s.Pause()
	paused := s.Elapsed()
	if s.IsRunning() {
		t.Fatal("expected stopwatch to be paused")
	}

	time.Sleep(20 * time.Millisecond)
	if s.Elapsed() != paused {
		t.Fatal("expected elapsed time to remain stable while paused")
	}
}

func TestStopwatchDoubleStartDoesNotReset(t *testing.T) {
	var s Stopwatch

	s.Start()
	time.Sleep(15 * time.Millisecond)
	before := s.Elapsed()
	s.Start()
	time.Sleep(10 * time.Millisecond)
	after := s.Elapsed()

	if after < before {
		t.Fatal("expected second Start while running to not reset elapsed time")
	}
}
