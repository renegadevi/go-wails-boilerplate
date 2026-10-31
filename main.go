package main

import (
	"embed"
	"fmt"
	"os"

	"boilerplate/internal/applog"

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
	app.Menu.Set(applicationMenu(app))
	window := app.Window.NewWithOptions(windowOptions())
	// The GTK window exists by the time it is shown. Earlier than that,
	// NativeWindow is still nil.
	window.OnWindowEvent(events.Common.WindowShow, func(*application.WindowEvent) {
		hideLinuxFramelessEdge(window)
	})

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}
