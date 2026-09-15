package download

import (
	"sync"
	"time"

	v1 "github.com/google/go-containerregistry/pkg/v1"

	"HarborDownloader/internal/model"
)

type Aggregator struct {
	mu            sync.Mutex
	taskID        string
	phase         string
	message       string
	layerSizes    []int64
	totalLayers   int
	cachedLayers  int
	missingLayers int
	lastComplete  int64
	lastTime      time.Time
	speed         int64
	runID         string
	jobName       string
	image         string
	imageIndex    int
	imageTotal    int
}

func NewAggregator(taskID string, layerSizes []int64) *Aggregator {
	return &Aggregator{
		taskID:      taskID,
		phase:       model.StatusDownloading,
		layerSizes:  layerSizes,
		totalLayers: len(layerSizes),
		lastTime:    time.Now(),
	}
}

func (a *Aggregator) SetPhase(phase, message string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.phase = phase
	a.message = message
}

func (a *Aggregator) SetCache(cached, missing int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cachedLayers = cached
	a.missingLayers = missing
}

func (a *Aggregator) SetMeta(meta progressMeta) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.runID = meta.RunID
	a.jobName = meta.JobName
	a.image = meta.Image
	a.imageIndex = meta.Index
	a.imageTotal = meta.Total
}

func (a *Aggregator) EventFromUpdate(u v1.Update) model.ProgressEvent {
	a.mu.Lock()
	defer a.mu.Unlock()
	now := time.Now()
	dt := now.Sub(a.lastTime).Seconds()
	if dt >= 0.2 && u.Complete >= a.lastComplete {
		inst := float64(u.Complete-a.lastComplete) / dt
		if a.speed == 0 {
			a.speed = int64(inst)
		} else {
			a.speed = int64(0.35*inst + 0.65*float64(a.speed))
		}
		a.lastComplete = u.Complete
		a.lastTime = now
	}
	return a.snapshotLocked(u.Complete, u.Total)
}

func (a *Aggregator) Snapshot(complete, total int64) model.ProgressEvent {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.snapshotLocked(complete, total)
}

func (a *Aggregator) snapshotLocked(complete, total int64) model.ProgressEvent {
	var pct float64
	if total > 0 {
		pct = float64(complete) / float64(total) * 100
		if pct > 100 {
			pct = 100
		}
	}
	var eta int64
	if a.speed > 0 && total > complete {
		eta = (total - complete) / a.speed
	}
	return model.ProgressEvent{
		TaskID:          a.taskID,
		Phase:           a.phase,
		Current:         complete,
		Total:           total,
		Percentage:      pct,
		SpeedBytes:      a.speed,
		DownloadedBytes: complete,
		ETASeconds:      eta,
		CurrentLayer:    currentLayer(complete, a.layerSizes),
		TotalLayers:     a.totalLayers,
		CachedLayers:    a.cachedLayers,
		MissingLayers:   a.missingLayers,
		Message:         a.message,
		RunID:           a.runID,
		JobName:         a.jobName,
		Image:           a.image,
		ImageIndex:      a.imageIndex,
		ImageTotal:      a.imageTotal,
	}
}

func currentLayer(complete int64, sizes []int64) int {
	if len(sizes) == 0 {
		return 0
	}
	if complete <= 0 {
		return 1
	}
	var acc int64
	for i, s := range sizes {
		acc += s
		if complete < acc {
			return i + 1
		}
	}
	return len(sizes)
}
