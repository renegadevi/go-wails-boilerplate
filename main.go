package main

import (
	"embed"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	service := NewApp()
	app := application.New(application.Options{
		Name:        AppTitle,
		Description: AppDescription,
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
	app.Window.NewWithOptions(windowOptions())

	err := app.Run()
	if err != nil {
		println("Error:", err.Error())
	}
}
