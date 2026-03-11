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

var (
	procSetWindowPos = user32.NewProc("SetWindowPos")
)

// Windows constants used to control window z-order and behavior.
const (
	HWND_TOPMOST uintptr = ^uintptr(0)
	SWP_NOSIZE           = 0x0001
	SWP_NOMOVE           = 0x0002
)

// SetAlwaysOnTop forces a window (by title) to stay above other windows.
func SetAlwaysOnTop(windowTitle string) {
	// 1. Find the window by title.
	// Note: Fyne creates windows with the title we define.
	// We need to convert the Go string to UTF-16 pointer.
	windowTitlePtr, _ := windows.UTF16PtrFromString(windowTitle)

	hwnd, _, _ := user32.NewProc("FindWindowW").Call(
		0,
		uintptr(unsafe.Pointer(windowTitlePtr)),
	)

	if hwnd == 0 {
		return // Window not found yet.
	}

	// 2. Apply the "TopMost" flag.
	procSetWindowPos.Call(
		hwnd,
		uintptr(HWND_TOPMOST),
		0, 0, 0, 0,
		uintptr(SWP_NOMOVE|SWP_NOSIZE), // Keep current size and position, change z-order only.
	)
}
