//go:build windows

package tray

import _ "embed"

//go:embed tracker.ico
var defaultTrayIcon []byte
