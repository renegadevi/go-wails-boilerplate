//go:build production && !devtools

// Package inspector gives dev builds the web inspector. Release builds have
// none: Wails compiles the devtools out, and the shortcut is not registered.
package inspector

import "github.com/wailsapp/wails/v3/pkg/application"

// KeyBindings returns no key bindings in release builds.
func KeyBindings() map[string]func(application.Window) { return nil }

// EnableInspectElement does nothing in release builds.
func EnableInspectElement(*application.WebviewWindow) {}
