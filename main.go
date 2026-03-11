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

	// 2. Initialize the interface
	overlay.Init()
	defer overlay.Close()

	// 3. Main loop: refresh overlay frequently; scanner polling stays throttled in service.
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	for {
		select {
		case <-ticker.C:
			service.Tick()
		case sig := <-sigCh:
			fmt.Printf("Received signal %s, shutting down...\n", sig)
			service.PauseAll()
			if err := service.SaveHistorySnapshot(); err != nil {
				fmt.Println("Error saving history during shutdown:", err)
			}
			return
		}
	}
}
