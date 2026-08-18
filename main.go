package main

import (
	"context"
	"embed"
	"os/exec"

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

type explorerOpener struct{}

func (explorerOpener) Open(path string) error { return exec.Command("explorer.exe", path).Start() }
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
	w.facade = app.NewWithDependencies(appRunner{engine: batch.Engine{}}, runtimeSink{ctx}, runtimeDialogs{ctx}, app.DefaultPreferencesStore(), explorerOpener{})
}
func (w *App) GetAppInfo() app.AppInfo             { return w.facade.GetAppInfo() }
func (w *App) SelectFiles() ([]string, error)      { return w.facade.SelectFiles() }
func (w *App) SelectFolder() (string, error)       { return w.facade.SelectFolder() }
func (w *App) SelectOutputFolder() (string, error) { return w.facade.SelectOutputFolder() }
func (w *App) SuggestOutputFolder(inputs []string) (string, error) {
	return w.facade.SuggestOutputFolder(inputs)
}
func (w *App) DescribeInputs(inputs []string) []app.SelectedInput {
	return w.facade.DescribeInputs(inputs)
}
func (w *App) GetPreferences() (app.Preferences, error) { return w.facade.GetPreferences() }
func (w *App) SavePreferences(p app.Preferences) error  { return w.facade.SavePreferences(p) }
func (w *App) OpenOutputFolder(path string) error       { return w.facade.OpenOutputFolder(path) }
func (w *App) StartBatch(req app.BatchStartRequest) (app.BatchStartResponse, error) {
	return w.facade.StartBatch(req)
}
func (w *App) CancelBatch() error { return w.facade.CancelBatch() }

func main() {
	application := newWailsApp()
	if err := wails.Run(&options.App{Title: "MusicUnlocker", Width: 1100, Height: 760, MinWidth: 900, MinHeight: 600, AssetServer: &assetserver.Options{Assets: assets}, BackgroundColour: &options.RGBA{R: 248, G: 250, B: 252, A: 1}, DragAndDrop: &options.DragAndDrop{EnableFileDrop: true, DisableWebViewDrop: true}, OnStartup: application.startup, Bind: []interface{}{application}}); err != nil {
		println("Error:", err.Error())
	}
}
