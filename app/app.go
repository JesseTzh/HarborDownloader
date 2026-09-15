package app

import (
	"context"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"HarborDownloader/internal/config"
	"HarborDownloader/internal/credential"
	"HarborDownloader/internal/download"
	"HarborDownloader/internal/logger"
	"HarborDownloader/internal/model"
)

type App struct {
	ctx     context.Context
	cancel  context.CancelFunc
	version string
	commit  string
	buildAt string

	cfg   *config.Service
	cred  *credential.Store
	log   *logger.Logger
	tasks *download.Service
}

func New(version, commit, buildAt string) *App {
	return &App{
		version: version,
		commit:  commit,
		buildAt: buildAt,
		cred:    credential.New(),
	}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx, a.cancel = context.WithCancel(ctx)
	log, err := logger.New()
	if err == nil {
		a.log = log
		a.log.Info("Harbor Downloader starting version " + a.version)
	}
	cfg, err := config.New()
	if err == nil {
		a.cfg = cfg
		saved := cfg.Load()
		if saved.Registry != "" {
			a.cred.Set(saved.Registry, saved.Username, saved.Password)
		}
	}
	a.tasks = download.NewService(a.ctx, a.log, &eventBridge{app: a})
}

func (a *App) Shutdown(ctx context.Context) {
	_ = ctx
	if a.cancel != nil {
		a.cancel()
	}
	if a.cred != nil {
		a.cred.Clear()
	}
	if a.log != nil {
		a.log.Info("Harbor Downloader shutdown")
	}
}

func (a *App) emitLog(msg string) {
	if a.log != nil {
		a.log.Info(msg)
	}
}

type eventBridge struct {
	app *App
}

func (b *eventBridge) Progress(ev model.ProgressEvent) {
	if b.app == nil || b.app.ctx == nil {
		return
	}
	wailsruntime.EventsEmit(b.app.ctx, EventProgress, ev)
}

func (b *eventBridge) Status(task model.DownloadTask) {
	if b.app == nil || b.app.ctx == nil {
		return
	}
	wailsruntime.EventsEmit(b.app.ctx, EventStatus, task)
}

func (b *eventBridge) Error(ev model.AppErrorEvent) {
	if b.app == nil || b.app.ctx == nil {
		return
	}
	wailsruntime.EventsEmit(b.app.ctx, EventError, ev)
}

func (b *eventBridge) Completed(ev model.DownloadCompleteEvent) {
	if b.app == nil || b.app.ctx == nil {
		return
	}
	wailsruntime.EventsEmit(b.app.ctx, EventCompleted, ev)
}

func (b *eventBridge) JobStatus(ev model.JobStatusEvent) {
	if b.app == nil || b.app.ctx == nil {
		return
	}
	wailsruntime.EventsEmit(b.app.ctx, EventJobStatus, ev)
}

func (b *eventBridge) JobCompleted(ev model.JobCompleteEvent) {
	if b.app == nil || b.app.ctx == nil {
		return
	}
	wailsruntime.EventsEmit(b.app.ctx, EventJobCompleted, ev)
}
