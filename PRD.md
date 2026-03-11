Product Requirements Document (PRD): Game Time Tracker
1. Project Overview
A lightweight, local Go-based service that monitors system processes to track active game sessions. It provides real-time feedback via an RTSS (RivaTuner Statistics Server) on-screen overlay and logs historical playtime data to help the user review and reduce their gaming habits.

2. Current State vs. Target State
Current State: The application successfully hardcodes a list of games, detects if they are running and in focus, tracks the session time in memory, and pushes strings to the RTSS shared memory.

Target State: The application needs to persistently save data between sessions, allow dynamic configuration of the "watch list" without recompiling, provide a TUI (Terminal User Interface) dashboard for interaction, and gracefully handle its own lifecycle.

3. Core Features to Implement
Feature 1: Configuration & Data Persistence (Storage)
To track habits over time, the data must survive application restarts.

Requirements:

Replace the hardcoded gameList in main.go with a configuration file (e.g., config.json or config.yaml).

Create a local database or JSON file (e.g., playtime_history.json) to store playtime history.

Data Structure: Track at least gameName, date (YYYY-MM-DD), totalTimeSecs, and lastPlayedDate.

Daily Tracking Requirement: Persist daily totals per game so the application can compute "time played today" across restarts.

Daily Data Structure: Track at least gameName, date (YYYY-MM-DD), totalTimeSecs, and lastPlayedDate.

Save Triggers: The app must save the current stopwatch data to the file when a game is closed, or when the tracker itself is shut down.

Feature 2: The Terminal User Interface (TUI)
The TUI will act as the control panel for the background tracker.

Requirements:

Implement a TUI using a library like charmbracelet/bubbletea.

View 1 (Dashboard): Show a list of all tracked games and their total historical playtime.

View 2 (Active Status): Show what the background scanner is currently doing (e.g., "Monitoring...", "Tracking CS2 - 00:15:22"). The timer value must represent playtime accumulated for the current day (not lifetime total).

Concurrency: The TUI must run on the main thread while the ticker loop from your current main.go runs continuously in a separate Goroutine. They should communicate via Go channels.

Feature 3: Graceful Shutdown
Currently, the program runs infinitely until forcibly killed.

Requirements:

Listen for OS interrupt signals (SIGINT/SIGTERM) using os/signal.

When an interrupt is received, safely stop all active stopwatches, trigger a final save to the persistent storage, unmap the RTSS memory, and then exit.

4. Technical Debt & Refactoring
Before building new features, the existing code requires a few critical adjustments:

The RTSS Memory Scope Bug: In your overlay.go, the InitOverlay() function maps the view of the file (procMapViewOfFile.Call), but you have defer procUnmapViewOfFile.Call(addr) and defer procCloseHandle.Call(handle). Because of the defer, the memory is unmapped the moment InitOverlay() finishes executing. Subsequent calls to UpdateText will write to a memory address that the program no longer holds, which can cause silent failures or access violations.

Fix: Remove the defer statements from InitOverlay. Create a CloseOverlay() function that handles the unmapping and handle closing, and call defer ui.CloseOverlay() in your main.go.

Remove Unused Fyne Code: In windows.go, the SetAlwaysOnTop function and constants are leftover from a previous GUI approach. Since you are using RTSS for the overlay, detector.SetAlwaysOnTop("GameTimerOverlay") in main.go is unnecessary and should be removed to keep the codebase clean.

5. Recommended Implementation Order
Refactor: Fix the memory unmapping bug in overlay.go, remove the SetAlwaysOnTop code, and implement a graceful shutdown using os/signal.

Storage: Build a storage.go package. Define your structs for saving/loading data to a JSON file. Test it by loading the file on startup and saving when the ticker detects a game has closed.

Concurrency: Move your for range ticker.C loop into a Goroutine so it runs in the background.

TUI: Initialize bubbletea on the main thread. Have your background Goroutine send message updates to the TUI to render the dashboard.

6. Daily Playtime Display (Priority Update)
To align the overlay and TUI with behavior goals, the displayed timer for active tracking must be "today's total" for the game.

Requirements:

When a tracked game is active, display today's accumulated time (historical today + current session delta).

On startup, load today's persisted value per game and continue from it.

On game close/shutdown, persist only the delta for today so repeated saves do not double count.

Acceptance:

If the app restarts on the same day, the timer resumes from the previously saved daily value.

At day rollover, displayed time resets for the new date while historical daily records remain accessible.