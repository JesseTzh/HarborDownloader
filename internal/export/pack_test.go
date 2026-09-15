package export

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/klauspost/compress/zstd"
)

func TestPackTarsMergesAndCompresses(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.tar"), []byte("hello-a"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.tar"), bytes.Repeat([]byte("b"), 4096), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("skip"), 0o644); err != nil {
		t.Fatal(err)
	}

	path, size, err := PackTars(context.Background(), dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(path) != PackedName {
		t.Fatalf("path %s", path)
	}
	if size <= 0 {
		t.Fatal("empty archive")
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	dec, err := zstd.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer dec.Close()
	tr := tar.NewReader(dec)
	got := map[string][]byte{}
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		b, err := io.ReadAll(tr)
		if err != nil {
			t.Fatal(err)
		}
		got[hdr.Name] = b
	}
	if len(got) != 2 {
		t.Fatalf("members %+v", got)
	}
	if string(got["a.tar"]) != "hello-a" {
		t.Fatalf("a.tar %q", got["a.tar"])
	}
	if !bytes.Equal(got["b.tar"], bytes.Repeat([]byte("b"), 4096)) {
		t.Fatal("b.tar mismatch")
	}
}

func TestPackTarsEmptyDir(t *testing.T) {
	_, _, err := PackTars(context.Background(), t.TempDir(), nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPackTarsCancel(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.tar"), bytes.Repeat([]byte("a"), 64*1024), 0o644); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := PackTars(ctx, dir, nil); err == nil {
		t.Fatal("expected cancel")
	}
	if _, err := os.Stat(filepath.Join(dir, PackedName)); !os.IsNotExist(err) {
		t.Fatalf("leftover archive: %v", err)
	}
}
