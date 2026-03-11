package main

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"strings"
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

type overlayAdapter interface {
	Init()
	UpdateText(text string)
	Close()
}

func main() {
	debug := flag.Bool("debug", false, "enable debug logs in tracker.log")
	serverURL := flag.String("server-url", "", "tracking server base URL (fallback: TRACKER_SERVER_URL)")
	apiKeyFlag := flag.String("api-key", "", "tracking server API key (fallback: TRACKER_API_KEY)")
	configPath := flag.String("config", "config.json", "tracker config file")
	overlayEnabled := flag.Bool("overlay", true, "enable RTSS overlay output")
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

	url := strings.TrimSpace(*serverURL)
	if url == "" {
		url = strings.TrimSpace(os.Getenv("TRACKER_SERVER_URL"))
	}
	if url == "" {
		fmt.Println("Error: missing server URL. Set -server-url or TRACKER_SERVER_URL.")
		return
	}

	apiKey := strings.TrimSpace(*apiKeyFlag)
	if apiKey == "" {
		apiKey = strings.TrimSpace(os.Getenv("TRACKER_API_KEY"))
	}

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

	cfg, err := configuration.Load(*configPath)
	if err != nil {
		slog.Error("config_load_failed", "file", *configPath, "error", err)
		fmt.Println("Error loading config:", err)
		return
	}

	historyRepo := infrahistory.NewRemoteRepository(url, apiKey, nil)
	initialHistory, err := historyRepo.Load()
	if err != nil {
		slog.Error("history_load_failed", "server_url", url, "error", err)
		fmt.Println("Error loading remote playtime history:", err)
		return
	}

	scanner := infrascanner.NewProcessScanner(cfg.WatchedProcesses)
	var overlay overlayAdapter = infraoverlay.NewRTSSOverlay()
	if !*overlayEnabled {
		overlay = infraoverlay.NewNoopOverlay()
	}

	service := apptracking.NewServiceWithHistory(scanner, overlay, historyRepo)
	service.SetInitialHistory(initialHistory, time.Now())
	runtime := apptracking.NewRuntime(service, 200*time.Millisecond)

	overlay.Init()
	defer overlay.Close()

	runtime.Start()
	statusCh := runtime.StatusUpdates()
	historyCh := runtime.HistoryUpdates()
	errCh := runtime.Errors()

	model := tui.NewModel(statusCh, historyCh, errCh, initialHistory)
	program := tea.NewProgram(model)

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
