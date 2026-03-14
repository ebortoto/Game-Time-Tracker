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

- [x] Add database service to docker-compose
  - Add MySQL container, volume, and healthcheck.
  - Keep server container as app runtime.
  - Status: docker-compose now includes `mysql` service with persistent volume, healthcheck, and `tracker-server` dependency on healthy DB while keeping server as app runtime.

- [x] Add DB environment variables to server config
  - `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`.
  - Validate required vars when DB mode is enabled.
  - Status: server now reads `DB_HOST/DB_PORT/DB_USER/DB_PASSWORD/DB_NAME` for MySQL DSN construction and uses mode-aware env validation for `HISTORY_BACKEND=mysql`.

- [x] Create SQL schema migration script
  - Create table for daily history with unique key `(game_name, date)`.
  - Add startup migration execution.
  - Status: added `migrations/001_create_daily_history.sql` with unique key on `(game_name, play_date)` and startup execution via `RunMySQLMigrations`.

- [x] Implement MySQL history repository
  - `Load()` returns daily entries.
  - `Save(entries)` performs upsert increments.
  - Status: added `MySQLRepository` with `Load()` select mapping and `Save()` transactional upsert increment logic using `ON DUPLICATE KEY UPDATE`.

- [x] Wire repository mode selection
  - `HISTORY_BACKEND=json|mysql` env switch.
  - Keep JSON as temporary fallback during migration.
  - Status: server now switches repository by `HISTORY_BACKEND` (`mysql` or `json`) with JSON fallback/default path preserved.

- [x] Add repository tests for upsert and load behavior
  - Insert new day entry.
  - Increment existing day entry.
  - Status: added MySQL repository tests with sqlmock for upsert path and load row mapping behavior.

## D) Environment Files

- [x] Add `.env.example` with documented variables
  - Include server, client, auth, and DB variables.
  - Keep example values safe and non-secret.
  - Status: added `.env.example` with shared auth, server, client, backend mode, and DB placeholders using safe non-secret defaults.

- [x] Add `.env` loader to server and client startup
  - Load `.env` before parsing defaults.
  - Keep CLI flags as highest precedence.
  - Status: both `cmd/server` and `cmd/client` now load `.env` at startup, derive flag defaults from env values, and preserve CLI flags as highest precedence.

- [x] Add env validation helper
  - Fail fast for missing required vars by mode.
  - Print actionable error messages.
  - Status: added `internal/configuration/env.go` validation helpers for client URL requirements and backend-mode checks (`json|mysql`, with MySQL required DB vars enforcement).

- [x] Update README with env usage patterns
  - Local run with `.env`.
  - Docker compose env flow.
  - Status: README now documents `.env` setup, precedence order (CLI > env > .env > defaults), local run flow, and Docker compose environment behavior.

## E) Future Hosted API Readiness (small steps now)

- [x] Add explicit API base URL + token docs
  - Document local and remote examples.
  - Status: README and API contract now include explicit local/hosted base URL and token configuration examples.

- [x] Ensure no hardcoded localhost in runtime defaults
  - Use env/flags for all endpoints.
  - Status: runtime endpoint selection remains env/flag-driven; removed MySQL host localhost fallback in server runtime DSN defaults.

- [x] Add deployment notes section in PRD/README
  - State migration path from local Docker DB to internet-hosted stack.
  - Status: added deployment notes in README and PRD describing local-first stack and migration steps to hosted server/database.

## Suggested Short Execution Order

1. D) Environment Files
2. A) Windows Detection Bootstrap
3. B) Tray Icon + Open TUI
4. C) Dockerized Database (Local First)
5. E) Future Hosted API Readiness
