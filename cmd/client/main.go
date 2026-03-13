package main

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	apptracking "game-time-tracker/internal/application/tracking"
	"game-time-tracker/internal/configuration"
	infrahistory "game-time-tracker/internal/infrastructure/history"
	infraoverlay "game-time-tracker/internal/infrastructure/overlay"
	infraruntime "game-time-tracker/internal/infrastructure/runtime"
	infrascanner "game-time-tracker/internal/infrastructure/scanner"
	infratray "game-time-tracker/internal/infrastructure/tray"
	"game-time-tracker/internal/tui"
	"game-time-tracker/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

type overlayAdapter interface {
	Init()
	UpdateText(text string)
	Close()
}

type runtimeInitializer func() (infraruntime.StartupDiagnostics, error)

func main() {
	debug := flag.Bool("debug", false, "enable debug logs in tracker.log")
	serverURL := flag.String("server-url", "", "tracking server base URL (fallback: TRACKER_SERVER_URL)")
	apiKeyFlag := flag.String("api-key", "", "tracking server API key (fallback: TRACKER_API_KEY)")
	configPath := flag.String("config", "config.json", "tracker config file")
	overlayEnabled := flag.Bool("overlay", true, "enable RTSS overlay output")
	startHidden := flag.Bool("start-hidden", false, "start in tray without opening TUI")
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

	if code := runWindowsBootstrap(infraruntime.InitializeWindowsRuntime, os.Stdout); code != 0 {
		os.Exit(code)
	}

	trayService := infratray.NewService()
	openTUIReqCh := make(chan struct{}, 1)
	exitReqCh := make(chan struct{}, 1)

	var activeProgramMu sync.Mutex
	var activeProgram *tea.Program

	setActiveProgram := func(program *tea.Program) {
		activeProgramMu.Lock()
		defer activeProgramMu.Unlock()
		activeProgram = program
	}
	clearActiveProgram := func(program *tea.Program) {
		activeProgramMu.Lock()
		defer activeProgramMu.Unlock()
		if activeProgram == program {
			activeProgram = nil
		}
	}
	signalActiveProgram := func(signalName string) {
		activeProgramMu.Lock()
		program := activeProgram
		activeProgramMu.Unlock()
		if program != nil {
			program.Send(tui.SignalMsg{Signal: signalName})
		}
	}
	requestOpenTUI := func() {
		select {
		case openTUIReqCh <- struct{}{}:
		default:
		}
	}
	requestExit := func(signalName string) {
		signalActiveProgram(signalName)
		select {
		case exitReqCh <- struct{}{}:
		default:
		}
	}

	tuiBridge := infratray.NewTUIBridge(func() error {
		slog.Info("tray_show_requested")
		requestOpenTUI()
		return nil
	})
	if err := trayService.Start(); err != nil {
		slog.Error("tray_start_failed", "error", err)
		fmt.Println("Error starting tray service:", err)
		return
	}
	trayService.SetHandler(infratray.MenuActionOpenTUI, func() {
		opened, err := tuiBridge.Open()
		if err != nil {
			slog.Error("tray_menu_show_failed", "error", err)
			return
		}
		if opened {
			slog.Info("tray_menu_show")
			return
		}
		slog.Info("tray_menu_show_noop", "reason", "already_open")
	})
	trayService.SetHandler(infratray.MenuActionExit, func() {
		slog.Info("tray_menu_exit")
		requestExit("tray_exit")
	})
	defer func() {
		if err := trayService.Stop(); err != nil {
			slog.Error("tray_stop_failed", "error", err)
		}
	}()

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

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		for sig := range sigCh {
			slog.Info("os_signal_received", "signal", sig.String())
			requestExit(sig.String())
		}
	}()
	defer signal.Stop(sigCh)

	runTUI := func() error {
		releaseConsole, err := infraruntime.EnsureConsoleWindow()
		if err != nil {
			return err
		}
		defer releaseConsole()

		model := tui.NewModel(statusCh, historyCh, errCh, initialHistory)
		program := tea.NewProgram(model, tea.WithInput(os.Stdin), tea.WithOutput(os.Stdout))
		setActiveProgram(program)
		defer clearActiveProgram(program)
		fmt.Print("\033[H\033[2J")
		_, err = program.Run()
		return err
	}

	if !*startHidden {
		if opened, err := tuiBridge.Open(); err != nil {
			slog.Error("startup_show_failed", "error", err)
		} else if opened {
			slog.Info("startup_show")
		}
	}

	shutdownRequested := false
	for !shutdownRequested {
		select {
		case <-openTUIReqCh:
			if err := runTUI(); err != nil {
				slog.Error("tui_runtime_error", "error", err)
				fmt.Println("TUI runtime error:", err)
			}
			tuiBridge.MarkClosed()
			slog.Info("tui_closed_to_tray")
		case <-exitReqCh:
			shutdownRequested = true
		}
	}

	if err := runtime.Stop(); err != nil {
		slog.Error("shutdown_save_failed", "error", err)
		fmt.Println("Error saving history during shutdown:", err)
		return
	}
	slog.Info("shutdown_complete")
}

func runWindowsBootstrap(initialize runtimeInitializer, out io.Writer) int {
	diagnostics, err := initialize()
	if err != nil {
		slog.Error("windows_runtime_init_failed", "summary", diagnostics.Summary(), "error", err)
		for _, step := range diagnostics.Steps {
			slog.Error("windows_runtime_init_step", "name", step.Name, "ok", step.OK, "detail", step.Detail)
		}
		_, _ = fmt.Fprintln(out, "Error initializing Windows runtime:", err)
		return 1
	}

	slog.Info("windows_runtime_initialized", "summary", diagnostics.Summary())
	for _, step := range diagnostics.Steps {
		slog.Info("windows_runtime_step", "name", step.Name, "ok", step.OK, "detail", step.Detail)
	}
	if diagnostics.Platform != "windows" {
		msg := "Warning: Windows-only runtime features are unavailable on this platform. Running with compatibility fallback."
		slog.Warn("windows_runtime_fallback", "platform", diagnostics.Platform, "message", msg)
		_, _ = fmt.Fprintln(out, msg)
	}
	return 0
}
