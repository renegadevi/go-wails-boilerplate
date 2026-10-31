package main

import (
	"embed"
	"fmt"
	"os"

	"boilerplate/internal/applog"
	"boilerplate/internal/inspector"
	"boilerplate/internal/window"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	service := NewApp()
	app := application.New(application.Options{
		Name:        AppTitle,
		Description: AppDescription,
		// Dev Wails lines use the app log's 24-hour clock. Nil keeps production silent.
		Logger: applog.System(),
		Services: []application.Service{
			application.NewService(service),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	// An unset menu on macOS is replaced with the Wails default menu.
	// Templates with no menu still set an empty one.
	app.Menu.Set(window.Menu(app))
	windowOptions := window.Options(AppTitle)
	windowOptions.KeyBindings = inspector.KeyBindings()
	mainWindow := app.Window.NewWithOptions(windowOptions)
	mainWindow.OnWindowEvent(events.Common.WindowShow, func(*application.WindowEvent) {
		window.OnShow(mainWindow)
		inspector.EnableInspectElement(mainWindow)
	})

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}
