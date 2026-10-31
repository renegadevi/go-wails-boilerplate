//go:build !production || devtools

package main

import (
	goruntime "runtime"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// inspectorShortcut opens the web inspector in dev builds. It is meant for
// developers, so it avoids the browser and Electron defaults that users try.
func inspectorShortcut() string {
	if goruntime.GOOS == "darwin" {
		return "Cmd+Option+Shift+I"
	}
	return "Ctrl+Shift+Alt+I"
}

// openInspector opens the web inspector. Wails opens it on Windows and Linux.
// inspector_darwin_dev.go replaces it, because Wails only does that on macOS
// with -tags private_mac_apis.
var openInspector = func(window application.Window) { window.OpenDevTools() }

// enableInspectElement adds Inspect Element to the webview menu. Wails does
// that on Windows and Linux. inspector_darwin_dev.go fills it in for macOS.
var enableInspectElement = func(*application.WebviewWindow) {}

func inspectorKeyBindings() map[string]func(application.Window) {
	return map[string]func(application.Window){
		inspectorShortcut(): func(window application.Window) { openInspector(window) },
	}
}
