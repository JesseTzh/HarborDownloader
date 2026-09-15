package model

import "testing"

func TestValidateJob(t *testing.T) {
	if err := ValidateJob(Job{}); err == nil {
		t.Fatal("empty job should fail")
	}
	job := Job{
		Name:   "示例服务",
		Images: []JobImage{{Image: "project/app-gateway:1.0.0", Platform: DefaultPlatform()}},
	}
	if err := ValidateJob(job); err != nil {
		t.Fatal(err)
	}
	job.Images = []JobImage{{Image: "  "}}
	if err := ValidateJob(job); err == nil {
		t.Fatal("blank image should fail")
	}
}

func TestNormalizeJobDropsEmptyImages(t *testing.T) {
	job := NormalizeJob(Job{
		Name:      "  demo  ",
		OutputDir: " /tmp/out ",
		Images: []JobImage{
			{Image: ""},
			{Image: "  project/a:1  "},
			{Image: "project/b:2", Platform: Platform{OS: "linux", Architecture: "arm64"}},
		},
	})
	if job.Name != "demo" || job.OutputDir != "/tmp/out" {
		t.Fatalf("%+v", job)
	}
	if len(job.Images) != 2 {
		t.Fatalf("images %+v", job.Images)
	}
	if job.Images[0].Platform.Architecture != "amd64" {
		t.Fatalf("default platform %+v", job.Images[0].Platform)
	}
	if job.Images[1].Platform.Architecture != "arm64" {
		t.Fatalf("kept platform %+v", job.Images[1].Platform)
	}
}

func TestNormalizeJobKeepsTargetTag(t *testing.T) {
	job := NormalizeJob(Job{
		Name: "demo",
		Images: []JobImage{
			{Image: "project/a:1", TargetTag: "  app-gateway:1.0.0  "},
		},
	})
	if job.Images[0].TargetTag != "app-gateway:1.0.0" {
		t.Fatalf("target tag %+v", job.Images[0])
	}
}

func TestValidateJobRejectsDigestTargetTag(t *testing.T) {
	job := Job{
		Name: "demo",
		Images: []JobImage{{
			Image:     "project/a:1",
			TargetTag: "app@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			Platform:  DefaultPlatform(),
		}},
	}
	if err := ValidateJob(job); err == nil {
		t.Fatal("digest target tag should fail")
	}
}

func TestNormalizeJobKeepsPack(t *testing.T) {
	job := NormalizeJob(Job{
		Name: "demo",
		Pack: true,
		Images: []JobImage{
			{Image: "project/a:1"},
		},
	})
	if !job.Pack {
		t.Fatal("pack should be kept")
	}
}
