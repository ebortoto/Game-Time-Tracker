package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"game-time-tracker/internal/infrastructure/api"
	infrahistory "game-time-tracker/internal/infrastructure/history"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	historyPath := flag.String("history-file", "playtime_history.json", "history storage file")
	apiKeyFlag := flag.String("api-key", "", "API key for client authentication (fallback: TRACKER_API_KEY)")
	flag.Parse()

	apiKey := strings.TrimSpace(*apiKeyFlag)
	if apiKey == "" {
		apiKey = strings.TrimSpace(os.Getenv("TRACKER_API_KEY"))
	}

	repo := infrahistory.NewJSONRepository(*historyPath)
	server := api.NewHistoryServer(apiKey, repo)

	slog.Info("history_server_starting", "addr", *addr, "history_file", *historyPath, "auth_enabled", apiKey != "")
	if err := http.ListenAndServe(*addr, server.Handler()); err != nil {
		slog.Error("history_server_failed", "error", err)
		fmt.Println("Server error:", err)
		os.Exit(1)
	}
}
