package model

import (
	"fmt"
	"strings"
	"time"
)

type JobImage struct {
	Image     string   `json:"image"`
	TargetTag string   `json:"targetTag,omitempty"`
	Platform  Platform `json:"platform"`
}

type Job struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Pack      bool       `json:"pack,omitempty"`
	OutputDir string     `json:"outputDir,omitempty"`
	Images    []JobImage `json:"images"`
}

type JobItem struct {
	Image      string `json:"image"`
	TargetTag  string `json:"targetTag,omitempty"`
	Platform   string `json:"platform"`
	Status     string `json:"status"`
	OutputPath string `json:"outputPath,omitempty"`
	Digest     string `json:"digest,omitempty"`
	Size       int64  `json:"size,omitempty"`
	Error      string `json:"error,omitempty"`
}

type JobStatusEvent struct {
	RunID       string    `json:"runId"`
	JobID       string    `json:"jobId"`
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	Current     int       `json:"current"`
	Total       int       `json:"total"`
	Items       []JobItem `json:"items"`
	PackagePath string    `json:"packagePath,omitempty"`
	Error       string    `json:"error,omitempty"`
}

type JobCompleteEvent struct {
	RunID       string    `json:"runId"`
	JobID       string    `json:"jobId"`
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	Items       []JobItem `json:"items"`
	PackagePath string    `json:"packagePath,omitempty"`
	Error       string    `json:"error,omitempty"`
}

type JobRun struct {
	ID          string    `json:"id"`
	JobID       string    `json:"jobId"`
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	Current     int       `json:"current"`
	Total       int       `json:"total"`
	Items       []JobItem `json:"items"`
	PackagePath string    `json:"packagePath,omitempty"`
	StartedAt   time.Time `json:"startedAt"`
	FinishedAt  time.Time `json:"finishedAt,omitempty"`
	Error       string    `json:"error,omitempty"`
}

func ValidateJob(job Job) error {
	if strings.TrimSpace(job.Name) == "" {
		return fmt.Errorf("请填写任务名称。")
	}
	n := 0
	for _, img := range job.Images {
		if strings.TrimSpace(img.Image) == "" {
			continue
		}
		n++
		if _, err := ParseTargetTag(img.TargetTag); err != nil {
			return err
		}
	}
	if n == 0 {
		return fmt.Errorf("请至少添加一个镜像。")
	}
	return nil
}

func NormalizeJob(job Job) Job {
	job.Name = strings.TrimSpace(job.Name)
	job.OutputDir = strings.TrimSpace(job.OutputDir)
	images := make([]JobImage, 0, len(job.Images))
	for _, img := range job.Images {
		image := strings.TrimSpace(img.Image)
		if image == "" {
			continue
		}
		plat := img.Platform
		if plat.OS == "" {
			plat = DefaultPlatform()
		}
		target := strings.TrimSpace(img.TargetTag)
		if parsed, err := ParseTargetTag(target); err == nil && parsed.Raw != "" {
			target = parsed.Raw
		}
		images = append(images, JobImage{
			Image:     image,
			TargetTag: target,
			Platform:  plat,
		})
	}
	job.Images = images
	return job
}
