//go:build windows

package runtime

import (
	"fmt"

	"game-time-tracker/internal/detector"
)

// InitializeWindowsRuntime validates Windows-specific runtime dependencies.
func InitializeWindowsRuntime() (StartupDiagnostics, error) {
	diagnostics := StartupDiagnostics{Platform: "windows"}

	diagnostics.Steps = append(diagnostics.Steps, DiagnosticStep{
		Name:   "windows-runtime",
		OK:     true,
		Detail: "windows startup path selected",
	})

	_, err := detector.DebugGetForegroundPID()
	if err != nil {
		diagnostics.Steps = append(diagnostics.Steps, DiagnosticStep{
			Name:   "foreground-window-access",
			OK:     false,
			Detail: err.Error(),
		})
		return diagnostics, fmt.Errorf("windows initialization failed: %w", err)
	}

	diagnostics.Steps = append(diagnostics.Steps, DiagnosticStep{
		Name:   "foreground-window-access",
		OK:     true,
		Detail: "GetForegroundWindow/GetWindowThreadProcessId available",
	})
	return diagnostics, nil
}
