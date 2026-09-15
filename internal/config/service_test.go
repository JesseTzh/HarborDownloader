package config

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"HarborDownloader/internal/model"
)

func TestSavePersistsRegistryCredentials(t *testing.T) {
	dir := t.TempDir()
	s := &Service{path: filepath.Join(dir, "config.json")}
	err := s.Save(model.FileConfig{
		Registry:  "harbor.company.local",
		Username:  "robot$project+download",
		Password:  "s3cret",
		Insecure:  true,
		OutputDir: `D:\docker-images`,
	})
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(s.path)
	if err != nil {
		t.Fatal(err)
	}
	var cfg model.FileConfig
	if err := json.Unmarshal(b, &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Registry != "harbor.company.local" || cfg.Username != "robot$project+download" || cfg.Password != "s3cret" || !cfg.Insecure {
		t.Fatalf("%+v", cfg)
	}
	loaded := s.Load()
	if loaded.OutputDir != `D:\docker-images` || loaded.Password != "s3cret" || !loaded.Insecure {
		t.Fatalf("%+v", loaded)
	}
}

func TestUpdatePreservesUnrelatedFields(t *testing.T) {
	dir := t.TempDir()
	s := &Service{path: filepath.Join(dir, "config.json")}
	if err := s.Save(model.FileConfig{
		Registry:  "harbor.company.local",
		Username:  "alice",
		Password:  "pw",
		Insecure:  true,
		OutputDir: `D:\docker-images`,
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.Update(func(c *model.FileConfig) {
		c.Registry = "harbor.other.local"
	}); err != nil {
		t.Fatal(err)
	}
	loaded := s.Load()
	if loaded.Registry != "harbor.other.local" {
		t.Fatalf("registry: %+v", loaded)
	}
	if loaded.Username != "alice" || loaded.Password != "pw" || !loaded.Insecure || loaded.OutputDir != `D:\docker-images` {
		t.Fatalf("lost fields: %+v", loaded)
	}
}

func TestJobsPersistAndSurviveUnrelatedUpdate(t *testing.T) {
	dir := t.TempDir()
	s := &Service{path: filepath.Join(dir, "config.json")}
	job := model.Job{
		ID:   "job-1",
		Name: "demo",
		Pack: true,
		Images: []model.JobImage{
			{Image: "project/gateway:1", TargetTag: "gateway:1.0.0", Platform: model.DefaultPlatform()},
			{Image: "project/auth:1", Platform: model.Platform{OS: "linux", Architecture: "arm64"}},
		},
	}
	if err := s.Save(model.FileConfig{
		Registry:  "harbor.company.local",
		OutputDir: `D:\docker-images`,
		Jobs:      []model.Job{job},
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.Update(func(c *model.FileConfig) {
		c.Username = "alice"
	}); err != nil {
		t.Fatal(err)
	}
	loaded := s.Load()
	if loaded.Username != "alice" {
		t.Fatalf("username: %+v", loaded)
	}
	if len(loaded.Jobs) != 1 || loaded.Jobs[0].Name != "demo" || len(loaded.Jobs[0].Images) != 2 || !loaded.Jobs[0].Pack {
		t.Fatalf("jobs: %+v", loaded.Jobs)
	}
	if loaded.Jobs[0].Images[0].TargetTag != "gateway:1.0.0" {
		t.Fatalf("target tag: %+v", loaded.Jobs[0].Images[0])
	}
}

func TestDecodeAcceptsEncryptedExport(t *testing.T) {
	plain := []byte(`{
  "registry": "harbor.local",
  "username": "alice",
  "password": "s3cret",
  "jobs": [{"name":"demo","images":[{"image":"project/a:1","platform":{"os":"linux","architecture":"amd64"}}]}]
}`)
	enc, err := EncryptExport(plain)
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := Decode(enc)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Registry != "harbor.local" || cfg.Username != "alice" || cfg.Password != "s3cret" {
		t.Fatalf("%+v", cfg)
	}
	if cfg.HasPassword {
		t.Fatal("HasPassword should not survive decode")
	}
}

func TestDecodeRejectsUnknownJSON(t *testing.T) {
	if _, err := Decode([]byte(`{"foo":1}`)); err == nil {
		t.Fatal("expected error")
	}
	if _, err := Decode([]byte(`[]`)); err == nil {
		t.Fatal("expected error")
	}
	if _, err := Decode([]byte(``)); err == nil {
		t.Fatal("expected error")
	}
	if _, err := Decode([]byte(`not json`)); err == nil {
		t.Fatal("expected error")
	}
}

func TestDecodeRejectsInvalidJob(t *testing.T) {
	_, err := Decode([]byte(`{"registry":"harbor.local","jobs":[{"name":"","outputDir":"/tmp","images":[{"image":"project/a:1","platform":{"os":"linux","architecture":"amd64"}}]}]}`))
	if err == nil {
		t.Fatal("expected invalid job")
	}
}

func TestDecodeAcceptsJobWithoutOutputDir(t *testing.T) {
	cfg, err := Decode([]byte(`{"registry":"harbor.local","jobs":[{"name":"demo","pack":true,"images":[{"image":"project/a:1","platform":{"os":"linux","architecture":"amd64"}}]}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Jobs) != 1 || cfg.Jobs[0].Name != "demo" || !cfg.Jobs[0].Pack {
		t.Fatalf("%+v", cfg.Jobs)
	}
}

func TestSaveOmitsHasPassword(t *testing.T) {
	dir := t.TempDir()
	s := &Service{path: filepath.Join(dir, "config.json")}
	if err := s.Save(model.FileConfig{
		Registry:    "harbor.company.local",
		Password:    "s3cret",
		HasPassword: true,
		OutputDir:   `D:\docker-images`,
	}); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(s.path)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(b, []byte("hasPassword")) {
		t.Fatalf("hasPassword persisted: %s", b)
	}
}

func TestImportExportRoundTrip(t *testing.T) {
	dir := t.TempDir()
	s := &Service{path: filepath.Join(dir, "config.json")}
	original := model.FileConfig{
		Registry:  "harbor.company.local",
		Username:  "robot$project+download",
		Password:  "s3cret",
		Insecure:  true,
		OutputDir: `D:\docker-images`,
		Jobs: []model.Job{{
			ID:   "job-1",
			Name: "demo",
			Pack: true,
			Images: []model.JobImage{
				{Image: "project/gateway:1", Platform: model.DefaultPlatform()},
			},
		}},
	}
	if err := s.Save(original); err != nil {
		t.Fatal(err)
	}
	exported, err := s.ExportBytes()
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(exported, []byte("s3cret")) {
		t.Fatalf("password leaked in export: %s", exported)
	}
	var env map[string]any
	if err := json.Unmarshal(exported, &env); err != nil {
		t.Fatal(err)
	}
	if env["kind"] != "HarborDownloader.config" {
		t.Fatalf("expected encrypted envelope: %s", exported)
	}
	other := &Service{path: filepath.Join(dir, "imported.json")}
	loaded, err := other.ImportBytes(exported)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Registry != original.Registry || loaded.Username != original.Username || loaded.Password != original.Password || !loaded.Insecure {
		t.Fatalf("registry: %+v", loaded)
	}
	if len(loaded.Jobs) != 1 || loaded.Jobs[0].ID != "job-1" || loaded.Jobs[0].Name != "demo" || !loaded.Jobs[0].Pack {
		t.Fatalf("jobs: %+v", loaded.Jobs)
	}
}

func TestImportReplacesExistingConfig(t *testing.T) {
	dir := t.TempDir()
	s := &Service{path: filepath.Join(dir, "config.json")}
	if err := s.Save(model.FileConfig{
		Registry:  "old.local",
		Username:  "old",
		Password:  "oldpw",
		OutputDir: `D:\old`,
		Jobs:      []model.Job{{ID: "old", Name: "old", Images: []model.JobImage{{Image: "old/a:1", Platform: model.DefaultPlatform()}}}},
	}); err != nil {
		t.Fatal(err)
	}
	incoming := []byte(`{
  "registry": "new.local",
  "username": "alice",
  "password": "pw",
  "insecure": true,
  "outputDir": "D:\\new",
  "jobs": [
    {
      "id": "job-2",
      "name": "gateway",
      "pack": true,
      "images": [{"image": "project/gateway:2", "platform": {"os": "linux", "architecture": "arm64"}}]
    }
  ]
}`)
	loaded, err := s.ImportBytes(incoming)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Registry != "new.local" || loaded.Username != "alice" || loaded.Password != "pw" || !loaded.Insecure {
		t.Fatalf("registry: %+v", loaded)
	}
	if len(loaded.Jobs) != 1 || loaded.Jobs[0].ID != "job-2" || loaded.Jobs[0].Images[0].Platform.Architecture != "arm64" || !loaded.Jobs[0].Pack {
		t.Fatalf("jobs: %+v", loaded.Jobs)
	}
}
