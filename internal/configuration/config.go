package configuration

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config holds runtime settings loaded from disk.
type Config struct {
	WatchedProcesses []string `json:"watchedProcesses"`
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	if len(cfg.WatchedProcesses) == 0 {
		return Config{}, fmt.Errorf("config watchedProcesses must not be empty")
	}

	return cfg, nil
}
