//go:build !windows

package runtime

// InitializeWindowsRuntime is a no-op outside Windows and returns diagnostics.
func InitializeWindowsRuntime() (StartupDiagnostics, error) {
	diagnostics := StartupDiagnostics{Platform: "non-windows"}
	diagnostics.Steps = append(diagnostics.Steps, DiagnosticStep{
		Name:   "windows-runtime",
		OK:     true,
		Detail: "skipped on non-windows platform",
	})
	return diagnostics, nil
}
