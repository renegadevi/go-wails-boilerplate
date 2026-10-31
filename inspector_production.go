//go:build production && !devtools

package main

import "github.com/wailsapp/wails/v3/pkg/application"

// Release builds have no inspector. Wails compiles the devtools out, and the
// shortcut is not registered.
func inspectorKeyBindings() map[string]func(application.Window) { return nil }

var enableInspectElement = func(*application.WebviewWindow) {}
