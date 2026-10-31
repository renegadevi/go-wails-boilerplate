package main

import (
	"context"

	"boilerplate/internal/applog"
	"boilerplate/internal/window"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// App is the bridge the window calls through the generated Wails bindings.
// Add your own exported methods here and Wails will expose them to the frontend.
type App struct{}

// NewApp builds the process before the window opens.
func NewApp() *App {
	return &App{}
}

// Version is the app version shown in the UI.
func (a *App) Version() string {
	return AppVersion
}

// Greet is a minimal example of a bound method. Replace it with your own API.
func (a *App) Greet(name string) string {
	if name == "" {
		name = "world"
	}
	return "Hello, " + name + "!"
}

// Frameless reports whether the frontend should draw the toolbar.
func (a *App) Frameless() bool { return window.Frameless() }

// ContextMenu reports the right-click mode so the frontend can apply it.
func (a *App) ContextMenu() string { return window.ContextMenu() }

func (a *App) ServiceStartup(context.Context, application.ServiceOptions) error {
	applog.Init()
	window.PositionTrafficLights()
	applog.Pre{}.Info("ready", map[string]any{"version": AppVersion})
	return nil
}
