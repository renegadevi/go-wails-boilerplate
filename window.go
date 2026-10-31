package main

import (
	goruntime "runtime"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// windowBackground matches --color-bg in frontend/src/style.css.
// Linux paints the GTK window with it. A light theme otherwise leaves a
// pale line along the top of a frameless window.
const (
	windowBackgroundR uint8 = 20
	windowBackgroundG uint8 = 22
	windowBackgroundB uint8 = 24
)

// hideLinuxFramelessEdge paints over that line. Other systems leave it empty.
// window_linux.go fills it in.
var hideLinuxFramelessEdge = func(*application.WebviewWindow) {}

const (
	// WindowFrame is a system window with no menu.
	WindowFrame = "frame"
	// WindowFrameMenu is a system window with the native menu.
	// This is the cross-platform default.
	WindowFrameMenu = "frame-menu"
	// WindowFramelessMenu is a custom toolbar plus the native menu where the
	// system can show one. macOS keeps that menu in the screen menu bar.
	// Windows and Linux draw minimise, maximise, and close in the toolbar,
	// because the native menu is part of the frame.
	WindowFramelessMenu = "frameless-menu"
	// WindowFrameless is a custom toolbar and no native menu, on every system.
	WindowFrameless = "frameless"
)

// window is the window template. Pick one of the constants above.
const window = WindowFramelessMenu

// windowFrameless is true when the frontend draws the toolbar.
const windowFrameless = window == WindowFrameless || window == WindowFramelessMenu

// nativeFrameless is true when the OS window frame itself is removed.
// On macOS the native frame carries the rounded corners and the traffic
// lights, so we keep the frame there and hide its title bar through
// Mac.TitleBar instead of removing the frame entirely.
const nativeFrameless = windowFrameless && goruntime.GOOS != "darwin"

// showNativeMenu reports whether this window carries a native menu.
// A frameless window on Windows or Linux has no frame to hang that menu on.
func showNativeMenu() bool {
	switch window {
	case WindowFrameMenu:
		return true
	case WindowFramelessMenu:
		return goruntime.GOOS == "darwin"
	default:
		return false
	}
}

// installMacTrafficLights centers the macOS traffic lights in the custom toolbar.
// Windows and Linux leave it empty. window_darwin.go fills it in, because that
// adjustment is Objective-C and cannot be compiled for the other systems.
var installMacTrafficLights = func() {}

func positionDarwinTrafficLights() {
	if !windowFrameless {
		return
	}
	installMacTrafficLights()
}

// Frameless reports whether the frontend should draw the toolbar.
func (a *App) Frameless() bool { return windowFrameless }

func windowOptions() application.WebviewWindowOptions {
	opts := application.WebviewWindowOptions{
		Name:             "main",
		Title:            AppTitle,
		Width:            1200,
		Height:           800,
		MinWidth:         800,
		MinHeight:        600,
		BackgroundColour: application.NewRGB(windowBackgroundR, windowBackgroundG, windowBackgroundB),
		Frameless:        nativeFrameless,
		URL:              "/",
		// A dark window lets Windows paint a dark menu bar when the system
		// app theme is dark. A light system theme keeps a light menu bar so
		// the native menu text stays readable.
		Windows: application.WindowsWindow{
			Theme: application.Dark,
		},
		// The webview fills newly exposed edges with its own colour while a
		// resize or un-maximise catches up: white in the light appearance,
		// near BackgroundColour in the dark one. CSS cannot reach it.
		Mac: application.MacWindow{
			Appearance: application.NSAppearanceNameDarkAqua,
		},
		// Windows and Linux hang the menu on the frame. macOS uses the
		// screen menu bar, so this flag does nothing there.
		UseApplicationMenu: showNativeMenu(),
		// Only WebView2 reads this, and only in release builds. The frontend
		// applies the mode on every system, dev builds included.
		DefaultContextMenuDisabled: contextMenu != ContextMenuDefault,
		KeyBindings:                inspectorKeyBindings(),
	}
	if windowFrameless {
		// Keep the traffic lights. window_darwin.go moves them into the
		// taller custom toolbar. A framed window keeps the system title bar.
		opts.Mac.TitleBar = application.MacTitleBarHidden
	}
	return opts
}
