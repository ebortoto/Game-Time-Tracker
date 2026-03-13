Product Requirements Document (PRD): Game Time Tracker

1. Project Overview
Game Time Tracker is a Windows-first Go application that detects active games, tracks daily playtime, and presents status in a TUI. The runtime is split into a local client and a server API. The near-term goal is stronger desktop operability (Windows startup + tray UX) and environment-driven deployment with a Dockerized database.

2. Product Direction (March 2026 Update)
The active objectives are:

1. The detection framework must initialize reliably on Windows.
2. The app must expose a system tray icon that indicates the tracker is running.
3. Clicking the tray icon must expose an action to open the TUI.
4. Persistence must move to a database running in Docker for local development.
5. Runtime configuration must be controlled via environment files.

3. Current State vs Target State

Current State:
- Process scanner and tracking service are working.
- TUI is implemented with Bubble Tea.
- Server API exists and currently persists history in JSON.
- Docker is currently server-only.

Target State:
- Windows bootstrap path initializes detection and tray integrations safely.
- Tray icon provides at least Open TUI and Exit actions.
- Server persists to Dockerized database (instead of JSON) for local runs.
- Configuration is loaded from .env and environment variables in both server and client.
- Architecture remains compatible with a future internet-hosted API backend.

4. Functional Requirements

Feature A: Windows Detection Initialization
- On Windows startup, detection dependencies initialize once and fail with actionable logs.
- Unsupported platforms must fail gracefully for Windows-only features.
- Acceptance:
	- Tracker starts and scanner enters monitoring state on Windows.
	- Startup errors are explicit and include initialization step context.

Feature B: Tray Icon + Open TUI
- Add a tray icon while the tracker client is running.
- Tray menu must include:
	- Open TUI
	- Exit
- Open TUI behavior:
	- If TUI is not open, launch/focus it.
	- If TUI is already open, do not spawn duplicates.
- Acceptance:
	- Tray icon appears after startup.
	- User can open TUI through tray action.
	- Exit from tray performs graceful shutdown.

Feature C: Dockerized Database (Local)
- Replace JSON persistence path with database persistence for local development.
- Database must run in Docker Compose.
- Server connects using environment variables (host, port, user, password, db name).
- Acceptance:
	- `docker compose up` starts server dependencies, including the database.
	- Server can read/write playtime history through repository abstraction.
	- Existing history fields are preserved: gameName, date, totalTimeSecs, lastPlayedDate.

Feature D: Environment File Support
- Add .env-based configuration for server and client.
- Keep explicit command-line flags as highest precedence.
- Define and document required and optional variables.
- Acceptance:
	- Running with `.env` config does not require hardcoded secrets.
	- Missing required values fail fast with clear messages.

Feature E: Future Hosted API Compatibility
- Keep server API contract and client repository abstractions portable.
- Avoid local-only assumptions in domain/application layers.
- Acceptance:
	- Base URL and auth token can be switched by env without code changes.

5. Non-Functional Requirements
- Maintain DDD layering and inward dependencies.
- Keep scanner loop and UI responsive (no blocking tray/UI events).
- Maintain graceful shutdown semantics.
- Keep logs structured and concise.

6. Constraints and Risks
- Tray implementation is OS-specific; use build tags for platform code.
- TUI process lifecycle and tray callbacks require careful synchronization.
- Database migration from JSON to SQL needs idempotent schema migration and rollback-safe startup.

7. Delivery Plan

Phase 1: Environment + bootstrap foundation
- Add .env loading and env validation.
- Add Windows initialization checks and startup diagnostics.

8. Implementation Progress

- 2026-03-11: Completed the first Windows bootstrap task.
	- Added `InitializeWindowsRuntime()` in infrastructure runtime layer with platform-specific implementations.
	- Client startup now runs initialization before scanner runtime and logs per-step diagnostics.
	- Added a non-Windows bootstrap test to verify diagnostic behavior on unsupported platforms.

- 2026-03-11: Completed platform guard and fallback messaging task.
	- Confirmed build-tagged Windows/non-Windows initialization paths.
	- Client now prints and logs an explicit fallback warning when running outside Windows.

- 2026-03-11: Completed startup integration test task for bootstrap path.
	- Refactored client bootstrap sequence into a testable helper (`runWindowsBootstrap`).
	- Added tests with fake runtime initializers for success, non-Windows fallback messaging, and non-zero exit behavior on initialization failure.

- 2026-03-11: Completed minimal tray lifecycle hook task.
	- Added `internal/infrastructure/tray` service with start/stop lifecycle API.
	- Client startup now starts tray service and stops it during shutdown via deferred cleanup.
	- Added unit tests for tray lifecycle and idempotent start/stop behavior.

- 2026-03-11: Completed tray menu entry task.
	- Added explicit tray menu actions for `Open TUI` and `Exit` in tray infrastructure service.
	- Added handler registration/trigger API and test coverage for action dispatch.
	- Client now registers handlers for both tray menu actions during startup.

- 2026-03-11: Completed Open TUI command bridge task.
	- Added `TUIBridge` component to coordinate open requests and avoid duplicate TUI opens.
	- Wired tray `Open TUI` action to bridge behavior with open/no-op logging.
	- Added unit tests covering open when closed, no-op when open, reopen after close, and opener error handling.

- 2026-03-11: Completed tray/TUI manual verification checklist task.
	- Added a practical checklist in README covering startup, tray icon visibility, Open TUI behavior, and Exit behavior.

Phase 2: Tray UX
- Add tray icon, menu actions, and lifecycle hooks.
- Add Open TUI behavior with single-instance guard for TUI view.

Phase 3: Dockerized database migration
- Add DB service to compose.
- Implement SQL repository and wire it behind existing ports.
- Keep JSON path optional only as fallback during migration window.

Phase 4: Hosted API readiness
- Harden API/auth configuration via env.
- Document deployment knobs for remote hosting.