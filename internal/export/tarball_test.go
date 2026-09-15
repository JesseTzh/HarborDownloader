package export

import (
	"archive/tar"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/random"
)

func TestWriteTarball(t *testing.T) {
	img, err := random.Image(256, 1)
	if err != nil {
		t.Fatal(err)
	}
	ref, err := name.ParseReference("example.com/project/demo:1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	out := filepath.Join(dir, "demo.tar")
	if err := Write(context.Background(), out, ref, img, nil); err != nil {
		t.Fatal(err)
	}
	st, err := os.Stat(out)
	if err != nil {
		t.Fatal(err)
	}
	if st.Size() <= 0 {
		t.Fatal("empty tar")
	}
}

func TestWriteTarballRepoTagsUseOriginalTag(t *testing.T) {
	img, err := random.Image(256, 1)
	if err != nil {
		t.Fatal(err)
	}
	ref, err := name.NewTag("company/app:prod")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	out := filepath.Join(dir, "demo.tar")
	if err := Write(context.Background(), out, ref, img, nil); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(out)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	tr := tar.NewReader(f)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			t.Fatal("manifest.json missing")
		}
		if err != nil {
			t.Fatal(err)
		}
		if hdr.Name != "manifest.json" {
			continue
		}
		var m []struct {
			RepoTags []string `json:"RepoTags"`
		}
		if err := json.NewDecoder(tr).Decode(&m); err != nil {
			t.Fatal(err)
		}
		if len(m) != 1 || len(m[0].RepoTags) != 1 || m[0].RepoTags[0] != "company/app:prod" {
			t.Fatalf("RepoTags %+v", m)
		}
		return
	}
}
