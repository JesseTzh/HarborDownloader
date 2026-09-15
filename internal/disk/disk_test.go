package disk

import (
	"path/filepath"
	"testing"
)

func TestRequiredBytes(t *testing.T) {
	out, cache := RequiredBytes(100, 40, true)
	if out != 140 || cache != 0 {
		t.Fatalf("same volume %d %d", out, cache)
	}
	out, cache = RequiredBytes(100, 40, false)
	if out != 100 || cache != 40 {
		t.Fatalf("split %d %d", out, cache)
	}
	out, cache = RequiredBytes(100, 0, true)
	if out != 100 || cache != 0 {
		t.Fatalf("no missing %d %d", out, cache)
	}
}

func TestSameVolumeTemp(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "out")
	b := filepath.Join(dir, "cache")
	same, err := SameVolume(dir, dir)
	if err != nil {
		t.Fatal(err)
	}
	if !same {
		t.Fatal("temp dir should be same volume")
	}
	_ = a
	_ = b
}
