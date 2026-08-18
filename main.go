package main

import (
	"context"
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"musicunlocker/internal/app"
	"musicunlocker/internal/batch"
)

//go:embed all:frontend/dist
var assets embed.FS

type runtimeSink struct{ ctx context.Context }

func (s runtimeSink) Emit(name string, payload any) { runtime.EventsEmit(s.ctx, name, payload) }

type runtimeDialogs struct{ ctx context.Context }

func (d runtimeDialogs) SelectFiles() ([]string, error) {
	return runtime.OpenMultipleFilesDialog(d.ctx, runtime.OpenDialogOptions{Title: "Select music files"})
}
func (d runtimeDialogs) SelectFolder() (string, error) {
	return runtime.OpenDirectoryDialog(d.ctx, runtime.OpenDialogOptions{Title: "Select music folder"})
}
func (d runtimeDialogs) SelectOutputFolder() (string, error) {
	return runtime.OpenDirectoryDialog(d.ctx, runtime.OpenDialogOptions{Title: "Select output folder"})
}

type appRunner struct{ engine batch.Engine }

func (r appRunner) Run(ctx context.Context, req batch.Request) (batch.Summary, error) {
	return r.engine.Run(ctx, req)
}
func (r appRunner) RunWithObserver(ctx context.Context, req batch.Request, observer batch.Observer) (batch.Summary, error) {
	r.engine.Observer = observer
	return r.engine.Run(ctx, req)
}

type App struct {
	facade *app.Application
	ctx    context.Context
}

func newWailsApp() *App { return &App{} }
func (w *App) startup(ctx context.Context) {
	w.ctx = ctx
	w.facade = app.New(appRunner{engine: batch.Engine{}}, runtimeSink{ctx}, runtimeDialogs{ctx})
}
func (w *App) GetAppInfo() app.AppInfo             { return w.facade.GetAppInfo() }
func (w *App) SelectFiles() ([]string, error)      { return w.facade.SelectFiles() }
func (w *App) SelectFolder() (string, error)       { return w.facade.SelectFolder() }
func (w *App) SelectOutputFolder() (string, error) { return w.facade.SelectOutputFolder() }
func (w *App) StartBatch(req app.BatchStartRequest) (app.BatchStartResponse, error) {
	return w.facade.StartBatch(req)
}
func (w *App) CancelBatch() error { return w.facade.CancelBatch() }

func main() {
	application := newWailsApp()
	if err := wails.Run(&options.App{Title: "MusicUnlocker", Width: 1100, Height: 760, MinWidth: 900, MinHeight: 600, AssetServer: &assetserver.Options{Assets: assets}, BackgroundColour: &options.RGBA{R: 248, G: 250, B: 252, A: 1}, OnStartup: application.startup, Bind: []interface{}{application}}); err != nil {
		println("Error:", err.Error())
	}
}
