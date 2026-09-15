package download

import (
	"context"
	"testing"
	"time"

	"HarborDownloader/internal/model"
)

func TestCancelTask(t *testing.T) {
	s := NewStore()
	task := &model.DownloadTask{ID: "abc", Status: model.StatusDownloading, StartedAt: time.Now()}
	ctx, err := s.Begin(task, context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Cancel("abc"); err != nil {
		t.Fatal(err)
	}
	select {
	case <-ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("context was not cancelled")
	}
}

func TestSingleActiveTask(t *testing.T) {
	s := NewStore()
	t1 := &model.DownloadTask{ID: "a", Status: model.StatusDownloading}
	if _, err := s.Begin(t1, context.Background()); err != nil {
		t.Fatal(err)
	}
	t2 := &model.DownloadTask{ID: "b", Status: model.StatusPending}
	if _, err := s.Begin(t2, context.Background()); err == nil {
		t.Fatal("second task should be rejected")
	}
}

func TestTaskLifecycleGet(t *testing.T) {
	s := NewStore()
	task := &model.DownloadTask{ID: "x", Status: model.StatusPending}
	if _, err := s.Begin(task, context.Background()); err != nil {
		t.Fatal(err)
	}
	got, ok := s.Get("x")
	if !ok || got.Status != model.StatusPending {
		t.Fatalf("%v %v", ok, got)
	}
	task.Status = model.StatusCompleted
	s.Update(task)
	s.Finish("x")
	got, ok = s.Get("x")
	if !ok || got.Status != model.StatusCompleted {
		t.Fatalf("%v %v", ok, got)
	}
	if s.ActiveID() != "" {
		t.Fatal("active should be cleared")
	}
}
