package tracking

import (
	"testing"
	"time"
)

func TestRuntimePublishesStatus(t *testing.T) {
	scanner := &sequenceScanner{results: []fakeScanner{
		{found: true, gameName: "PapersPlease.exe", focused: true},
	}}
	overlay := &fakeOverlay{}
	repo := &fakeHistoryRepo{}
	service := NewServiceWithHistory(scanner, overlay, repo)
	service.scanInterval = 0

	runtime := NewRuntime(service, 10*time.Millisecond)
	runtime.Start()
	defer func() {
		_ = runtime.Stop()
	}()

	statusCh := runtime.StatusUpdates()
	statusReceived := false
	timeout := time.After(400 * time.Millisecond)

	for !statusReceived {
		select {
		case st, ok := <-statusCh:
			if !ok {
				t.Fatal("status channel closed unexpectedly")
			}
			if st.State != "" {
				statusReceived = true
			}
		case <-timeout:
			t.Fatal("timed out waiting for runtime status updates")
		}
	}
}

func TestRuntimeStopPerformsGracefulSave(t *testing.T) {
	scanner := fakeScanner{found: true, gameName: "PapersPlease.exe", focused: true}
	overlay := &fakeOverlay{}
	repo := &fakeHistoryRepo{}
	service := NewServiceWithHistory(scanner, overlay, repo)
	service.scanInterval = 0

	runtime := NewRuntime(service, 10*time.Millisecond)
	runtime.Start()

	time.Sleep(1200 * time.Millisecond)
	if err := runtime.Stop(); err != nil {
		t.Fatalf("expected graceful stop without error, got: %v", err)
	}

	if repo.saveCalls == 0 {
		t.Fatal("expected stop to trigger at least one save call")
	}
}
