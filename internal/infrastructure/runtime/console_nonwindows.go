//go:build !windows

package runtime

// EnsureConsoleWindow is a no-op outside Windows.
func EnsureConsoleWindow() (func(), error) {
	return func() {}, nil
}
