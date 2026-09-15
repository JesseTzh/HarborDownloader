package download

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/uuid"

	"HarborDownloader/internal/apperr"
	"HarborDownloader/internal/disk"
	"HarborDownloader/internal/export"
	"HarborDownloader/internal/filename"
	"HarborDownloader/internal/layercache"
	"HarborDownloader/internal/logger"
	"HarborDownloader/internal/model"
	"HarborDownloader/internal/registry"
)

type Emitter interface {
	Progress(model.ProgressEvent)
	Status(model.DownloadTask)
	Error(model.AppErrorEvent)
	Completed(model.DownloadCompleteEvent)
	JobStatus(model.JobStatusEvent)
	JobCompleted(model.JobCompleteEvent)
}

type progressMeta struct {
	RunID   string
	JobName string
	Image   string
	Index   int
	Total   int
}

type Service struct {
	store  *Store
	log    *logger.Logger
	emit   Emitter
	parent context.Context
}

func NewService(parent context.Context, log *logger.Logger, emit Emitter) *Service {
	return &Service{
		store:  NewStore(),
		log:    log,
		emit:   emit,
		parent: parent,
	}
}

func (s *Service) Get(id string) (model.DownloadTask, bool) {
	return s.store.Get(id)
}

func (s *Service) Busy() bool {
	return s != nil && s.store.ActiveID() != ""
}

func (s *Service) Cancel(id string) error {
	s.log.Info("Cancel requested " + id)
	return s.store.Cancel(id)
}

func (s *Service) Start(req model.DownloadRequest) (string, error) {
	if s.store.ActiveID() != "" {
		return "", apperr.New(apperr.CodeInvalid, "已有下载任务正在进行，MVP 同时只能下载一个任务。", nil)
	}
	platform, err := registry.ParsePlatform(req.Platform)
	if err != nil {
		return "", err
	}
	parsed, _, err := registry.ParseImageReference(req.Image, req.Registry, req.Insecure)
	if err != nil {
		return "", err
	}
	if req.OutputDir == "" {
		return "", apperr.New(apperr.CodeInvalid, "请选择输出目录。", nil)
	}
	if err := os.MkdirAll(req.OutputDir, 0o755); err != nil {
		return "", apperr.New(apperr.CodeInvalid, "无法创建输出目录。", err)
	}
	nameHint := filename.DefaultTarName(parsed, platform)
	outPath, err := filename.ResolveOutputPath(req.OutputDir, nameHint)
	if err != nil {
		return "", apperr.New(apperr.CodeInvalid, "输出路径无效。", err)
	}

	task := &model.DownloadTask{
		ID:         uuid.NewString(),
		Image:      parsed.String(),
		OutputPath: outPath,
		Platform:   platform,
		Status:     model.StatusPending,
		StartedAt:  time.Now(),
	}
	ctx, err := s.store.Begin(task, s.parent)
	if err != nil {
		return "", err
	}

	s.log.Info("Download started " + task.Image + " platform " + platform.String())
	go func() {
		defer s.store.Finish(task.ID)
		s.downloadImage(ctx, req, task, parsed, progressMeta{Image: task.Image})
	}()
	return task.ID, nil
}

func (s *Service) downloadImage(ctx context.Context, req model.DownloadRequest, task *model.DownloadTask, parsed model.ImageReference, meta progressMeta) {
	tmp := task.OutputPath + ".downloading"
	defer func() {
		if task.Status != model.StatusCompleted {
			export.RemoveIfExists(tmp)
		}
	}()

	set := func(status, message string) {
		task.Status = status
		if message != "" && status == model.StatusFailed {
			task.Error = message
		}
		if status == model.StatusCompleted || status == model.StatusFailed || status == model.StatusCancelled {
			task.FinishedAt = time.Now()
		}
		s.store.Update(task)
		if s.emit != nil {
			s.emit.Status(*task)
		}
	}

	fail := func(err error) {
		if ctx.Err() != nil || errors.Is(err, context.Canceled) {
			s.log.Info("Download cancelled " + task.ID)
			task.Error = "任务已取消。"
			set(model.StatusCancelled, task.Error)
			if s.emit != nil {
				s.emit.Error(model.AppErrorEvent{TaskID: task.ID, Message: task.Error})
			}
			return
		}
		classified := apperr.Classify(err)
		s.log.Error("Download failed " + classified.Code)
		task.Error = classified.Message
		set(model.StatusFailed, classified.Message)
		if s.emit != nil {
			s.emit.Error(model.AppErrorEvent{TaskID: task.ID, Message: classified.Message})
		}
	}

	set(model.StatusChecking, "")
	if s.emit != nil {
		s.emit.Progress(withMeta(model.ProgressEvent{
			TaskID:  task.ID,
			Phase:   model.StatusChecking,
			Message: "正在解析镜像并连接 Registry",
			Image:   task.Image,
		}, meta))
	}

	client := registry.NewClient(model.RegistryConfig{
		Registry: req.Registry,
		Username: req.Username,
		Password: req.Password,
		Insecure: req.Insecure,
	})

	img, parsed, err := client.Image(ctx, req.Image, req.Registry, task.Platform)
	if err != nil {
		fail(err)
		return
	}
	task.Image = parsed.String()
	if parsed.Digest != "" {
		task.Digest = parsed.Digest
		s.log.Info("Manifest " + parsed.Digest)
	}

	est, layerCount, layerSizes, err := registry.EstimateSize(img)
	if err != nil {
		fail(err)
		return
	}
	s.log.Info("Layer count " + strconv.Itoa(layerCount))

	var cacheStats layercache.Stats
	lc, cacheErr := layercache.Open()
	if cacheErr != nil {
		s.log.Warn("Layer cache unavailable, downloading without cache")
		cacheStats.TotalLayers = layerCount
		cacheStats.MissingLayers = layerCount
		cacheStats.MissingBytes = est
		cacheStats.TotalBytes = est
	} else {
		cacheStats, err = layercache.Inspect(img, lc)
		if err != nil {
			s.log.Warn("Layer cache inspect failed, treating all layers as missing")
			cacheStats.TotalLayers = layerCount
			cacheStats.MissingLayers = layerCount
			cacheStats.MissingBytes = est
			cacheStats.TotalBytes = est
		} else {
			s.log.Info("Layer cache hits " + strconv.Itoa(cacheStats.CachedLayers) + "/" + strconv.Itoa(cacheStats.TotalLayers))
		}
		img = lc.Wrap(img)
	}

	cachePath := ""
	if lc != nil {
		cachePath = lc.Path()
	}
	if err := disk.EnsureDownloadSpace(filepath.Dir(task.OutputPath), cachePath, est, cacheStats.MissingBytes); err != nil {
		fail(apperr.New(apperr.CodeDiskFull, err.Error(), err))
		return
	}

	agg := NewAggregator(task.ID, layerSizes)
	agg.SetCache(cacheStats.CachedLayers, cacheStats.MissingLayers)
	agg.SetMeta(meta)
	agg.SetPhase(model.StatusDownloading, cacheStats.Message())
	set(model.StatusDownloading, "")

	nopts := []name.Option{name.WeakValidation}
	if req.Insecure {
		nopts = append(nopts, name.Insecure)
	}
	refName := parsed.Raw
	if refName == "" {
		refName = task.Image
	}
	ref, err := name.ParseReference(refName, nopts...)
	if err != nil {
		fail(err)
		return
	}
	writeRef := ref
	if target := strings.TrimSpace(req.TargetTag); target != "" {
		tagged, err := name.NewTag(target, nopts...)
		if err != nil {
			fail(apperr.New(apperr.CodeInvalid, "无法解析保存为的镜像名。\n请使用例如 app-gateway:1.0.0 或 company/app:prod 的格式。", err))
			return
		}
		writeRef = tagged
		if s.log != nil {
			s.log.Info("Export as " + tagged.String())
		}
	}

	progress := make(chan v1.Update, 256)
	done := make(chan struct{})
	go func() {
		defer close(done)
		var lastEmit time.Time
		for u := range progress {
			if u.Error != nil {
				continue
			}
			ev := agg.EventFromUpdate(u)
			if s.emit != nil && (lastEmit.IsZero() || time.Since(lastEmit) >= 200*time.Millisecond || u.Complete == u.Total) {
				s.emit.Progress(ev)
				lastEmit = time.Now()
			}
		}
	}()

	s.log.Info("Download writing tarball")
	set(model.StatusExporting, "")
	exportMsg := "正在写入 TAR"
	if cacheStats.CachedLayers > 0 {
		exportMsg = cacheStats.Message()
	}
	agg.SetPhase(model.StatusExporting, exportMsg)
	err = export.Write(ctx, tmp, writeRef, img, progress)
	close(progress)
	<-done
	if err != nil {
		fail(err)
		return
	}
	if ctx.Err() != nil {
		fail(ctx.Err())
		return
	}

	st, err := os.Stat(tmp)
	if err != nil || st.Size() <= 0 {
		fail(apperr.New(apperr.CodeUnknown, "导出校验失败：文件不存在或大小为 0。", err))
		return
	}
	if err := os.Rename(tmp, task.OutputPath); err != nil {
		fail(err)
		return
	}
	task.Size = st.Size()
	s.log.Info("Tarball writing completed")
	if lc != nil {
		if err := lc.Remember(task.Image, task.Digest, task.Platform.String(), img); err != nil {
			s.log.Warn("Layer cache index update failed")
		}
	}
	set(model.StatusCompleted, "")
	if s.emit != nil {
		s.emit.Progress(agg.Snapshot(st.Size(), st.Size()))
		s.emit.Completed(model.DownloadCompleteEvent{
			TaskID:     task.ID,
			Image:      task.Image,
			Platform:   task.Platform.String(),
			Digest:     task.Digest,
			OutputPath: task.OutputPath,
			Size:       task.Size,
		})
	}
}
