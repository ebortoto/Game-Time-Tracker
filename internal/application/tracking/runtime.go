package tracking

import (
	"fmt"
	"log/slog"
	"time"

	historydomain "game-time-tracker/internal/domain/history"
)

// RuntimeStatus is the channel payload used by TUI/console layers.
type RuntimeStatus struct {
	State    string
	GameName string
	Elapsed  time.Duration
	Updated  time.Time
}

// Runtime owns the tracker loop in a single goroutine and publishes updates through channels.
type Runtime struct {
	service      *Service
	tickInterval time.Duration

	statusCh  chan RuntimeStatus
	historyCh chan []historydomain.Entry
	errCh     chan error
	stopReqCh chan chan error

	lastLoggedState string
	lastLoggedGame  string
}

func NewRuntime(service *Service, tickInterval time.Duration) *Runtime {
	if tickInterval <= 0 {
		tickInterval = 200 * time.Millisecond
	}
	return &Runtime{
		service:      service,
		tickInterval: tickInterval,
		statusCh:     make(chan RuntimeStatus, 1),
		historyCh:    make(chan []historydomain.Entry, 1),
		errCh:        make(chan error, 1),
		stopReqCh:    make(chan chan error, 1),
	}
}

func (r *Runtime) Start() {
	slog.Info("runtime_started", "tick_interval_ms", r.tickInterval.Milliseconds())
	r.service.SetHistorySavedHandler(func(entries []historydomain.Entry) {
		r.publishHistory(entries)
	})

	go func() {
		ticker := time.NewTicker(r.tickInterval)
		defer ticker.Stop()
		defer close(r.statusCh)
		defer close(r.historyCh)
		defer close(r.errCh)

		for {
			select {
			case <-ticker.C:
				r.service.Tick()
				r.publishStatus(r.service.CurrentStatus())
			case resp := <-r.stopReqCh:
				slog.Info("runtime_stopping")
				r.service.PauseAll()
				err := r.service.SaveHistorySnapshot()
				if err != nil {
					slog.Error("runtime_stop_save_failed", "error", err)
				} else {
					slog.Info("runtime_stopped")
				}
				resp <- err
				close(resp)
				return
			}
		}
	}()
}

func (r *Runtime) Stop() error {
	resp := make(chan error, 1)
	r.stopReqCh <- resp
	return <-resp
}

func (r *Runtime) StatusUpdates() <-chan RuntimeStatus {
	return r.statusCh
}

func (r *Runtime) HistoryUpdates() <-chan []historydomain.Entry {
	return r.historyCh
}

func (r *Runtime) Errors() <-chan error {
	return r.errCh
}

func (r *Runtime) publishStatus(snapshot StatusSnapshot) {
	status := RuntimeStatus{Updated: time.Now()}
	if !snapshot.Found {
		status.State = "monitoring"
		r.logStateTransition(status)
		r.sendStatus(status)
		return
	}

	status.GameName = snapshot.GameName
	status.Elapsed = snapshot.Elapsed
	if snapshot.Focused {
		status.State = "tracking"
	} else {
		status.State = "paused"
	}
	r.logStateTransition(status)
	r.sendStatus(status)
}

func (r *Runtime) logStateTransition(status RuntimeStatus) {
	if status.State == r.lastLoggedState && status.GameName == r.lastLoggedGame {
		return
	}
	slog.Info("tracking_state_changed", "state", status.State, "game", status.GameName)
	r.lastLoggedState = status.State
	r.lastLoggedGame = status.GameName
}

func (r *Runtime) sendStatus(status RuntimeStatus) {
	select {
	case r.statusCh <- status:
	default:
		select {
		case <-r.statusCh:
		default:
		}
		r.statusCh <- status
	}
}

func (r *Runtime) publishHistory(entries []historydomain.Entry) {
	if len(entries) == 0 {
		return
	}
	copied := append([]historydomain.Entry(nil), entries...)
	select {
	case r.historyCh <- copied:
	default:
		select {
		case <-r.historyCh:
		default:
		}
		r.historyCh <- copied
	}
}

func (r *Runtime) ReportError(err error) {
	if err == nil {
		return
	}
	select {
	case r.errCh <- err:
	default:
		select {
		case <-r.errCh:
		default:
		}
		r.errCh <- fmt.Errorf("%w", err)
	}
}
