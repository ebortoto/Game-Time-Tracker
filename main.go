package main

import (
	"fmt"
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

	for range ticker.C {
		service.Tick()
	}
}
