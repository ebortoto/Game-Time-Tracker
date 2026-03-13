package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	infraruntime "game-time-tracker/internal/infrastructure/runtime"
)

func TestRunWindowsBootstrapSuccess(t *testing.T) {
	out := &bytes.Buffer{}
	initializer := func() (infraruntime.StartupDiagnostics, error) {
		return infraruntime.StartupDiagnostics{
			Platform: "windows",
			Steps: []infraruntime.DiagnosticStep{
				{Name: "windows-runtime", OK: true, Detail: "ok"},
			},
		}, nil
	}

	code := runWindowsBootstrap(initializer, out)
	if code != 0 {
		t.Fatalf("expected success exit code 0, got %d", code)
	}
	if out.Len() != 0 {
		t.Fatalf("expected no stdout output on success, got %q", out.String())
	}
}

func TestRunWindowsBootstrapFailureReturnsNonZero(t *testing.T) {
	out := &bytes.Buffer{}
	initializer := func() (infraruntime.StartupDiagnostics, error) {
		return infraruntime.StartupDiagnostics{
			Platform: "windows",
			Steps: []infraruntime.DiagnosticStep{
				{Name: "foreground-window-access", OK: false, Detail: "missing capability"},
			},
		}, errors.New("bootstrap failed")
	}

	code := runWindowsBootstrap(initializer, out)
	if code != 1 {
		t.Fatalf("expected failure exit code 1, got %d", code)
	}
	if !strings.Contains(out.String(), "Error initializing Windows runtime:") {
		t.Fatalf("expected error message in stdout, got %q", out.String())
	}
}

func TestRunWindowsBootstrapNonWindowsFallbackMessage(t *testing.T) {
	out := &bytes.Buffer{}
	initializer := func() (infraruntime.StartupDiagnostics, error) {
		return infraruntime.StartupDiagnostics{
			Platform: "non-windows",
			Steps: []infraruntime.DiagnosticStep{
				{Name: "windows-runtime", OK: true, Detail: "skipped"},
			},
		}, nil
	}

	code := runWindowsBootstrap(initializer, out)
	if code != 0 {
		t.Fatalf("expected success exit code 0, got %d", code)
	}
	if !strings.Contains(out.String(), "Windows-only runtime features are unavailable") {
		t.Fatalf("expected fallback warning, got %q", out.String())
	}
}
