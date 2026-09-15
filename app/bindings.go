package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/google/uuid"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"HarborDownloader/internal/apperr"
	"HarborDownloader/internal/config"
	"HarborDownloader/internal/disk"
	"HarborDownloader/internal/filename"
	"HarborDownloader/internal/layercache"
	"HarborDownloader/internal/model"
	"HarborDownloader/internal/registry"
)

func (a *App) GetVersion() string {
	v := a.version
	if v == "" {
		v = "dev"
	}
	return v
}

func (a *App) GetBuildInfo() map[string]string {
	return map[string]string{
		"version": a.version,
		"commit":  a.commit,
		"buildAt": a.buildAt,
	}
}

func (a *App) LoadConfig() model.FileConfig {
	if a.cfg == nil {
		return model.FileConfig{OutputDir: config.DefaultOutputDir()}
	}
	cfg := a.cfg.Load()
	a.rememberCred(cfg.Registry, cfg.Username, cfg.Password)
	return redactForUI(cfg)
}

func (a *App) SaveConfig(cfg model.FileConfig) model.Result {
	if err := a.persistConfig(func(cur *model.FileConfig) {
		cur.Registry = cfg.Registry
		cur.Username = cfg.Username
		cur.Password = keepPassword(cfg.Password, cur.Password)
		cur.Insecure = cfg.Insecure
		if cfg.OutputDir != "" {
			cur.OutputDir = cfg.OutputDir
		}
	}); err != nil {
		return failResult(err)
	}
	return model.Result{OK: true, Message: "已保存"}
}

func (a *App) ExportConfig() model.Result {
	if a.ctx == nil {
		return failResult(fmt.Errorf("应用尚未启动"))
	}
	if a.cfg == nil {
		return failResult(fmt.Errorf("配置服务不可用"))
	}
	path, err := wailsruntime.SaveFileDialog(a.ctx, wailsruntime.SaveDialogOptions{
		Title:           "导出配置",
		DefaultFilename: "HarborDownloader-config.json",
		Filters: []wailsruntime.FileFilter{
			{DisplayName: "JSON (*.json)", Pattern: "*.json"},
		},
	})
	if err != nil {
		return failResult(err)
	}
	if path == "" {
		return model.Result{OK: true}
	}
	if filepath.Ext(path) == "" {
		path += ".json"
	}
	b, err := a.cfg.ExportBytes()
	if err != nil {
		return failResult(err)
	}
	if err := os.WriteFile(path, b, 0o600); err != nil {
		return failResult(err)
	}
	if a.log != nil {
		a.log.Info("Config exported")
	}
	return model.Result{OK: true, Message: "已导出到 " + path}
}

func (a *App) ImportConfig() model.Result {
	if a.ctx == nil {
		return failResult(fmt.Errorf("应用尚未启动"))
	}
	if a.cfg == nil {
		return failResult(fmt.Errorf("配置服务不可用"))
	}
	if a.tasks != nil && a.tasks.Busy() {
		return failResult(fmt.Errorf("下载进行中，无法导入配置。"))
	}
	path, err := wailsruntime.OpenFileDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title: "导入配置",
		Filters: []wailsruntime.FileFilter{
			{DisplayName: "JSON (*.json)", Pattern: "*.json"},
		},
	})
	if err != nil {
		return failResult(err)
	}
	if path == "" {
		return model.Result{OK: true}
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return failResult(err)
	}
	cfg, err := config.Decode(b)
	if err != nil {
		return failResult(err)
	}
	for i := range cfg.Jobs {
		if cfg.Jobs[i].ID == "" {
			cfg.Jobs[i].ID = uuid.NewString()
		}
	}
	if err := a.cfg.Save(cfg); err != nil {
		return failResult(err)
	}
	a.rememberCred(cfg.Registry, cfg.Username, cfg.Password)
	if a.log != nil {
		a.log.Info("Config imported")
	}
	msg := fmt.Sprintf("已导入仓库设置和 %d 个任务", len(cfg.Jobs))
	return model.Result{OK: true, Message: msg}
}

func (a *App) GetDefaultOutputDir() string {
	if a.cfg != nil {
		c := a.cfg.Load()
		if c.OutputDir != "" {
			return c.OutputDir
		}
	}
	return config.DefaultOutputDir()
}

func (a *App) SelectOutputDir() (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("应用尚未启动")
	}
	dir, err := wailsruntime.OpenDirectoryDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title:            "选择输出目录",
		DefaultDirectory: a.GetDefaultOutputDir(),
	})
	if err != nil {
		return "", err
	}
	return dir, nil
}

func (a *App) OpenOutputDir(path string) model.Result {
	if path == "" {
		return failResult(fmt.Errorf("路径为空"))
	}
	target := path
	if st, err := os.Stat(path); err == nil && !st.IsDir() {
		target = filepath.Dir(path)
	} else if err != nil {
		target = filepath.Dir(path)
	}
	abs, err := filepath.Abs(target)
	if err != nil {
		return failResult(err)
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", abs)
	case "darwin":
		cmd = exec.Command("open", abs)
	default:
		cmd = exec.Command("xdg-open", abs)
	}
	if err := cmd.Start(); err != nil {
		return failResult(err)
	}
	return model.Result{OK: true, Message: abs}
}

func (a *App) ParseImage(image, registryHost string, insecure bool) (model.ImageReference, error) {
	parsed, _, err := registry.ParseImageReference(image, registryHost, insecure)
	if err != nil {
		return model.ImageReference{}, err
	}
	return parsed, nil
}

func (a *App) ListPlatforms() []model.Platform {
	return model.SupportedPlatforms()
}

func (a *App) TestRegistry(cfg model.RegistryConfig) model.TestRegistryResult {
	cfg.Username, cfg.Password = a.resolveCred(cfg.Registry, cfg.Username, cfg.Password)
	a.emitLog("Registry connect " + cfg.Registry)
	client := registry.NewClient(cfg)
	result := client.Test(a.ctx, cfg)
	if result.OK {
		a.emitLog("Authentication success")
		_ = a.persistConfig(func(cur *model.FileConfig) {
			cur.Registry = cfg.Registry
			cur.Username = cfg.Username
			cur.Password = keepPassword(cfg.Password, cur.Password)
			cur.Insecure = cfg.Insecure
		})
	} else {
		a.emitLog("Registry connect failed")
	}
	return result
}

func (a *App) StartDownload(req model.DownloadRequest) (string, error) {
	req.Username, req.Password = a.resolveCred(req.Registry, req.Username, req.Password)
	if a.cred != nil && (req.Username != "" || req.Password != "") {
		a.cred.Set(req.Registry, req.Username, req.Password)
	}
	if a.tasks == nil {
		return "", fmt.Errorf("下载服务不可用")
	}
	id, err := a.tasks.Start(req)
	if err != nil {
		return "", classifiedError(err)
	}
	_ = a.persistConfig(func(cur *model.FileConfig) {
		cur.Registry = req.Registry
		cur.Username = req.Username
		cur.Password = keepPassword(req.Password, cur.Password)
		cur.Insecure = req.Insecure
		if req.OutputDir != "" {
			cur.OutputDir = req.OutputDir
		}
	})
	return id, nil
}

func (a *App) CancelDownload(taskID string) model.Result {
	if a.tasks == nil {
		return failResult(fmt.Errorf("下载服务不可用"))
	}
	if err := a.tasks.Cancel(taskID); err != nil {
		return failResult(err)
	}
	return model.Result{OK: true, Message: "正在取消"}
}

func (a *App) GetDownload(taskID string) (model.DownloadTask, error) {
	if a.tasks == nil {
		return model.DownloadTask{}, fmt.Errorf("下载服务不可用")
	}
	task, ok := a.tasks.Get(taskID)
	if !ok {
		return model.DownloadTask{}, fmt.Errorf("任务不存在")
	}
	return task, nil
}

func (a *App) ListJobs() []model.Job {
	if a.cfg == nil {
		return []model.Job{}
	}
	jobs := a.cfg.Load().Jobs
	if jobs == nil {
		return []model.Job{}
	}
	return jobs
}

func (a *App) SaveJob(job model.Job) (model.Job, error) {
	job = model.NormalizeJob(job)
	if err := model.ValidateJob(job); err != nil {
		return model.Job{}, classifiedError(err)
	}
	if job.ID == "" {
		job.ID = uuid.NewString()
	}
	if err := a.persistConfig(func(cur *model.FileConfig) {
		replaced := false
		for i := range cur.Jobs {
			if cur.Jobs[i].ID == job.ID {
				cur.Jobs[i] = job
				replaced = true
				break
			}
		}
		if !replaced {
			cur.Jobs = append(cur.Jobs, job)
		}
	}); err != nil {
		return model.Job{}, err
	}
	return job, nil
}

func (a *App) DeleteJob(id string) model.Result {
	if id == "" {
		return failResult(fmt.Errorf("任务不存在"))
	}
	found := false
	if err := a.persistConfig(func(cur *model.FileConfig) {
		next := make([]model.Job, 0, len(cur.Jobs))
		for _, j := range cur.Jobs {
			if j.ID == id {
				found = true
				continue
			}
			next = append(next, j)
		}
		cur.Jobs = next
	}); err != nil {
		return failResult(err)
	}
	if !found {
		return failResult(fmt.Errorf("任务不存在"))
	}
	return model.Result{OK: true, Message: "已删除"}
}

func (a *App) StartJob(id string) (string, error) {
	if a.tasks == nil {
		return "", fmt.Errorf("下载服务不可用")
	}
	if a.cfg == nil {
		return "", fmt.Errorf("配置服务不可用")
	}
	cfg := a.cfg.Load()
	var job *model.Job
	for i := range cfg.Jobs {
		if cfg.Jobs[i].ID == id {
			job = &cfg.Jobs[i]
			break
		}
	}
	if job == nil {
		return "", fmt.Errorf("任务不存在")
	}
	copied := *job
	root := cfg.OutputDir
	if root == "" {
		root = config.DefaultOutputDir()
	}
	jobDir, err := filename.JobOutputDir(root, copied.Name)
	if err != nil {
		return "", classifiedError(apperr.New(apperr.CodeInvalid, "无法解析任务下载目录。", err))
	}
	copied.OutputDir = jobDir
	cred := model.RegistryConfig{
		Registry: cfg.Registry,
		Username: cfg.Username,
		Password: cfg.Password,
		Insecure: cfg.Insecure,
	}
	cred.Username, cred.Password = a.resolveCred(cred.Registry, cred.Username, cred.Password)
	if a.cred != nil && (cred.Username != "" || cred.Password != "") {
		a.cred.Set(cred.Registry, cred.Username, cred.Password)
	}
	runID, err := a.tasks.StartJob(copied, cred)
	if err != nil {
		return "", classifiedError(err)
	}
	return runID, nil
}

func (a *App) GetLayerCache() model.LayerCacheInfo {
	inv := a.GetLayerCacheInventory()
	return model.LayerCacheInfo{
		Path:              inv.Path,
		Files:             inv.Files,
		Bytes:             inv.Bytes,
		Images:            inv.Images,
		UnreferencedFiles: inv.UnreferencedFiles,
		UnreferencedBytes: inv.UnreferencedBytes,
	}
}

func (a *App) GetLayerCacheInventory() model.LayerCacheInventory {
	dir, err := layercache.Dir()
	if err != nil {
		return emptyInventory()
	}
	return inventoryToModel(layercache.New(dir).Inventory())
}

func (a *App) DeleteCachedLayer(digest string) model.Result {
	if a.tasks != nil && a.tasks.Busy() {
		return failResult(fmt.Errorf("下载进行中，无法删除缓存层。"))
	}
	c, err := layercache.Open()
	if err != nil {
		return failResult(err)
	}
	if err := c.DeleteLayer(digest); err != nil {
		return failResult(err)
	}
	if a.log != nil {
		a.log.Info("Layer cache deleted " + digest)
	}
	return model.Result{OK: true, Message: "已删除该层。"}
}

func (a *App) PruneLayerCache() model.Result {
	if a.tasks != nil && a.tasks.Busy() {
		return failResult(fmt.Errorf("下载进行中，无法清理缓存。"))
	}
	c, err := layercache.Open()
	if err != nil {
		return failResult(err)
	}
	removed, err := c.ClearUnreferenced()
	if err != nil {
		return failResult(err)
	}
	if a.log != nil {
		a.log.Info("Layer cache pruned unreferenced layers")
	}
	msg := "没有可清理的未引用层。"
	if removed.Files > 0 {
		msg = fmt.Sprintf("已删除 %d 个未引用层，释放 %s。", removed.Files, disk.FormatBytes(removed.Bytes))
	}
	return model.Result{OK: true, Message: msg}
}

func (a *App) ClearLayerCache() model.Result {
	if a.tasks != nil && a.tasks.Busy() {
		return failResult(fmt.Errorf("下载进行中，无法清理缓存。"))
	}
	c, err := layercache.Open()
	if err != nil {
		return failResult(err)
	}
	removed, err := c.Clear()
	if err != nil {
		return failResult(err)
	}
	if a.log != nil {
		a.log.Info("Layer cache cleared")
	}
	msg := "缓存已清空。"
	if removed.Files > 0 {
		msg = fmt.Sprintf("已删除 %d 个文件，释放 %s。", removed.Files, disk.FormatBytes(removed.Bytes))
	}
	return model.Result{OK: true, Message: msg}
}

func (a *App) OpenLayerCacheDir() model.Result {
	dir, err := layercache.Dir()
	if err != nil {
		return failResult(err)
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return failResult(err)
	}
	return a.OpenOutputDir(dir)
}

func emptyInventory() model.LayerCacheInventory {
	return model.LayerCacheInventory{
		Layers:       []model.CachedLayer{},
		ImageRecords: []model.CachedImage{},
	}
}

func inventoryToModel(inv layercache.Inventory) model.LayerCacheInventory {
	out := model.LayerCacheInventory{
		Path:              inv.Usage.Path,
		Files:             inv.Usage.Files,
		Bytes:             inv.Usage.Bytes,
		Images:            inv.Usage.Images,
		UnreferencedFiles: inv.Usage.UnreferencedFiles,
		UnreferencedBytes: inv.Usage.UnreferencedBytes,
		Layers:            make([]model.CachedLayer, 0, len(inv.Layers)),
		ImageRecords:      make([]model.CachedImage, 0, len(inv.Images)),
	}
	for _, l := range inv.Layers {
		cl := model.CachedLayer{
			Digest:     l.Digest,
			Size:       l.Size,
			Partial:    l.Partial,
			Referenced: l.Referenced,
			Images:     make([]model.CachedLayerImage, 0, len(l.Images)),
		}
		for _, img := range l.Images {
			cl.Images = append(cl.Images, model.CachedLayerImage{
				Image:    img.Image,
				Platform: img.Platform,
				Digest:   img.Digest,
			})
		}
		out.Layers = append(out.Layers, cl)
	}
	for _, img := range inv.Images {
		updated := ""
		if !img.UpdatedAt.IsZero() {
			updated = img.UpdatedAt.UTC().Format("2006-01-02 15:04:05")
		}
		out.ImageRecords = append(out.ImageRecords, model.CachedImage{
			Image:       img.Image,
			Digest:      img.Digest,
			Platform:    img.Platform,
			UpdatedAt:   updated,
			LayerCount:  img.LayerCount,
			CachedCount: img.CachedCount,
		})
	}
	return out
}

func (a *App) persistConfig(update func(*model.FileConfig)) error {
	if a.cfg == nil {
		return fmt.Errorf("配置服务不可用")
	}
	if err := a.cfg.Update(update); err != nil {
		return err
	}
	cfg := a.cfg.Load()
	a.rememberCred(cfg.Registry, cfg.Username, cfg.Password)
	return nil
}

func (a *App) rememberCred(registry, username, password string) {
	if a.cred == nil || registry == "" {
		return
	}
	a.cred.Set(registry, username, password)
}

func redactForUI(cfg model.FileConfig) model.FileConfig {
	cfg.HasPassword = cfg.Password != ""
	cfg.Password = ""
	return cfg
}

func keepPassword(next, prev string) string {
	if next != "" {
		return next
	}
	return prev
}

func (a *App) resolveCred(registry, username, password string) (string, string) {
	if password != "" {
		return username, password
	}
	if a.cred != nil {
		if u, p, ok := a.cred.Get(registry); ok {
			if username == "" {
				username = u
			}
			if p != "" {
				return username, p
			}
		}
	}
	if a.cfg != nil {
		saved := a.cfg.Load()
		if username == "" {
			username = saved.Username
		}
		if saved.Password != "" {
			password = saved.Password
		}
	}
	return username, password
}

func failResult(err error) model.Result {
	if err == nil {
		return model.Result{OK: true}
	}
	return model.Result{OK: false, Error: classifiedError(err).Error(), Message: classifiedError(err).Error()}
}

func classifiedError(err error) error {
	if err == nil {
		return nil
	}
	return apperr.Classify(err)
}
