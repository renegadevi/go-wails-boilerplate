package main

import "github.com/wailsapp/wails/v3/pkg/application"

// applicationMenu builds the native application menu. Which window template shows it
// is chosen by window in window.go.
//
// macOS gets the App menu from Wails, plus Edit and Window. Windows and
// Linux get Edit and Window. Edit calls the webview for undo, cut, copy,
// paste, and select all. A template with no menu returns an empty menu.
// On macOS, leaving the menu unset makes Wails install its own default.
func applicationMenu(app *application.App) *application.Menu {
	root := app.NewMenu()
	if !showNativeMenu() {
		return root
	}
	root.AddRole(application.AppMenu)
	root.AddRole(application.EditMenu)
	root.AddRole(application.WindowMenu)
	return root
}
