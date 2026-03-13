package runtime

import "fmt"

// StartupDiagnostics captures bootstrap checks executed at startup.
type StartupDiagnostics struct {
	Platform string
	Steps    []DiagnosticStep
}

// DiagnosticStep represents one startup validation step.
type DiagnosticStep struct {
	Name   string
	OK     bool
	Detail string
}

func (d StartupDiagnostics) Summary() string {
	status := "ok"
	for _, step := range d.Steps {
		if !step.OK {
			status = "error"
			break
		}
	}
	return fmt.Sprintf("platform=%s checks=%d status=%s", d.Platform, len(d.Steps), status)
}
