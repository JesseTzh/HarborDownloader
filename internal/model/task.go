package model

import "time"

const (
	StatusPending     = "pending"
	StatusChecking    = "checking"
	StatusDownloading = "downloading"
	StatusExporting   = "exporting"
	StatusPackaging   = "packaging"
	StatusCompleted   = "completed"
	StatusCancelled   = "cancelled"
	StatusFailed      = "failed"
)

type DownloadTask struct {
	ID         string    `json:"id"`
	Image      string    `json:"image"`
	OutputPath string    `json:"outputPath"`
	Platform   Platform  `json:"platform"`
	Status     string    `json:"status"`
	StartedAt  time.Time `json:"startedAt"`
	FinishedAt time.Time `json:"finishedAt,omitempty"`
	Error      string    `json:"error,omitempty"`
	Digest     string    `json:"digest,omitempty"`
	Size       int64     `json:"size,omitempty"`
}

type DownloadRequest struct {
	Image     string   `json:"image"`
	TargetTag string   `json:"targetTag,omitempty"`
	OutputDir string   `json:"outputDir"`
	Platform  Platform `json:"platform"`
	Registry  string   `json:"registry"`
	Username  string   `json:"username"`
	Password  string   `json:"password"`
	Insecure  bool     `json:"insecure"`
}
