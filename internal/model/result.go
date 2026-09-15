package model

type Result struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

type TestRegistryResult struct {
	OK                bool   `json:"ok"`
	RegistryReachable bool   `json:"registryReachable"`
	AuthSuccess       bool   `json:"authSuccess"`
	HarborReachable   bool   `json:"harborReachable"`
	Message           string `json:"message"`
	Error             string `json:"error,omitempty"`
}

type DownloadCompleteEvent struct {
	TaskID     string `json:"taskId"`
	Image      string `json:"image"`
	Platform   string `json:"platform"`
	Digest     string `json:"digest"`
	OutputPath string `json:"outputPath"`
	Size       int64  `json:"size"`
}

type AppErrorEvent struct {
	TaskID  string `json:"taskId"`
	Message string `json:"message"`
}

type FileConfig struct {
	Registry    string `json:"registry"`
	Username    string `json:"username,omitempty"`
	Password    string `json:"password,omitempty"`
	HasPassword bool   `json:"hasPassword,omitempty"`
	Insecure    bool   `json:"insecure,omitempty"`
	OutputDir   string `json:"outputDir"`
	Jobs        []Job  `json:"jobs,omitempty"`
}

type LayerCacheInfo struct {
	Path              string `json:"path"`
	Files             int    `json:"files"`
	Bytes             int64  `json:"bytes"`
	Images            int    `json:"images"`
	UnreferencedFiles int    `json:"unreferencedFiles"`
	UnreferencedBytes int64  `json:"unreferencedBytes"`
}

type CachedLayerImage struct {
	Image    string `json:"image"`
	Platform string `json:"platform"`
	Digest   string `json:"digest,omitempty"`
}

type CachedLayer struct {
	Digest     string             `json:"digest"`
	Size       int64              `json:"size"`
	Partial    bool               `json:"partial"`
	Referenced bool               `json:"referenced"`
	Images     []CachedLayerImage `json:"images"`
}

type CachedImage struct {
	Image       string `json:"image"`
	Digest      string `json:"digest,omitempty"`
	Platform    string `json:"platform"`
	UpdatedAt   string `json:"updatedAt,omitempty"`
	LayerCount  int    `json:"layerCount"`
	CachedCount int    `json:"cachedCount"`
}

type LayerCacheInventory struct {
	Path              string        `json:"path"`
	Files             int           `json:"files"`
	Bytes             int64         `json:"bytes"`
	Images            int           `json:"images"`
	UnreferencedFiles int           `json:"unreferencedFiles"`
	UnreferencedBytes int64         `json:"unreferencedBytes"`
	Layers            []CachedLayer `json:"layers"`
	ImageRecords      []CachedImage `json:"imageRecords"`
}
