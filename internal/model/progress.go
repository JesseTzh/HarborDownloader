package model

type ProgressEvent struct {
	TaskID          string  `json:"taskId"`
	Phase           string  `json:"phase"`
	Current         int64   `json:"current"`
	Total           int64   `json:"total"`
	Percentage      float64 `json:"percentage"`
	SpeedBytes      int64   `json:"speedBytes"`
	DownloadedBytes int64   `json:"downloadedBytes"`
	ETASeconds      int64   `json:"etaSeconds"`
	CurrentLayer    int     `json:"currentLayer"`
	TotalLayers     int     `json:"totalLayers"`
	CachedLayers    int     `json:"cachedLayers"`
	MissingLayers   int     `json:"missingLayers"`
	Message         string  `json:"message"`
	RunID           string  `json:"runId,omitempty"`
	JobName         string  `json:"jobName,omitempty"`
	Image           string  `json:"image,omitempty"`
	ImageIndex      int     `json:"imageIndex,omitempty"`
	ImageTotal      int     `json:"imageTotal,omitempty"`
}
