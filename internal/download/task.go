package download

import (
	"context"
	"sync"

	"HarborDownloader/internal/apperr"
	"HarborDownloader/internal/model"
)

type running struct {
	cancel context.CancelFunc
	jobID  string
	task   *model.DownloadTask
}

type Store struct {
	mu     sync.Mutex
	active *running
	tasks  sync.Map
}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) Begin(task *model.DownloadTask, parent context.Context) (context.Context, error) {
	if task == nil {
		return nil, apperr.New(apperr.CodeInvalid, "任务无效。", nil)
	}
	return s.begin(task.ID, task, parent)
}

func (s *Store) BeginJob(jobID string, parent context.Context) (context.Context, error) {
	if jobID == "" {
		return nil, apperr.New(apperr.CodeInvalid, "任务无效。", nil)
	}
	return s.begin(jobID, nil, parent)
}

func (s *Store) begin(jobID string, task *model.DownloadTask, parent context.Context) (context.Context, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.active != nil {
		return nil, apperr.New(apperr.CodeInvalid, "已有下载任务正在进行，同时只能运行一个任务。", nil)
	}
	ctx, cancel := context.WithCancel(parent)
	s.active = &running{cancel: cancel, jobID: jobID, task: task}
	if task != nil {
		copied := *task
		s.tasks.Store(task.ID, copied)
	}
	return ctx, nil
}

func (s *Store) SetTask(task *model.DownloadTask) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.active != nil {
		s.active.task = task
	}
	if task != nil {
		copied := *task
		s.tasks.Store(task.ID, copied)
	}
}

func (s *Store) Update(task *model.DownloadTask) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.active != nil && s.active.task != nil && s.active.task.ID == task.ID {
		s.active.task = task
	}
	copied := *task
	s.tasks.Store(task.ID, copied)
}

func (s *Store) Get(id string) (model.DownloadTask, bool) {
	v, ok := s.tasks.Load(id)
	if !ok {
		return model.DownloadTask{}, false
	}
	return v.(model.DownloadTask), true
}

func (s *Store) Cancel(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.active == nil {
		return apperr.New(apperr.CodeInvalid, "没有可取消的任务。", nil)
	}
	matchJob := s.active.jobID == id
	matchTask := s.active.task != nil && s.active.task.ID == id
	if !matchJob && !matchTask {
		return apperr.New(apperr.CodeInvalid, "没有可取消的任务。", nil)
	}
	if s.active.task != nil && isTerminal(s.active.task.Status) && matchTask && !matchJob {
		return apperr.New(apperr.CodeInvalid, "任务已经结束。", nil)
	}
	s.active.cancel()
	return nil
}

func (s *Store) Finish(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.active == nil {
		return
	}
	if s.active.jobID == id || (s.active.task != nil && s.active.task.ID == id) {
		s.active = nil
	}
}

func (s *Store) ActiveID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.active == nil {
		return ""
	}
	if s.active.task != nil {
		return s.active.task.ID
	}
	return s.active.jobID
}

func (s *Store) ActiveJobID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.active == nil {
		return ""
	}
	return s.active.jobID
}

func isTerminal(status string) bool {
	return status == model.StatusCompleted || status == model.StatusFailed || status == model.StatusCancelled
}
