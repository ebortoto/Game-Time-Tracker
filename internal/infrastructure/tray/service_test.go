package tray

import (
	"sync"
	"testing"
	"time"
)

type fakeMenuItem struct {
	clicked chan struct{}
}

func newFakeMenuItem() *fakeMenuItem {
	return &fakeMenuItem{clicked: make(chan struct{}, 4)}
}

func (m *fakeMenuItem) ClickedChannel() <-chan struct{} {
	return m.clicked
}

type fakeRuntime struct {
	mu       sync.Mutex
	onReady  func()
	onExit   func()
	items    []*fakeMenuItem
	quitting bool
}

func (f *fakeRuntime) Run(onReady func(), onExit func()) {
	f.mu.Lock()
	f.onReady = onReady
	f.onExit = onExit
	f.mu.Unlock()
	onReady()
	for {
		f.mu.Lock()
		if f.quitting {
			f.mu.Unlock()
			onExit()
			return
		}
		f.mu.Unlock()
		time.Sleep(5 * time.Millisecond)
	}
}

func (f *fakeRuntime) Quit() {
	f.mu.Lock()
	f.quitting = true
	f.mu.Unlock()
}

func (f *fakeRuntime) SetTooltip(text string) {}

func (f *fakeRuntime) SetIcon(icon []byte) {}

func (f *fakeRuntime) AddMenuItem(title string, tooltip string) runtimeMenuItem {
	f.mu.Lock()
	defer f.mu.Unlock()
	item := newFakeMenuItem()
	f.items = append(f.items, item)
	return item
}

func (f *fakeRuntime) click(index int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if index >= 0 && index < len(f.items) {
		f.items[index].clicked <- struct{}{}
	}
}

func TestServiceLifecycle(t *testing.T) {
	rt := &fakeRuntime{}
	svc := NewServiceWithRuntime(rt)

	if svc.Running() {
		t.Fatalf("expected service to start stopped")
	}

	if err := svc.Start(); err != nil {
		t.Fatalf("start failed: %v", err)
	}
	if !svc.Running() {
		t.Fatalf("expected service running after Start")
	}
	if !svc.HasEntry(MenuActionOpenTUI) {
		t.Fatalf("expected Open TUI entry to exist")
	}
	if !svc.HasEntry(MenuActionExit) {
		t.Fatalf("expected Exit entry to exist")
	}

	if err := svc.Stop(); err != nil {
		t.Fatalf("stop failed: %v", err)
	}
	if svc.Running() {
		t.Fatalf("expected service stopped after Stop")
	}
}

func TestServiceLifecycleIdempotent(t *testing.T) {
	rt := &fakeRuntime{}
	svc := NewServiceWithRuntime(rt)

	if err := svc.Start(); err != nil {
		t.Fatalf("first start failed: %v", err)
	}
	if err := svc.Start(); err != nil {
		t.Fatalf("second start failed: %v", err)
	}

	if err := svc.Stop(); err != nil {
		t.Fatalf("first stop failed: %v", err)
	}
	if err := svc.Stop(); err != nil {
		t.Fatalf("second stop failed: %v", err)
	}
}

func TestServiceTriggerHandlers(t *testing.T) {
	rt := &fakeRuntime{}
	svc := NewServiceWithRuntime(rt)
	if err := svc.Start(); err != nil {
		t.Fatalf("start failed: %v", err)
	}

	openCalls := 0
	exitCalls := 0
	svc.SetHandler(MenuActionOpenTUI, func() { openCalls++ })
	svc.SetHandler(MenuActionExit, func() { exitCalls++ })

	if err := svc.Trigger(MenuActionOpenTUI); err != nil {
		t.Fatalf("open trigger failed: %v", err)
	}
	if err := svc.Trigger(MenuActionExit); err != nil {
		t.Fatalf("exit trigger failed: %v", err)
	}

	if openCalls != 1 {
		t.Fatalf("expected open handler 1 call, got %d", openCalls)
	}
	if exitCalls != 1 {
		t.Fatalf("expected exit handler 1 call, got %d", exitCalls)
	}
}

func TestServiceMenuClickInvokesHandlers(t *testing.T) {
	rt := &fakeRuntime{}
	svc := NewServiceWithRuntime(rt)

	if err := svc.Start(); err != nil {
		t.Fatalf("start failed: %v", err)
	}
	t.Cleanup(func() { _ = svc.Stop() })

	openCalls := 0
	exitCalls := 0
	svc.SetHandler(MenuActionOpenTUI, func() { openCalls++ })
	svc.SetHandler(MenuActionExit, func() { exitCalls++ })

	rt.click(0)
	rt.click(1)
	time.Sleep(30 * time.Millisecond)

	if openCalls != 1 || exitCalls != 1 {
		t.Fatalf("expected both handlers to be called once, got open=%d exit=%d", openCalls, exitCalls)
	}
}

func TestServiceTriggerFailsWhenStopped(t *testing.T) {
	rt := &fakeRuntime{}
	svc := NewServiceWithRuntime(rt)
	if err := svc.Trigger(MenuActionOpenTUI); err == nil {
		t.Fatalf("expected error when triggering while stopped")
	}
}
