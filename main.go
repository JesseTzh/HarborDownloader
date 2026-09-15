package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"

	"HarborDownloader/app"
)

//go:embed all:frontend/dist
var assets embed.FS

var (
	Version = "dev"
	Commit  = "unknown"
	BuildAt = "unknown"
)

func main() {
	application := app.New(Version, Commit, BuildAt)

	err := wails.Run(&options.App{
		Title:     "Harbor Downloader",
		Width:     880,
		Height:    860,
		MinWidth:  760,
		MinHeight: 700,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 244, G: 245, B: 247, A: 255},
		OnStartup:        application.Startup,
		OnShutdown:       application.Shutdown,
		Bind: []interface{}{
			application,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}
