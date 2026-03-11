package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	apptracking "game-time-tracker/internal/application/tracking"
	"game-time-tracker/internal/configuration"
	infrahistory "game-time-tracker/internal/infrastructure/history"
	infraoverlay "game-time-tracker/internal/infrastructure/overlay"
	infraruntime "game-time-tracker/internal/infrastructure/runtime"
	infrascanner "game-time-tracker/internal/infrastructure/scanner"
)

func main() {
	releaseLock, alreadyRunning, err := infraruntime.AcquireSingleInstance()
	if err != nil {
		fmt.Println("Error starting single-instance lock:", err)
		return
	}
	if alreadyRunning {
		fmt.Println("Another Game Time Tracker instance is already running.")
		return
	}
	defer releaseLock()

	// 1. Configuration + persistence setup
	cfg, err := configuration.Load("config.json")
	if err != nil {
		fmt.Println("Error loading config.json:", err)
		return
	}

	historyRepo := infrahistory.NewJSONRepository("playtime_history.json")
	if _, err := historyRepo.Load(); err != nil {
		fmt.Println("Error loading playtime history:", err)
		return
	}

	scanner := infrascanner.NewProcessScanner(cfg.WatchedProcesses)
	overlay := infraoverlay.NewRTSSOverlay()
	service := apptracking.NewServiceWithHistory(scanner, overlay, historyRepo)
	runtime := apptracking.NewRuntime(service, 200*time.Millisecond)

	// 2. Initialize the interface
	overlay.Init()
	defer overlay.Close()

	// 3. Run scanner/tracking in a background goroutine.
	runtime.Start()
	statusCh := runtime.StatusUpdates()
	historyCh := runtime.HistoryUpdates()
	errCh := runtime.Errors()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	for {
		select {
		case <-statusCh:
			// Reserved for TUI status rendering on main thread.
		case <-historyCh:
			// Reserved for TUI/dashboard history refresh events.
		case err := <-errCh:
			if err != nil {
				fmt.Println("Runtime error:", err)
			}
		case sig := <-sigCh:
			fmt.Printf("Received signal %s, shutting down...\n", sig)
			if err := runtime.Stop(); err != nil {
				fmt.Println("Error saving history during shutdown:", err)
			}
			return
		}
	}
}
