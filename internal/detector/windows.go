package detector

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32                       = windows.NewLazySystemDLL("user32.dll")
	procGetForegroundWindow      = user32.NewProc("GetForegroundWindow")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
)

// getForegroundPID returns the process ID (PID) for the currently focused window.
func getForegroundPID() (uint32, error) {
	// 1. Get the handle (identifier) of the active window.
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return 0, nil // No window in focus.
	}

	// 2. Ask Windows for the PID that owns this handle.
	var pid uint32
	// The function returns a ThreadID, but it writes into the 'pid' pointer argument.
	procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))

	return pid, nil
}

// DebugGetForegroundPID is only a public wrapper for tests in main.
func DebugGetForegroundPID() (uint32, error) {
	return getForegroundPID()
}
