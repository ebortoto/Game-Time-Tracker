package tray

import (
	"errors"
	"testing"
)

func TestTUIBridgeOpenWhenClosed(t *testing.T) {
	calls := 0
	bridge := NewTUIBridge(func() error {
		calls++
		return nil
	})

	opened, err := bridge.Open()
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !opened {
		t.Fatalf("expected opened=true when bridge is closed")
	}
	if calls != 1 {
		t.Fatalf("expected opener to be called once, got %d", calls)
	}
	if !bridge.IsOpen() {
		t.Fatalf("expected bridge to be open")
	}
}

func TestTUIBridgeNoDuplicateOpen(t *testing.T) {
	calls := 0
	bridge := NewTUIBridge(func() error {
		calls++
		return nil
	})

	if _, err := bridge.Open(); err != nil {
		t.Fatalf("first open failed: %v", err)
	}
	opened, err := bridge.Open()
	if err != nil {
		t.Fatalf("second open failed: %v", err)
	}
	if opened {
		t.Fatalf("expected opened=false when already open")
	}
	if calls != 1 {
		t.Fatalf("expected opener to run only once, got %d", calls)
	}
}

func TestTUIBridgeReopenAfterClose(t *testing.T) {
	calls := 0
	bridge := NewTUIBridge(func() error {
		calls++
		return nil
	})

	if _, err := bridge.Open(); err != nil {
		t.Fatalf("first open failed: %v", err)
	}
	bridge.MarkClosed()

	opened, err := bridge.Open()
	if err != nil {
		t.Fatalf("reopen failed: %v", err)
	}
	if !opened {
		t.Fatalf("expected opened=true after close")
	}
	if calls != 2 {
		t.Fatalf("expected opener to run twice, got %d", calls)
	}
}

func TestTUIBridgeOpenError(t *testing.T) {
	bridge := NewTUIBridge(func() error {
		return errors.New("launcher failed")
	})

	opened, err := bridge.Open()
	if err == nil {
		t.Fatalf("expected open error")
	}
	if opened {
		t.Fatalf("expected opened=false on error")
	}
	if bridge.IsOpen() {
		t.Fatalf("expected bridge closed when open fails")
	}
}
