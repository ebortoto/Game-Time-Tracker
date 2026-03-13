//go:build windows

package runtime

import (
	"fmt"
	"os"
	"syscall"
)

var (
	kernel32DLL      = syscall.NewLazyDLL("kernel32.dll")
	procAllocConsole = kernel32DLL.NewProc("AllocConsole")
	procFreeConsole  = kernel32DLL.NewProc("FreeConsole")
	procGetConsole   = kernel32DLL.NewProc("GetConsoleWindow")
)

// EnsureConsoleWindow ensures the process has a console attached and rewires stdio.
func EnsureConsoleWindow() (func(), error) {
	hwnd, _, _ := procGetConsole.Call()
	if hwnd != 0 {
		return func() {}, nil
	}

	ret, _, err := procAllocConsole.Call()
	if ret == 0 {
		if err != nil && err != syscall.Errno(0) {
			return nil, fmt.Errorf("alloc console failed: %w", err)
		}
		return nil, fmt.Errorf("alloc console failed")
	}

	oldIn := os.Stdin
	oldOut := os.Stdout
	oldErr := os.Stderr

	in, err := os.OpenFile("CONIN$", os.O_RDWR, 0)
	if err != nil {
		_, _, _ = procFreeConsole.Call()
		return nil, fmt.Errorf("open CONIN$ failed: %w", err)
	}
	out, err := os.OpenFile("CONOUT$", os.O_RDWR, 0)
	if err != nil {
		_ = in.Close()
		_, _, _ = procFreeConsole.Call()
		return nil, fmt.Errorf("open CONOUT$ failed: %w", err)
	}

	os.Stdin = in
	os.Stdout = out
	os.Stderr = out

	cleanup := func() {
		os.Stdin = oldIn
		os.Stdout = oldOut
		os.Stderr = oldErr
		_ = in.Close()
		_ = out.Close()
		_, _, _ = procFreeConsole.Call()
	}
	return cleanup, nil
}
