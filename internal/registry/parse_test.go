package registry

import (
	"strings"
	"testing"

	"HarborDownloader/internal/model"
)

func TestParseImageReference(t *testing.T) {
	tests := []struct {
		name     string
		image    string
		registry string
		wantReg  string
		wantRepo string
		wantTag  string
		wantErr  bool
	}{
		{
			name:     "full reference",
			image:    "harbor.company.local/project/app-gateway:1.0.0",
			registry: "harbor.company.local",
			wantReg:  "harbor.company.local",
			wantRepo: "project/app-gateway",
			wantTag:  "1.0.0",
		},
		{
			name:     "repo only uses default registry",
			image:    "project/app-gateway:1.0.0-SNAPSHOT",
			registry: "harbor.company.local",
			wantReg:  "harbor.company.local",
			wantRepo: "project/app-gateway",
			wantTag:  "1.0.0-SNAPSHOT",
		},
		{
			name:     "strips https scheme",
			image:    "https://harbor.company.local/project/app-gateway:1.0.0",
			registry: "harbor.company.local",
			wantReg:  "harbor.company.local",
			wantRepo: "project/app-gateway",
			wantTag:  "1.0.0",
		},
		{
			name:     "path traversal rejected",
			image:    "../evil:1",
			registry: "harbor.company.local",
			wantErr:  true,
		},
		{
			name:     "empty image",
			image:    "  ",
			registry: "harbor.company.local",
			wantErr:  true,
		},
		{
			name:     "missing registry",
			image:    "project/gateway:1",
			registry: "",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := ParseImageReference(tt.image, tt.registry, false)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %+v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Registry != tt.wantReg {
				t.Errorf("registry=%s want %s", got.Registry, tt.wantReg)
			}
			if got.Repository != tt.wantRepo {
				t.Errorf("repository=%s want %s", got.Repository, tt.wantRepo)
			}
			if got.Tag != tt.wantTag {
				t.Errorf("tag=%s want %s", got.Tag, tt.wantTag)
			}
		})
	}
}

func TestNormalizeRegistry(t *testing.T) {
	n, err := NormalizeRegistry("harbor.company.local", false)
	if err != nil {
		t.Fatal(err)
	}
	if n.Host != "harbor.company.local" || n.Scheme != "https" {
		t.Fatalf("%+v", n)
	}
	_, err = NormalizeRegistry("http://harbor.company.local", false)
	if err == nil {
		t.Fatal("http without insecure should fail")
	}
	n, err = NormalizeRegistry("http://harbor.company.local", true)
	if err != nil {
		t.Fatal(err)
	}
	if n.Scheme != "http" || !n.HTTP {
		t.Fatalf("%+v", n)
	}
}

func TestParsePlatform(t *testing.T) {
	p, err := ParsePlatform(model.Platform{})
	if err != nil {
		t.Fatal(err)
	}
	if p.String() != "linux/amd64" {
		t.Fatalf("%s", p)
	}
	_, err = ParsePlatform(model.Platform{OS: "windows", Architecture: "amd64"})
	if err == nil {
		t.Fatal("windows should be rejected in MVP")
	}
	p, err = ParsePlatform(model.Platform{OS: "linux", Architecture: "arm64"})
	if err != nil {
		t.Fatal(err)
	}
	if p.String() != "linux/arm64" {
		t.Fatalf("%s", p)
	}
}

func TestHasRegistryHost(t *testing.T) {
	if !hasRegistryHost("harbor.company.local/a/b:1") {
		t.Fatal("expected host")
	}
	if hasRegistryHost("project/app-gateway:1") {
		t.Fatal("project/repo should not look like a host")
	}
	if !strings.Contains("ok", "ok") {
		t.Fatal("sanity")
	}
}
