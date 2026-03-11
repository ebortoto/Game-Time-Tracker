package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	apptracking "game-time-tracker/internal/application/tracking"
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

	// 1. Configuration
	// TIP: Add "notepad.exe" or "calc.exe" (legacy version) for quick testing.
	gameList := []string{"PapersPlease.exe"}
	scanner := infrascanner.NewProcessScanner(gameList)
	overlay := infraoverlay.NewRTSSOverlay()
	service := apptracking.NewService(scanner, overlay)

	// 2. Initialize the interface
	overlay.Init()
	defer overlay.Close()

	// 3. Main loop: keeps the process alive and updates tracking once per second.
	ticker := time.NewTicker(1 * time.Second)
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
