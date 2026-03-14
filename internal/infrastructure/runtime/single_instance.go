package runtime

import (
	"errors"
	"net"
	"strings"
	"syscall"
)

const instanceLockAddr = "127.0.0.1:45731"

// AcquireSingleInstance prevents multiple tracker processes from running
// simultaneously, which would cause competing RTSS updates.
func AcquireSingleInstance() (release func(), alreadyRunning bool, err error) {
	ln, err := net.Listen("tcp", instanceLockAddr)
	if err != nil {
		if isAddrInUseError(err) {
			return nil, true, nil
		}
		return nil, false, err
	}
	if ln == nil {
		return nil, false, errors.New("failed to acquire instance lock")
	}

	release = func() {
		_ = ln.Close()
	}
	return release, false, nil
}

func isAddrInUseError(err error) bool {
	if errors.Is(err, syscall.EADDRINUSE) {
		return true
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "address already in use") {
		return true
	}
	// Windows localized bind error often appears as "only one usage of each socket address..."
	if strings.Contains(msg, "only one usage of each socket address") {
		return true
	}
	return false
}
