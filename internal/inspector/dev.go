//go:build !production || devtools

// Package inspector gives dev builds the web inspector: a hidden shortcut on
// every system, and Inspect Element in the right-click menu on macOS.
// Release builds compile production.go instead, which has neither.
package inspector

import (
	goruntime "runtime"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// shortcut opens the web inspector in dev builds. It is meant for developers,
// so it avoids the browser and Electron defaults that users try.
func shortcut() string {
	if goruntime.GOOS == "darwin" {
		return "Cmd+Option+Shift+I"
	}
	return "Ctrl+Shift+Alt+I"
}

// openInspector opens the web inspector. Wails opens it on Windows and Linux.
// darwin_dev.go replaces it, because Wails only does that on macOS with
// -tags private_mac_apis.
var openInspector = func(window application.Window) { window.OpenDevTools() }

// enableInspectElement adds Inspect Element to the webview menu. Wails does
// that on Windows and Linux. darwin_dev.go fills it in for macOS.
var enableInspectElement = func(*application.WebviewWindow) {}

// KeyBindings is the window key binding that opens the inspector.
func KeyBindings() map[string]func(application.Window) {
	return map[string]func(application.Window){
		shortcut(): func(window application.Window) { openInspector(window) },
	}
}

// EnableInspectElement adds Inspect Element to the right-click menu. Call it
// once the window is shown.
func EnableInspectElement(window *application.WebviewWindow) {
	enableInspectElement(window)
}
