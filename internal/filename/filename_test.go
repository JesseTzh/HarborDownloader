package filename

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"HarborDownloader/internal/model"
)

func TestSanitize(t *testing.T) {
	got := Sanitize(`a/b:c*d?e"f<g>h|i`)
	if strings.ContainsAny(got, `/\:*?"<>|`) {
		t.Fatalf("illegal chars remain: %s", got)
	}
	if Sanitize("..") != "image" {
		t.Fatalf(".. should become image")
	}
	if Sanitize("CON") == "CON" {
		t.Fatalf("reserved name should be prefixed")
	}
}

func TestDefaultTarName(t *testing.T) {
	ref := model.ImageReference{
		Registry:   "harbor.company.local",
		Repository: "project/app-gateway",
		Tag:        "1.0.0-SNAPSHOT",
	}
	plat := model.Platform{OS: "linux", Architecture: "amd64"}
	got := DefaultTarName(ref, plat)
	want := "app-gateway_1.0.0-SNAPSHOT_linux-amd64.tar"
	if got != want {
		t.Fatalf("got %s want %s", got, want)
	}
}

func TestDefaultTarNameForTargetTag(t *testing.T) {
	ref := model.ImageReference{
		Registry:   "harbor.company.local",
		Repository: "project/app-gateway",
		Tag:        "1.0.0-SNAPSHOT",
	}
	plat := model.Platform{OS: "linux", Architecture: "amd64"}
	got, err := DefaultTarNameFor(ref, plat, "company/app:prod")
	if err != nil {
		t.Fatal(err)
	}
	want := "app_prod_linux-amd64.tar"
	if got != want {
		t.Fatalf("got %s want %s", got, want)
	}
	got, err = DefaultTarNameFor(ref, plat, "")
	if err != nil {
		t.Fatal(err)
	}
	if got != DefaultTarName(ref, plat) {
		t.Fatalf("empty target should keep source name %s", got)
	}
}

func TestJobOutputDir(t *testing.T) {
	dir := t.TempDir()
	got, err := JobOutputDir(dir, `生产/环境:镜像`)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Dir(got) != dir {
		t.Fatalf("parent %s", got)
	}
	if strings.ContainsAny(filepath.Base(got), `/\:*?"<>|`) {
		t.Fatalf("illegal dir name %s", got)
	}
}

func TestPrepareJobOutputDirRemovesPreviousPackages(t *testing.T) {
	dir := t.TempDir()
	jobDir := filepath.Join(dir, "demo")
	if err := os.MkdirAll(jobDir, 0o755); err != nil {
		t.Fatal(err)
	}
	keep := filepath.Join(jobDir, "notes.txt")
	if err := os.WriteFile(keep, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"a.tar", "all.tar.zst", "a.tar.downloading", "all.tar.zst.tmp"} {
		if err := os.WriteFile(filepath.Join(jobDir, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := PrepareJobOutputDir(jobDir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(keep); err != nil {
		t.Fatal("notes.txt should be kept")
	}
	entries, err := os.ReadDir(jobDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "notes.txt" {
		t.Fatalf("leftover %+v", entries)
	}
}

func TestResolveOutputPath(t *testing.T) {
	dir := t.TempDir()
	p, err := ResolveOutputPath(dir, "foo.tar")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(p) != "foo.tar" {
		t.Fatal(p)
	}
	_, err = ResolveOutputPath(dir, "../escape.tar")
	if err == nil {
		t.Fatal("expected traversal reject")
	}
	_, err = ResolveOutputPath(dir, "a/b.tar")
	if err == nil {
		t.Fatal("expected nested name reject")
	}
}
