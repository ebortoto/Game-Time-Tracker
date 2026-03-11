# Game Time Tracker - TODO

Derived from PRD.md + current implementation state.

## 0) New mandatory requirement: DDD + refactor

- [x] Implement the project using Domain-Driven Design (DDD)
  - Organize code by domain boundaries, not by technical utility only.
  - Proposed bounded contexts:
    - `tracking` (game session detection and stopwatch rules)
    - `history` (playtime aggregation and persistence rules)
    - `configuration` (watch-list and runtime settings)
    - `overlay` (RTSS output as infrastructure adapter)
    - `console` (TUI as presentation layer)
  - Acceptance: core business logic can run without RTSS/TUI dependencies.

- [x] Refactor code to DDD layers
  - `domain`: entities, value objects, domain services, repository interfaces.
  - `application`: use cases (start tracking, tick, pause, save history, shutdown).
  - `infrastructure`: process scanner adapter, file/json storage, RTSS adapter.
  - `interfaces`: CLI/TUI models and controllers.
  - Acceptance: dependencies point inward (`interfaces` -> `application` -> `domain`).

- [x] Refactor where needed before adding major features
  - Move stopwatch and playtime rules to domain types with tests.
  - Replace direct package coupling in `main.go` with application services.
  - Introduce interfaces for scanner, overlay writer, and history repository.
  - Acceptance: major flows are testable via mocks/fakes without OS process scanning.
  - Status: stopwatch moved to domain, main wired to application service, scanner/overlay/history repository interfaces introduced.
  - Completed: unit tests added for domain stopwatch and application tracking service rules.

## 1) Critical fixes (do first)

- [ ] Fix RTSS memory lifetime in internal/ui/overlay.go
  - Remove `defer procUnmapViewOfFile.Call(addr)` from InitOverlay.
  - Remove `defer procCloseHandle.Call(handle)` from InitOverlay.
  - Keep mapped address and handle in package-level vars.
  - Add `CloseOverlay()` to unmap view and close handle on shutdown.
  - Add nil/zero guards in `UpdateText` to avoid invalid writes.
  - Acceptance: overlay updates work for minutes without crash/access violation.

- [ ] Remove obsolete always-on-top code path
  - Delete `detector.SetAlwaysOnTop("GameTimerOverlay")` usage in main.go.
  - Remove unused topmost helpers/constants in internal/detector/windows.go.
  - Acceptance: build passes and behavior is unchanged for RTSS overlay text.

## 2) Graceful shutdown

- [ ] Handle SIGINT/SIGTERM in main.go
  - Use `os/signal` and `syscall`.
  - Stop ticker cleanly.
  - Pause all running stopwatches before exit.
  - Trigger final persistence save.
  - Call `ui.CloseOverlay()` before exit.
  - Acceptance: Ctrl+C exits cleanly with no data loss.

## 3) Storage and persistence

- [ ] Replace hardcoded game list with config file
  - Add config file (config.json or config.yaml).
  - Fields: list of watched process names.
  - Load on startup with validation and fallback error message.
  - Acceptance: add/remove games without recompiling.

- [ ] Create storage package for history
  - Suggested file: internal/storage/storage.go.
  - Add model: `GameName`, `TotalTimeSecs`, `LastPlayedDate`.
  - Implement `LoadHistory(path)` and `SaveHistory(path, data)`.
  - Acceptance: data survives app restart.

- [ ] Persist playtime on events
  - Save when tracked game closes.
  - Save when app shuts down.
  - Keep in-memory session deltas merged into historical totals.
  - Acceptance: historical totals increase correctly across sessions.

## 4) Concurrency architecture for TUI

- [ ] Move scanner loop to background goroutine
  - Keep TUI on main thread.
  - Use channels for scanner status + history updates.
  - Protect shared maps/state (single owner goroutine or mutex).
  - Acceptance: no data races under `go test -race` (or `go run -race`).

## 5) Terminal UI (Bubble Tea)

- [ ] Add Bubble Tea dependencies and app model
  - Create a TUI package (for example internal/tui).
  - Define messages for status and history refresh.

- [ ] Implement View 1: Dashboard
  - Show tracked games and total historical playtime.
  - Include last played date.

- [ ] Implement View 2: Active status
  - Show scanner state: monitoring / tracking / paused.
  - Show active game and elapsed session timer.

- [ ] Wire controls and navigation
  - Keybinds to switch views, refresh, quit.
  - Acceptance: user can monitor status and totals from terminal while tracker runs.

## 6) Quality and maintenance

- [ ] Add basic tests
  - Unit tests for timer math.
  - Unit tests for storage load/save/merge behavior.
  - Acceptance: core logic validated automatically.

- [ ] Improve README
  - Setup instructions (RTSS requirement, run commands).
  - Config format examples.
  - Expected runtime behavior and shutdown flow.

- [ ] Add logging strategy
  - Structured logs for scan tick, state transitions, save operations, shutdown.
  - Keep logs concise for long-running sessions.

## Suggested milestone order

1. DDD architecture + foundational refactor
2. Critical fixes
3. Graceful shutdown
4. Storage + config
5. Background scanner + channels
6. Bubble Tea TUI
7. Tests + README polish
