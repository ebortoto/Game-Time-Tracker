package main

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
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
	"game-time-tracker/internal/tui"
	"game-time-tracker/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	debug := flag.Bool("debug", false, "enable debug logs in tracker.log")
	flag.Parse()

	if *debug {
		logFile, err := os.OpenFile("tracker.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err == nil {
			defer logFile.Close()
			slog.SetDefault(slog.New(slog.NewJSONHandler(logFile, &slog.HandlerOptions{Level: slog.LevelDebug})))
		} else {
			slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
		}
	} else {
		slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	}
	ui.SetDebugEnabled(*debug)

	releaseLock, alreadyRunning, err := infraruntime.AcquireSingleInstance()
	if err != nil {
		slog.Error("single_instance_lock_failed", "error", err)
		fmt.Println("Error starting single-instance lock:", err)
		return
	}
	if alreadyRunning {
		slog.Warn("single_instance_already_running")
		fmt.Println("Another Game Time Tracker instance is already running.")
		return
	}
	defer releaseLock()

	// 1. Configuration + persistence setup
	cfg, err := configuration.Load("config.json")
	if err != nil {
		slog.Error("config_load_failed", "file", "config.json", "error", err)
		fmt.Println("Error loading config.json:", err)
		return
	}

	historyRepo := infrahistory.NewJSONRepository("playtime_history.json")
	initialHistory, err := historyRepo.Load()
	if err != nil {
		slog.Error("history_load_failed", "file", "playtime_history.json", "error", err)
		fmt.Println("Error loading playtime history:", err)
		return
	}
	slog.Info("startup_complete", "watched_processes", len(cfg.WatchedProcesses), "history_records", len(initialHistory))

	scanner := infrascanner.NewProcessScanner(cfg.WatchedProcesses)
	overlay := infraoverlay.NewRTSSOverlay()
	service := apptracking.NewServiceWithHistory(scanner, overlay, historyRepo)
	service.SetInitialHistory(initialHistory, time.Now())
	runtime := apptracking.NewRuntime(service, 200*time.Millisecond)

	// 2. Initialize the interface
	overlay.Init()
	defer overlay.Close()

	// 3. Run scanner/tracking in a background goroutine.
	runtime.Start()
	statusCh := runtime.StatusUpdates()
	historyCh := runtime.HistoryUpdates()
	errCh := runtime.Errors()

	model := tui.NewModel(statusCh, historyCh, errCh, initialHistory)
	program := tea.NewProgram(model)

	// Clear terminal before rendering the TUI.
	fmt.Print("\033[H\033[2J")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	sigDone := make(chan struct{})
	go func() {
		select {
		case sig := <-sigCh:
			program.Send(tui.SignalMsg{Signal: sig.String()})
		case <-sigDone:
		}
	}()

	if _, err := program.Run(); err != nil {
		slog.Error("tui_runtime_error", "error", err)
		fmt.Println("TUI runtime error:", err)
	}

	close(sigDone)
	defer signal.Stop(sigCh)

	if err := runtime.Stop(); err != nil {
		slog.Error("shutdown_save_failed", "error", err)
		fmt.Println("Error saving history during shutdown:", err)
		return
	}
	slog.Info("shutdown_complete")
}
