package runtime

import (
	"errors"
	"net"
	"strings"
)

const instanceLockAddr = "127.0.0.1:45731"

// AcquireSingleInstance prevents multiple tracker processes from running
// simultaneously, which would cause competing RTSS updates.
func AcquireSingleInstance() (release func(), alreadyRunning bool, err error) {
	ln, err := net.Listen("tcp", instanceLockAddr)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "address already in use") {
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
