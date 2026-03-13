package tray

import "github.com/getlantern/systray"

type systrayRuntime struct{}

type systrayMenuItem struct {
	item *systray.MenuItem
}

func newSystrayRuntime() runtimeAdapter {
	return &systrayRuntime{}
}

func (r *systrayRuntime) Run(onReady func(), onExit func()) {
	systray.Run(onReady, onExit)
}

func (r *systrayRuntime) Quit() {
	systray.Quit()
}

func (r *systrayRuntime) SetTooltip(text string) {
	systray.SetTooltip(text)
}

func (r *systrayRuntime) SetIcon(icon []byte) {
	systray.SetIcon(icon)
}

func (r *systrayRuntime) AddMenuItem(title string, tooltip string) runtimeMenuItem {
	return &systrayMenuItem{item: systray.AddMenuItem(title, tooltip)}
}

func (m *systrayMenuItem) ClickedChannel() <-chan struct{} {
	return m.item.ClickedCh
}
