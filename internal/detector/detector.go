package detector

import (
	"fmt"
	"sort"
	"strings"

	"github.com/shirou/gopsutil/v3/process"
)

type Config struct {
	TargetGames []string
}

func New(games []string) *Config {
	for i, g := range games {
		games[i] = strings.ToLower(g)
	}
	sort.Strings(games)
	return &Config{TargetGames: games}
}

// Scan returns: (found bool, name string, focused bool)
func (c *Config) Scan() (bool, string, bool) {
	// 1. Find the PID of the window currently in focus.
	focusedPID, err := getForegroundPID()
	if err != nil {
		fmt.Println("Error getting focused window:", err)
	}

	// 1.1 Resolve the focused process first to avoid flapping
	// when there are multiple instances/helper processes for the same game.
	if focusedPID != 0 {
		focusedProc, err := process.NewProcess(int32(focusedPID))
		if err == nil {
			focusedName, err := focusedProc.Name()
			if err == nil {
				focusedNameLower := strings.ToLower(focusedName)
				for _, game := range c.TargetGames {
					if focusedNameLower == game {
						return true, focusedName, true
					}
				}
			}
		}
	}

	processos, err := process.Processes()
	if err != nil {
		return false, "", false
	}

	// Since the focused process was already validated above, now we only need to
	// check whether any target game is running (paused state).
	foundGame := false
	foundName := ""

	for _, p := range processos {
		name, err := p.Name()
		if err != nil {
			continue
		}

		nameLower := strings.ToLower(name)

		for _, game := range c.TargetGames {
			if nameLower == game {
				foundGame = true
				if foundName == "" {
					foundName = name
				}
				break
			}
		}
	}

	if foundGame {
		return true, foundName, false
	}

	return false, "", false
}
