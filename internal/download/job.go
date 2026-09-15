package download

import (
	"context"
	"strconv"
	"time"

	"github.com/google/uuid"

	"HarborDownloader/internal/apperr"
	"HarborDownloader/internal/export"
	"HarborDownloader/internal/filename"
	"HarborDownloader/internal/model"
	"HarborDownloader/internal/registry"
)

func withMeta(ev model.ProgressEvent, meta progressMeta) model.ProgressEvent {
	if meta.RunID != "" {
		ev.RunID = meta.RunID
	}
	if meta.JobName != "" {
		ev.JobName = meta.JobName
	}
	if meta.Image != "" {
		ev.Image = meta.Image
	}
	if meta.Index > 0 {
		ev.ImageIndex = meta.Index
	}
	if meta.Total > 0 {
		ev.ImageTotal = meta.Total
	}
	return ev
}

func (s *Service) StartJob(job model.Job, cred model.RegistryConfig) (string, error) {
	job = model.NormalizeJob(job)
	if err := model.ValidateJob(job); err != nil {
		return "", apperr.New(apperr.CodeInvalid, err.Error(), err)
	}
	if s.store.ActiveID() != "" {
		return "", apperr.New(apperr.CodeInvalid, "已有下载任务正在进行，同时只能运行一个任务。", nil)
	}
	if job.OutputDir == "" {
		return "", apperr.New(apperr.CodeInvalid, "请在设置中配置默认下载根目录。", nil)
	}
	if err := filename.PrepareJobOutputDir(job.OutputDir); err != nil {
		return "", apperr.New(apperr.CodeInvalid, "无法准备任务下载目录。", err)
	}

	runID := uuid.NewString()
	items := make([]model.JobItem, len(job.Images))
	for i, img := range job.Images {
		items[i] = model.JobItem{
			Image:     img.Image,
			TargetTag: img.TargetTag,
			Platform:  img.Platform.String(),
			Status:    model.StatusPending,
		}
	}
	run := &model.JobRun{
		ID:        runID,
		JobID:     job.ID,
		Name:      job.Name,
		Status:    model.StatusPending,
		Total:     len(job.Images),
		Items:     items,
		StartedAt: time.Now(),
	}
	ctx, err := s.store.BeginJob(runID, s.parent)
	if err != nil {
		return "", err
	}
	if s.log != nil {
		s.log.Info("Job started " + job.Name + " images " + strconv.Itoa(len(job.Images)))
	}
	go s.runJob(ctx, job, cred, run)
	return runID, nil
}

func (s *Service) runJob(ctx context.Context, job model.Job, cred model.RegistryConfig, run *model.JobRun) {
	defer s.store.Finish(run.ID)

	setJob := func(status string) {
		run.Status = status
		if status == model.StatusCompleted || status == model.StatusFailed || status == model.StatusCancelled {
			run.FinishedAt = time.Now()
		}
		s.emitJobStatus(run)
	}

	setJob(model.StatusDownloading)

	cancelled := false
	for i, img := range job.Images {
		if ctx.Err() != nil {
			cancelled = true
			s.markRemaining(run, i, model.StatusCancelled, "任务已取消。")
			break
		}
		run.Current = i + 1
		run.Items[i].Status = model.StatusChecking
		s.emitJobStatus(run)

		platform, err := registry.ParsePlatform(img.Platform)
		if err != nil {
			s.failItem(run, i, err)
			continue
		}
		parsed, _, err := registry.ParseImageReference(img.Image, cred.Registry, cred.Insecure)
		if err != nil {
			s.failItem(run, i, err)
			continue
		}
		nameHint, err := filename.DefaultTarNameFor(parsed, platform, img.TargetTag)
		if err != nil {
			s.failItem(run, i, err)
			continue
		}
		outPath, err := filename.ResolveOutputPath(job.OutputDir, nameHint)
		if err != nil {
			s.failItem(run, i, apperr.New(apperr.CodeInvalid, "输出路径无效。", err))
			continue
		}

		task := &model.DownloadTask{
			ID:         uuid.NewString(),
			Image:      parsed.String(),
			OutputPath: outPath,
			Platform:   platform,
			Status:     model.StatusPending,
			StartedAt:  time.Now(),
		}
		s.store.SetTask(task)
		run.Items[i].Image = task.Image
		run.Items[i].Platform = platform.String()

		if s.log != nil {
			s.log.Info("Job image " + strconv.Itoa(i+1) + "/" + strconv.Itoa(len(job.Images)) + " " + task.Image)
		}

		req := model.DownloadRequest{
			Image:     img.Image,
			TargetTag: img.TargetTag,
			OutputDir: job.OutputDir,
			Platform:  platform,
			Registry:  cred.Registry,
			Username:  cred.Username,
			Password:  cred.Password,
			Insecure:  cred.Insecure,
		}
		meta := progressMeta{
			RunID:   run.ID,
			JobName: run.Name,
			Image:   task.Image,
			Index:   i + 1,
			Total:   len(job.Images),
		}
		s.downloadImage(ctx, req, task, parsed, meta)

		run.Items[i].Status = task.Status
		run.Items[i].OutputPath = task.OutputPath
		run.Items[i].Digest = task.Digest
		run.Items[i].Size = task.Size
		run.Items[i].Error = task.Error
		if task.Status == model.StatusCancelled {
			cancelled = true
			s.markRemaining(run, i+1, model.StatusCancelled, "任务已取消。")
			break
		}
		s.emitJobStatus(run)
	}

	if cancelled || ctx.Err() != nil {
		run.Error = "任务已取消。"
		setJob(model.StatusCancelled)
	} else if jobHasFailure(run) && !jobHasSuccess(run) {
		run.Error = firstItemError(run)
		setJob(model.StatusFailed)
	} else if jobHasFailure(run) {
		run.Error = "部分镜像下载失败。"
		setJob(model.StatusFailed)
	} else if job.Pack {
		s.packJob(ctx, job, run, setJob)
	} else {
		setJob(model.StatusCompleted)
	}

	if s.emit != nil {
		s.emit.JobCompleted(model.JobCompleteEvent{
			RunID:       run.ID,
			JobID:       run.JobID,
			Name:        run.Name,
			Status:      run.Status,
			Items:       copyItems(run.Items),
			PackagePath: run.PackagePath,
			Error:       run.Error,
		})
	}
}

func (s *Service) packJob(ctx context.Context, job model.Job, run *model.JobRun, setJob func(string)) {
	if ctx.Err() != nil {
		run.Error = "任务已取消。"
		setJob(model.StatusCancelled)
		return
	}
	setJob(model.StatusPackaging)
	if s.log != nil {
		s.log.Info("Job packing " + job.Name)
	}
	if s.emit != nil {
		s.emit.Progress(model.ProgressEvent{
			TaskID:     run.ID,
			Phase:      model.StatusPackaging,
			Message:    "正在合并 TAR 并压缩为 all.tar.zst",
			RunID:      run.ID,
			JobName:    run.Name,
			Image:      export.PackedName,
			ImageIndex: run.Total,
			ImageTotal: run.Total,
		})
	}
	var lastEmit time.Time
	path, size, err := export.PackTars(ctx, job.OutputDir, func(current, total int64) {
		if s.emit == nil {
			return
		}
		now := time.Now()
		if !lastEmit.IsZero() && now.Sub(lastEmit) < 200*time.Millisecond && current < total {
			return
		}
		lastEmit = now
		pct := 0.0
		if total > 0 {
			pct = float64(current) / float64(total) * 100
			if pct > 100 {
				pct = 100
			}
		}
		s.emit.Progress(model.ProgressEvent{
			TaskID:          run.ID,
			Phase:           model.StatusPackaging,
			Current:         current,
			Total:           total,
			Percentage:      pct,
			DownloadedBytes: current,
			Message:         "正在合并 TAR 并压缩为 all.tar.zst",
			RunID:           run.ID,
			JobName:         run.Name,
			Image:           export.PackedName,
			ImageIndex:      run.Total,
			ImageTotal:      run.Total,
		})
	})
	if err != nil {
		if ctx.Err() != nil {
			run.Error = "任务已取消。"
			setJob(model.StatusCancelled)
			return
		}
		classified := apperr.Classify(err)
		run.Error = classified.Message
		if s.log != nil {
			s.log.Error("Job packing failed " + classified.Code)
		}
		if s.emit != nil {
			s.emit.Error(model.AppErrorEvent{TaskID: run.ID, Message: classified.Message})
		}
		setJob(model.StatusFailed)
		return
	}
	run.PackagePath = path
	if s.log != nil {
		s.log.Info("Job packed " + path + " size " + strconv.FormatInt(size, 10))
	}
	if s.emit != nil {
		s.emit.Progress(model.ProgressEvent{
			TaskID:          run.ID,
			Phase:           model.StatusPackaging,
			Current:         size,
			Total:           size,
			Percentage:      100,
			DownloadedBytes: size,
			Message:         "打包完成 " + export.PackedName,
			RunID:           run.ID,
			JobName:         run.Name,
			Image:           export.PackedName,
			ImageIndex:      run.Total,
			ImageTotal:      run.Total,
		})
	}
	setJob(model.StatusCompleted)
}

func (s *Service) failItem(run *model.JobRun, i int, err error) {
	classified := apperr.Classify(err)
	run.Items[i].Status = model.StatusFailed
	run.Items[i].Error = classified.Message
	if s.emit != nil {
		s.emit.Error(model.AppErrorEvent{TaskID: run.ID, Message: classified.Message})
	}
	s.emitJobStatus(run)
}

func (s *Service) markRemaining(run *model.JobRun, from int, status, message string) {
	for j := from; j < len(run.Items); j++ {
		if run.Items[j].Status == model.StatusPending || run.Items[j].Status == model.StatusChecking {
			run.Items[j].Status = status
			run.Items[j].Error = message
		}
	}
}

func (s *Service) emitJobStatus(run *model.JobRun) {
	if s.emit == nil {
		return
	}
	s.emit.JobStatus(model.JobStatusEvent{
		RunID:       run.ID,
		JobID:       run.JobID,
		Name:        run.Name,
		Status:      run.Status,
		Current:     run.Current,
		Total:       run.Total,
		Items:       copyItems(run.Items),
		PackagePath: run.PackagePath,
		Error:       run.Error,
	})
}

func copyItems(items []model.JobItem) []model.JobItem {
	out := make([]model.JobItem, len(items))
	copy(out, items)
	return out
}

func jobHasFailure(run *model.JobRun) bool {
	for _, it := range run.Items {
		if it.Status == model.StatusFailed {
			return true
		}
	}
	return false
}

func jobHasSuccess(run *model.JobRun) bool {
	for _, it := range run.Items {
		if it.Status == model.StatusCompleted {
			return true
		}
	}
	return false
}

func firstItemError(run *model.JobRun) string {
	for _, it := range run.Items {
		if it.Error != "" {
			return it.Error
		}
	}
	return "下载失败。"
}
