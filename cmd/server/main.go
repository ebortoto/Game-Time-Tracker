package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"game-time-tracker/internal/configuration"
	"game-time-tracker/internal/infrastructure/api"
	infrahistory "game-time-tracker/internal/infrastructure/history"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	if err := configuration.LoadDotEnv(".env"); err != nil {
		fmt.Println("Error loading .env:", err)
		return
	}

	addr := flag.String("addr", configuration.EnvOrDefault("TRACKER_SERVER_ADDR", ":8080"), "HTTP listen address")
	historyPath := flag.String("history-file", configuration.EnvOrDefault("TRACKER_HISTORY_FILE", "playtime_history.json"), "history storage file")
	apiKeyFlag := flag.String("api-key", configuration.EnvOrDefault("TRACKER_API_KEY", ""), "API key for client authentication (fallback: TRACKER_API_KEY)")
	flag.Parse()

	apiKey := strings.TrimSpace(*apiKeyFlag)
	historyBackend := configuration.EnvOrDefault("HISTORY_BACKEND", "json")
	if err := configuration.ValidateServerEnv(historyBackend); err != nil {
		fmt.Println("Error validating environment:", err)
		return
	}

	var repo api.HistoryStore
	switch strings.ToLower(strings.TrimSpace(historyBackend)) {
	case "mysql":
		migrationPath := configuration.EnvOrDefault("TRACKER_MIGRATION_FILE", "migrations/001_create_daily_history.sql")
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Local",
			configuration.EnvOrDefault("DB_USER", ""),
			configuration.EnvOrDefault("DB_PASSWORD", ""),
			configuration.EnvOrDefault("DB_HOST", ""),
			configuration.EnvOrDefault("DB_PORT", "3306"),
			configuration.EnvOrDefault("DB_NAME", "game_time_tracker"),
		)
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			fmt.Println("Error opening mysql connection:", err)
			return
		}
		defer db.Close()
		if err := db.Ping(); err != nil {
			fmt.Println("Error connecting to mysql:", err)
			return
		}
		if err := infrahistory.RunMySQLMigrations(db, migrationPath); err != nil {
			fmt.Println("Error running mysql migrations:", err)
			return
		}
		repo = infrahistory.NewMySQLRepository(db)
	default:
		repo = infrahistory.NewJSONRepository(*historyPath)
	}

	server := api.NewHistoryServer(apiKey, repo)

	slog.Info("history_server_starting", "addr", *addr, "history_file", *historyPath, "auth_enabled", apiKey != "")
	if err := http.ListenAndServe(*addr, server.Handler()); err != nil {
		slog.Error("history_server_failed", "error", err)
		fmt.Println("Server error:", err)
		return
	}
}
