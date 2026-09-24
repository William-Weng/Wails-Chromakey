package main

import (
	"embed"

	"log"

	"wails-chromakey/backend"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

//go:embed all:frontend/dist
var assets embed.FS

func init() {
}

func main() {

	app := application.New(application.Options{
		Name: "綠幕去背小工具",
		Services: []application.Service{
			application.NewService(&backend.ChromakeyService{}),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:          "綠幕去背小工具",
		Width:          370,
		Height:         600,
		MinWidth:       256,
		MinHeight:      560,
		EnableFileDrop: true, // 開啟檔案拖放
		URL:            "/",
	}).OnWindowEvent(events.Common.WindowFilesDropped, func(event *application.WindowEvent) {

		ctx := event.Context()
		files := ctx.DroppedFiles()

		if len(files) > 0 {
			application.Get().Event.Emit("image-file-dropped", files)
		}
	})

	err := app.Run()

	if err != nil {
		log.Fatal(err)
	}
}
