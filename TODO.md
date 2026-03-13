# Game Time Tracker - TODO

Derived from PRD.md + current implementation state.

## Completed Foundation (kept for context)

- [x] DDD refactor, scanner/runtime separation, and Bubble Tea TUI.
- [x] Daily playtime tracking with idempotent save behavior.
- [x] Server/client split with authenticated history API.
- [x] Server-only Docker compose setup.

## Current Product Objectives (March 2026)

1. Detection framework initializes cleanly on Windows.
2. Tray icon shows tracker status.
3. Tray click offers Open TUI action.
4. Database runs in Docker now (future: hosted online API/database).
5. Environment is configurable via .env files.

## A) Windows Detection Bootstrap

- [x] Add explicit Windows initialization flow in client startup
  - Create `InitializeWindowsRuntime()` in infrastructure layer.
  - Return startup diagnostics instead of silent failures.
  - Status: startup now calls `InitializeWindowsRuntime()` and logs per-step diagnostics before scanner runtime starts.

- [x] Add platform guard + fallback messages
  - Use build tags for Windows-specific init path.
  - Print clear warning when Windows-only feature is unavailable.
  - Status: build-tagged Windows/non-Windows init paths are active and client startup now emits a clear compatibility warning on non-Windows platforms.

- [x] Add startup integration test for bootstrap path
  - Test successful init path with fakes/mocks.
  - Test failure path message and non-zero exit behavior.
  - Status: added bootstrap integration tests in `cmd/client/main_test.go` with fake runtime initializers covering success, fallback warning, and non-zero failure behavior.

## B) Tray Icon + Open TUI

- [x] Add minimal tray package with lifecycle hooks
  - Start tray icon on client startup.
  - Stop tray icon on graceful shutdown.
  - Status: replaced stub lifecycle hooks with a real systray-backed service (`github.com/getlantern/systray`) and kept lifecycle coverage with runtime-fake tests.

- [x] Add tray menu entries
  - Add `Open TUI` action.
  - Add `Exit` action.
  - Status: tray service now defines `Open TUI` and `Exit` menu entries, supports handlers, and client startup registers both actions.

- [x] Implement Open TUI command bridge
  - If TUI is closed, open it.
  - If TUI is already open, focus/no-op (no duplicates).
  - Status: added tray `TUIBridge` with open/close state guard; tray `Open TUI` now triggers open when closed and logs no-op when already open.

- [x] Add quick manual verification checklist in README
  - Startup -> icon visible.
  - Open TUI works.
  - Exit from tray shuts down cleanly.
  - Status: README now includes a tray/TUI verification checklist covering startup, icon visibility, Open TUI behavior, and Exit behavior.

## C) Dockerized Database (Local First)

- [ ] Add database service to docker-compose
  - Add MySQL container, volume, and healthcheck.
  - Keep server container as app runtime.

- [ ] Add DB environment variables to server config
  - `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`.
  - Validate required vars when DB mode is enabled.

- [ ] Create SQL schema migration script
  - Create table for daily history with unique key `(game_name, date)`.
  - Add startup migration execution.

- [ ] Implement MySQL history repository
  - `Load()` returns daily entries.
  - `Save(entries)` performs upsert increments.

- [ ] Wire repository mode selection
  - `HISTORY_BACKEND=json|mysql` env switch.
  - Keep JSON as temporary fallback during migration.

- [ ] Add repository tests for upsert and load behavior
  - Insert new day entry.
  - Increment existing day entry.

## D) Environment Files

- [ ] Add `.env.example` with documented variables
  - Include server, client, auth, and DB variables.
  - Keep example values safe and non-secret.

- [ ] Add `.env` loader to server and client startup
  - Load `.env` before parsing defaults.
  - Keep CLI flags as highest precedence.

- [ ] Add env validation helper
  - Fail fast for missing required vars by mode.
  - Print actionable error messages.

- [ ] Update README with env usage patterns
  - Local run with `.env`.
  - Docker compose env flow.

## E) Future Hosted API Readiness (small steps now)

- [ ] Add explicit API base URL + token docs
  - Document local and remote examples.

- [ ] Ensure no hardcoded localhost in runtime defaults
  - Use env/flags for all endpoints.

- [ ] Add deployment notes section in PRD/README
  - State migration path from local Docker DB to internet-hosted stack.

## Suggested Short Execution Order

1. D) Environment Files
2. A) Windows Detection Bootstrap
3. B) Tray Icon + Open TUI
4. C) Dockerized Database (Local First)
5. E) Future Hosted API Readiness
