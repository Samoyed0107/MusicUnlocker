package main

import (
	"context"
	"embed"
	"os"
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

const musicUnlockerUniqueID = "7b6d4d6e-5d4d-4b6e-9f52-4a9f2a1c7e31"

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
func (w *App) ValidateNamingConfig(c app.NamingConfigDTO) app.NamingValidation {
	return app.ValidateNamingConfig(c)
}
func (w *App) OpenOutputFolder(path string) error           { return w.facade.OpenOutputFolder(path) }
func (w *App) GetPendingLaunchPaths() *app.LaunchPathsEvent { return w.facade.GetPendingLaunchPaths() }
func (w *App) StartBatch(req app.BatchStartRequest) (app.BatchStartResponse, error) {
	return w.facade.StartBatch(req)
}
func (w *App) CancelBatch() error { return w.facade.CancelBatch() }
func (w *App) GetPreviewURL(batchID uint64, index int) (app.PreviewResponse, error) {
	return w.facade.GetPreviewURL(batchID, index)
}
func (w *App) shutdown(ctx context.Context) {
	if w.facade != nil {
		_ = w.facade.Shutdown(ctx)
	}
}

func (w *App) receiveLaunchArgs(args []string, workingDirectory string, source app.LaunchSource) {
	parsed := app.ParseLaunchPaths(args, workingDirectory)
	if len(parsed.Paths) == 0 {
		return
	}
	w.facade.ReceiveLaunchPaths(w.facadeLaunchEvent(parsed.Paths, source))
}

func (w *App) facadeLaunchEvent(paths []string, source app.LaunchSource) app.LaunchPathsEvent {
	return app.LaunchPathsEvent{Paths: paths, Source: source}
}

func main() {
	application := newWailsApp()
	workingDirectory, _ := os.Getwd()
	initialArgs := append([]string(nil), os.Args[1:]...)
	appOptions := &options.App{Title: "MusicUnlocker", Width: 1100, Height: 760, MinWidth: 900, MinHeight: 600, AssetServer: &assetserver.Options{Assets: assets}, BackgroundColour: &options.RGBA{R: 248, G: 250, B: 252, A: 1}, DragAndDrop: &options.DragAndDrop{EnableFileDrop: true, DisableWebViewDrop: true}, OnStartup: func(ctx context.Context) {
		application.startup(ctx)
		application.receiveLaunchArgs(initialArgs, workingDirectory, app.LaunchSourceInitial)
	}, OnShutdown: application.shutdown, Bind: []interface{}{application}, SingleInstanceLock: &options.SingleInstanceLock{UniqueId: musicUnlockerUniqueID, OnSecondInstanceLaunch: func(data options.SecondInstanceData) {
		if application.facade == nil {
			return
		}
		application.receiveLaunchArgs(data.Args, data.WorkingDirectory, app.LaunchSourceSecond)
		if application.ctx != nil {
			runtime.WindowUnminimise(application.ctx)
			runtime.WindowShow(application.ctx)
		}
	}}}
	if err := wails.Run(appOptions); err != nil {
		println("Error:", err.Error())
	}
}
