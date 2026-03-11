# Game Time Tracker

Local Windows tracker for game sessions with RTSS overlay output and a Bubble Tea terminal dashboard.

## Requirements

- Windows (RTSS shared memory integration).
- Go 1.25+
- RivaTuner Statistics Server (RTSS) running.

## Setup

1. Install dependencies:

```bash
go mod tidy
```

2. Configure watched processes in `config.json`:

```json
{
	"watchedProcesses": [
		"PapersPlease.exe",
		"CS2.exe"
	]
}
```

3. Ensure RTSS is running before starting the tracker.

## Run

```bash
go run .
```

Enable debug logs (written to `tracker.log`, not to TUI terminal):

```bash
go run . -debug
```

## Build

```bash
go build ./...
```

## Tests

```bash
go test ./...
```

Race-focused check for tracking concurrency:

```bash
go test -race ./internal/application/tracking
```

## Runtime Behavior

- The tracker runs process scanning in a background runtime goroutine.
- Bubble Tea TUI runs on the main thread.
- RTSS overlay shows active state and elapsed timer.
- History is persisted to `playtime_history.json` when a tracked game closes and during shutdown.
- Single-instance lock prevents two trackers from running simultaneously.

## TUI Controls

- `1`: Dashboard view.
- `2`: Active status view.
- `tab`, `left`, `right`: switch views.
- `q` or `ctrl+c`: quit.

## Data Files

- `config.json`: watched process configuration.
- `playtime_history.json`: persisted game totals and last played timestamps.

## Logging Strategy

The app uses concise structured logs (`log/slog`) for long-running sessions.

- Lifecycle events: runtime start/stop, shutdown complete.
- State transitions: monitoring/tracking/paused changes.
- Persistence events: successful save count and save failures.
- Startup failures: config/history loading and lock acquisition errors.

Logs avoid per-tick spam by only reporting significant transitions and save operations.