# Game Time Tracker

Game tracker split into two runtime components:
- Client: process scanner + TUI + optional RTSS overlay.
- Server: authenticated HTTP API + persisted history storage.

## Requirements

- Go 1.25+
- Windows for RTSS overlay integration.
- RivaTuner Statistics Server (RTSS) running (client side only, optional when using `-overlay=false`).

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

3. Set shared auth key for client/server communication:

```bash
set TRACKER_API_KEY=change-me
```

## Run (Split Mode)

1. Start server:

```bash
go run ./cmd/server -addr :8080 -history-file playtime_history.json
```

2. Start client:

```bash
go run ./cmd/client -server-url http://localhost:8080 -config config.json
```

Optional client flags:
- `-debug` writes JSON logs to `tracker.log`.
- `-overlay=false` disables RTSS output (useful in non-Windows/container runs).

## Build

```bash
go build ./...
```

Build only server/client binaries:

```bash
go build ./cmd/server
go build ./cmd/client
```

## Docker

Run only the server in Docker:

```bash
docker compose up --build
```

Notes:
- Keep the TUI client on the host OS (recommended for process scanning + RTSS overlay).
- Start the client against Dockerized server:

```bash
go run ./cmd/client -server-url http://localhost:8080 -config config.json
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

- Client runs process scanning in a background runtime goroutine.
- Bubble Tea TUI runs on the client main thread.
- RTSS overlay (if enabled) shows active state and elapsed timer.
- Client pushes history deltas to server on game close and shutdown.
- Server persists merged history to `playtime_history.json` (or configured file).
- Single-instance lock prevents two clients from running simultaneously.

## TUI Controls

- `1`: Dashboard view.
- `2`: Active status view.
- `tab`, `left`, `right`: switch views.
- `q` or `ctrl+c`: quit.

## Data Files

- `config.json`: watched process configuration.
- `playtime_history.json`: server-side persisted game totals and last played timestamps.

## API Contract

See `docs/api-contract.md` for auth model and JSON schemas.

## Logging Strategy

The app uses concise structured logs (`log/slog`) for long-running sessions.

- Lifecycle events: runtime start/stop, shutdown complete.
- State transitions: monitoring/tracking/paused changes.
- Persistence events: successful save count and save failures.
- Startup failures: config/history loading and lock acquisition errors.

Logs avoid per-tick spam by only reporting significant transitions and save operations.