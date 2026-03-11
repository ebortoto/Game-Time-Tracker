package timer

import (
	"testing"
	"time"
)

func TestStopwatchStartPauseElapsed(t *testing.T) {
	var sw Stopwatch

	sw.Start()
	time.Sleep(20 * time.Millisecond)
	elapsedRunning := sw.Elapsed()
	if elapsedRunning <= 0 {
		t.Fatal("expected elapsed time to increase while running")
	}

	sw.Pause()
	paused := sw.Elapsed()
	if sw.Running {
		t.Fatal("expected stopwatch to be paused")
	}

	time.Sleep(20 * time.Millisecond)
	if sw.Elapsed() != paused {
		t.Fatal("expected elapsed time to stay stable while paused")
	}
}

func TestStopwatchSecondStartDoesNotReset(t *testing.T) {
	var sw Stopwatch

	sw.Start()
	time.Sleep(10 * time.Millisecond)
	before := sw.Elapsed()
	sw.Start()
	time.Sleep(10 * time.Millisecond)
	after := sw.Elapsed()

	if after < before {
		t.Fatalf("expected elapsed time to continue increasing, before=%v after=%v", before, after)
	}
}
