//go:build !windows

package detector

func getForegroundPID() (uint32, error) {
	return 0, nil
}

func DebugGetForegroundPID() (uint32, error) {
	return getForegroundPID()
}
