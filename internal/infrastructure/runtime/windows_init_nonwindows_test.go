//go:build !windows

package runtime

import "testing"

func TestInitializeWindowsRuntimeNonWindows(t *testing.T) {
	diagnostics, err := InitializeWindowsRuntime()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if diagnostics.Platform != "non-windows" {
		t.Fatalf("expected platform non-windows, got %s", diagnostics.Platform)
	}
	if len(diagnostics.Steps) == 0 {
		t.Fatalf("expected at least one diagnostic step")
	}
	if diagnostics.Steps[0].Name != "windows-runtime" {
		t.Fatalf("expected windows-runtime step, got %s", diagnostics.Steps[0].Name)
	}
	if !diagnostics.Steps[0].OK {
		t.Fatalf("expected step to be successful")
	}
}
