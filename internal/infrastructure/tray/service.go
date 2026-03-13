package tray

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type MenuAction string

const (
	MenuActionOpenTUI MenuAction = "open_tui"
	MenuActionExit    MenuAction = "exit"
)

type MenuEntry struct {
	Action MenuAction
	Label  string
}

type Handler func()

type runtimeMenuItem interface {
	ClickedChannel() <-chan struct{}
}

type runtimeAdapter interface {
	Run(onReady func(), onExit func())
	Quit()
	SetTooltip(text string)
	SetIcon(icon []byte)
	AddMenuItem(title string, tooltip string) runtimeMenuItem
}

// Service provides lifecycle hooks for tray integration.
type Service struct {
	mu        sync.Mutex
	running   bool
	entries   map[MenuAction]MenuEntry
	handlers  map[MenuAction]Handler
	runtime   runtimeAdapter
	started   atomic.Bool
	readyCh   chan struct{}
	exitCh    chan struct{}
	readyOnce sync.Once
	exitOnce  sync.Once
}

func NewService() *Service {
	return NewServiceWithRuntime(newSystrayRuntime())
}

func NewServiceWithRuntime(rt runtimeAdapter) *Service {
	return &Service{
		entries:  make(map[MenuAction]MenuEntry),
		handlers: make(map[MenuAction]Handler),
		runtime:  rt,
	}
}

func (s *Service) Start() error {
	s.mu.Lock()
	if s.started.Load() {
		s.running = true
		s.mu.Unlock()
		return nil
	}
	if len(s.entries) == 0 {
		s.entries[MenuActionOpenTUI] = MenuEntry{Action: MenuActionOpenTUI, Label: "Show"}
		s.entries[MenuActionExit] = MenuEntry{Action: MenuActionExit, Label: "Exit"}
	}
	s.readyCh = make(chan struct{})
	s.exitCh = make(chan struct{})
	s.readyOnce = sync.Once{}
	s.exitOnce = sync.Once{}
	s.started.Store(true)
	s.mu.Unlock()

	go s.runtime.Run(s.onReady, s.onExit)

	select {
	case <-s.readyCh:
		return nil
	case <-time.After(3 * time.Second):
		s.started.Store(false)
		return fmt.Errorf("tray service start timeout")
	}
}

func (s *Service) onReady() {
	s.runtime.SetTooltip("Game Time Tracker")
	if len(defaultTrayIcon) > 0 {
		s.runtime.SetIcon(defaultTrayIcon)
	}

	s.mu.Lock()
	entries := make(map[MenuAction]MenuEntry, len(s.entries))
	for action, entry := range s.entries {
		entries[action] = entry
	}
	s.running = true
	s.mu.Unlock()

	for action, entry := range entries {
		item := s.runtime.AddMenuItem(entry.Label, entry.Label)
		go s.listenMenu(action, item)
	}

	s.readyOnce.Do(func() { close(s.readyCh) })
}

func (s *Service) onExit() {
	s.mu.Lock()
	s.running = false
	s.mu.Unlock()
	s.exitOnce.Do(func() { close(s.exitCh) })
}

func (s *Service) listenMenu(action MenuAction, item runtimeMenuItem) {
	for range item.ClickedChannel() {
		_ = s.Trigger(action)
	}
}

func (s *Service) Stop() error {
	s.mu.Lock()
	if !s.started.Load() {
		s.running = false
		s.mu.Unlock()
		return nil
	}
	exitCh := s.exitCh
	s.mu.Unlock()

	s.runtime.Quit()

	select {
	case <-exitCh:
	case <-time.After(3 * time.Second):
		return fmt.Errorf("tray service stop timeout")
	}

	s.started.Store(false)
	s.mu.Lock()
	s.running = false
	s.mu.Unlock()
	return nil
}

func (s *Service) Running() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

func (s *Service) SetHandler(action MenuAction, handler Handler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[action] = handler
}

func (s *Service) HasEntry(action MenuAction) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.entries[action]
	return ok
}

func (s *Service) Trigger(action MenuAction) error {
	s.mu.Lock()
	handler := s.handlers[action]
	running := s.running
	_, hasEntry := s.entries[action]
	s.mu.Unlock()

	if !running {
		return fmt.Errorf("tray service not running")
	}
	if !hasEntry {
		return fmt.Errorf("menu action not found: %s", action)
	}
	if handler != nil {
		handler()
	}
	return nil
}
