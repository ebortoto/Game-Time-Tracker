package tray

import "sync"

// OpenTUIFunc is the launcher callback invoked when the TUI is closed and needs opening.
type OpenTUIFunc func() error

// TUIBridge prevents duplicate open requests while allowing reopen when marked closed.
type TUIBridge struct {
	mu      sync.Mutex
	isOpen  bool
	openTUI OpenTUIFunc
}

func NewTUIBridge(openTUI OpenTUIFunc) *TUIBridge {
	return &TUIBridge{openTUI: openTUI}
}

// Open opens TUI only when currently closed. It returns true when a new open was triggered.
func (b *TUIBridge) Open() (bool, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.isOpen {
		return false, nil
	}
	if b.openTUI != nil {
		if err := b.openTUI(); err != nil {
			return false, err
		}
	}
	b.isOpen = true
	return true, nil
}

func (b *TUIBridge) MarkOpen() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.isOpen = true
}

func (b *TUIBridge) MarkClosed() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.isOpen = false
}

func (b *TUIBridge) IsOpen() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.isOpen
}
