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

- [x] Fix RTSS memory lifetime in internal/ui/overlay.go
  - Remove `defer procUnmapViewOfFile.Call(addr)` from InitOverlay.
  - Remove `defer procCloseHandle.Call(handle)` from InitOverlay.
  - Keep mapped address and handle in package-level vars.
  - Add `CloseOverlay()` to unmap view and close handle on shutdown.
  - Add nil/zero guards in `UpdateText` to avoid invalid writes.
  - Acceptance: overlay updates work for minutes without crash/access violation.

- [x] Remove obsolete always-on-top code path
  - Delete `detector.SetAlwaysOnTop("GameTimerOverlay")` usage in main.go.
  - Remove unused topmost helpers/constants in internal/detector/windows.go.
  - Acceptance: build passes and behavior is unchanged for RTSS overlay text.

## 2) Graceful shutdown

- [x] Handle SIGINT/SIGTERM in main.go
  - Use `os/signal` and `syscall`.
  - Stop ticker cleanly.
  - Pause all running stopwatches before exit.
  - Trigger final persistence save.
  - Call `ui.CloseOverlay()` before exit.
  - Acceptance: Ctrl+C exits cleanly with no data loss.
  - Status: signal-driven shutdown added in main loop; shutdown now pauses timers and triggers final snapshot save through the application service.

## 3) Storage and persistence

- [x] Replace hardcoded game list with config file
  - Add config file (config.json or config.yaml).
  - Fields: list of watched process names.
  - Load on startup with validation and fallback error message.
  - Acceptance: add/remove games without recompiling.

- [x] Create storage package for history
  - Suggested file: internal/storage/storage.go.
  - Add model: `gameName`, `date`, `totalTimeSecs`, `lastPlayedDate`.
  - Implement `LoadHistory(path)` and `SaveHistory(path, data)`.
  - Acceptance: data survives app restart.

- [x] Persist playtime on events
  - Save when tracked game closes.
  - Save when app shuts down.
  - Keep in-memory session deltas merged into historical totals.
  - Acceptance: historical totals increase correctly across sessions.
  - Status: game-close transition triggers snapshot persistence and shutdown path performs final save.

## 4) Concurrency architecture for TUI

- [x] Move scanner loop to background goroutine
  - Keep TUI on main thread.
  - Use channels for scanner status + history updates.
  - Protect shared maps/state (single owner goroutine or mutex).
  - Acceptance: no data races under `go test -race` (or `go run -race`).
  - Status: added application runtime goroutine with status/history channels; main thread now orchestrates signals and consumes update streams.

## 5) Terminal UI (Bubble Tea)

- [x] Add Bubble Tea dependencies and app model
  - Create a TUI package (for example internal/tui).
  - Define messages for status and history refresh.

- [x] Implement View 1: Dashboard
  - Show tracked games and total historical playtime.
  - Include last played date.

- [x] Implement View 2: Active status
  - Show scanner state: monitoring / tracking / paused.
  - Show active game and elapsed session timer.

- [x] Wire controls and navigation
  - Keybinds to switch views, refresh, quit.
  - Acceptance: user can monitor status and totals from terminal while tracker runs.
  - Status: Bubble Tea runs on main thread while tracker runtime runs in background; status/history channels feed both views.

## 6) Quality and maintenance

- [x] Add basic tests
  - Unit tests for timer math.
  - Unit tests for storage load/save/merge behavior.
  - Acceptance: core logic validated automatically.
  - Status: added tests for internal timer stopwatch math, storage load/save round-trip, and history repository merge semantics.

- [x] Improve README
  - Setup instructions (RTSS requirement, run commands).
  - Config format examples.
  - Expected runtime behavior and shutdown flow.
  - Status: README now documents setup, config, run/build/test commands, runtime behavior, controls, and data files.

- [x] Add logging strategy
  - Structured logs for scan tick, state transitions, save operations, shutdown.
  - Keep logs concise for long-running sessions.
  - Status: structured `log/slog` events added for runtime lifecycle, state transitions, save outcomes, and startup/shutdown failures.

## 7) Next priority: Daily playtime display

- [x] Change displayed timer to "time played today"
  - Overlay and Active Status view must show today's accumulated time only.
  - Include historical time already played today plus current in-progress session.

- [x] Refactor persistence schema for daily aggregates
  - Store by game + date (YYYY-MM-DD) instead of only lifetime totals.
  - Use fields `gameName`, `date`, `totalTimeSecs`, and `lastPlayedDate` per day entry.

- [x] Prevent double counting on repeated saves
  - Persist session deltas for the current day, not full accumulated values on every save.
  - Ensure shutdown save and game-close save are idempotent for already persisted periods.

- [x] Handle day rollover
  - Reset displayed timer at local date change.
  - Keep previous days available for dashboard/history.

- [x] Add tests for daily behavior
  - Restart on same day resumes from saved daily value.
  - Crossing midnight starts a new day bucket.
  - Multiple saves in one session do not inflate totals.
  - Status: service tests now cover same-day baseline loading, rollover date-bucket persistence/reset, and delta idempotency.

## 8) Beautify the TUI

- [ ] Improve visual design and readability
  - Improve layout, typography/spacing, and section hierarchy for readability.
  - Add lightweight visual polish (table alignment, highlights, status emphasis).
  - Keep keyboard controls and low-latency updates intact.

## 9) Docker architecture: server + client

- [ ] Split runtime into server and client
  - Split into a server that stores tracking data and a client that scans/renders UI.
  - Add Dockerfiles (client/server) and a docker-compose stack for local spin-up.
  - Define API contract between client and server (auth + payload schema).

## 10) Steam library import menu

- [ ] Add menu for importing games from Steam library
  - Detect Steam install/library paths.
  - Parse installed games and expose selection in TUI menu.
  - Save selected games into tracker configuration.

## 11) Cloud MySQL persistence with GitHub Secrets

- [ ] Migrate persistence from JSON to managed MySQL
  - Replace file storage with repository implementation backed by MySQL.
  - Read credentials from environment variables (no hardcoded secrets).
  - Add CI/CD guidance for injecting credentials from GitHub Secrets.

## 12) YouTube watch time tracking

- [ ] Add YouTube watch-time activity tracking
  - Add detector for browser/YouTube activity and active playback heuristics.
  - Extend domain model to support non-game activity categories.
  - Include YouTube time in dashboard/status views with clear labels.

## Suggested milestone order

1. DDD architecture + foundational refactor
2. Critical fixes
3. Graceful shutdown
4. Storage + config
5. Background scanner + channels
6. Bubble Tea TUI
7. Tests + README polish
8. Daily playtime display
9. Beautify the TUI
10. Docker architecture: server + client
11. Steam library import menu
12. Cloud MySQL persistence with GitHub Secrets
13. YouTube watch time tracking
