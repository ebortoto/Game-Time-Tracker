package overlay

import "game-time-tracker/internal/ui"

// RTSSOverlay adapts the RTSS shared-memory writer to the application OverlayWriter port.
type RTSSOverlay struct{}

func NewRTSSOverlay() *RTSSOverlay {
	return &RTSSOverlay{}
}

func (o *RTSSOverlay) Init() {
	ui.InitOverlay()
}

func (o *RTSSOverlay) UpdateText(text string) {
	ui.UpdateText(text)
}

func (o *RTSSOverlay) Close() {
	ui.CloseOverlay()
}
