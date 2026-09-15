package download

import (
	"context"
	"testing"
	"time"

	"HarborDownloader/internal/model"
)

func TestStartJobRejectsInvalid(t *testing.T) {
	s := NewService(context.Background(), nil, nil)
	_, err := s.StartJob(model.Job{Name: "", OutputDir: t.TempDir()}, model.RegistryConfig{})
	if err == nil {
		t.Fatal("empty name should fail")
	}
	_, err = s.StartJob(model.Job{Name: "demo", OutputDir: t.TempDir()}, model.RegistryConfig{})
	if err == nil {
		t.Fatal("no images should fail")
	}
	_, err = s.StartJob(model.Job{
		Name:   "demo",
		Images: []model.JobImage{{Image: "project/a:1", Platform: model.DefaultPlatform()}},
	}, model.RegistryConfig{})
	if err == nil {
		t.Fatal("missing output dir should fail")
	}
	_, err = s.StartJob(model.Job{
		Name:      "demo",
		OutputDir: t.TempDir(),
		Images: []model.JobImage{{
			Image:     "project/a:1",
			TargetTag: "app@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			Platform:  model.DefaultPlatform(),
		}},
	}, model.RegistryConfig{})
	if err == nil {
		t.Fatal("invalid target tag should fail")
	}
}

func TestCancelJobByRunID(t *testing.T) {
	st := NewStore()
	ctx, err := st.BeginJob("run-1", context.Background())
	if err != nil {
		t.Fatal(err)
	}
	task := &model.DownloadTask{ID: "img-1", Status: model.StatusDownloading, StartedAt: time.Now()}
	st.SetTask(task)
	if err := st.Cancel("run-1"); err != nil {
		t.Fatal(err)
	}
	select {
	case <-ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("job context was not cancelled")
	}
}

func TestBeginJobRejectsSecond(t *testing.T) {
	st := NewStore()
	if _, err := st.BeginJob("a", context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := st.BeginJob("b", context.Background()); err == nil {
		t.Fatal("second job should be rejected")
	}
}
