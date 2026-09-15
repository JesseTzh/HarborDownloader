package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"HarborDownloader/internal/model"
)

type Service struct {
	path string
}

func New() (*Service, error) {
	dir, err := configDir()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	return &Service{path: filepath.Join(dir, "config.json")}, nil
}

func configDir() (string, error) {
	if d := os.Getenv("APPDATA"); d != "" {
		return filepath.Join(d, "HarborDownloader"), nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "HarborDownloader"), nil
}

func DefaultOutputDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "docker-images"
	}
	return filepath.Join(home, "docker-images")
}

func (s *Service) Load() model.FileConfig {
	cfg := model.FileConfig{OutputDir: DefaultOutputDir()}
	if s == nil {
		return cfg
	}
	b, err := os.ReadFile(s.path)
	if err != nil {
		return cfg
	}
	_ = json.Unmarshal(b, &cfg)
	if cfg.OutputDir == "" {
		cfg.OutputDir = DefaultOutputDir()
	}
	return cfg
}

func (s *Service) Save(cfg model.FileConfig) error {
	if s == nil {
		return nil
	}
	if cfg.OutputDir == "" {
		cfg.OutputDir = DefaultOutputDir()
	}
	cfg.HasPassword = false
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, b, 0o600)
}

func (s *Service) Update(fn func(*model.FileConfig)) error {
	if s == nil {
		return nil
	}
	cfg := s.Load()
	fn(&cfg)
	return s.Save(cfg)
}

func (s *Service) ExportBytes() ([]byte, error) {
	if s == nil {
		return nil, fmt.Errorf("配置服务不可用")
	}
	cfg := s.Load()
	cfg.HasPassword = false
	plain, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, err
	}
	return EncryptExport(plain)
}

func (s *Service) ImportBytes(b []byte) (model.FileConfig, error) {
	if s == nil {
		return model.FileConfig{}, fmt.Errorf("配置服务不可用")
	}
	cfg, err := Decode(b)
	if err != nil {
		return model.FileConfig{}, err
	}
	if err := s.Save(cfg); err != nil {
		return model.FileConfig{}, err
	}
	return s.Load(), nil
}

func Decode(b []byte) (model.FileConfig, error) {
	b = bytes.TrimSpace(b)
	if len(b) == 0 {
		return model.FileConfig{}, fmt.Errorf("配置文件为空")
	}
	opened, err := openExport(b)
	if err != nil {
		return model.FileConfig{}, err
	}
	b = bytes.TrimSpace(opened)
	if !json.Valid(b) {
		return model.FileConfig{}, fmt.Errorf("不是有效的 JSON 文件")
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		return model.FileConfig{}, fmt.Errorf("不是 Harbor Downloader 配置文件")
	}
	if !looksLikeFileConfig(raw) {
		return model.FileConfig{}, fmt.Errorf("不是 Harbor Downloader 配置文件")
	}
	var cfg model.FileConfig
	if err := json.Unmarshal(b, &cfg); err != nil {
		return model.FileConfig{}, fmt.Errorf("配置文件无法解析")
	}
	cfg.HasPassword = false
	jobs := make([]model.Job, 0, len(cfg.Jobs))
	for i, job := range cfg.Jobs {
		job = model.NormalizeJob(job)
		if err := model.ValidateJob(job); err != nil {
			name := job.Name
			if name == "" {
				name = fmt.Sprintf("第 %d 个任务", i+1)
			}
			return model.FileConfig{}, fmt.Errorf("任务「%s」无效：%s", name, err.Error())
		}
		jobs = append(jobs, job)
	}
	cfg.Jobs = jobs
	if cfg.OutputDir == "" {
		cfg.OutputDir = DefaultOutputDir()
	}
	return cfg, nil
}

func looksLikeFileConfig(raw map[string]json.RawMessage) bool {
	for _, key := range []string{"registry", "username", "password", "insecure", "outputDir", "jobs"} {
		if _, ok := raw[key]; ok {
			return true
		}
	}
	return false
}
