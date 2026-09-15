package layercache

import (
	"io"
	"os"
	"strings"
	"testing"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/cache"
	"github.com/google/go-containerregistry/pkg/v1/random"
)

func TestCacheRoundTripAndInspect(t *testing.T) {
	img, err := random.Image(512, 2)
	if err != nil {
		t.Fatal(err)
	}
	c := New(t.TempDir())
	before, err := Inspect(img, c)
	if err != nil {
		t.Fatal(err)
	}
	if before.CachedLayers != 0 || before.MissingLayers != 2 {
		t.Fatalf("before %+v", before)
	}

	wrapped := c.Wrap(img)
	layers, err := wrapped.Layers()
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range layers {
		rc, err := l.Compressed()
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.Copy(io.Discard, rc); err != nil {
			t.Fatal(err)
		}
		if err := rc.Close(); err != nil {
			t.Fatal(err)
		}
	}

	after, err := Inspect(img, c)
	if err != nil {
		t.Fatal(err)
	}
	if after.CachedLayers != 2 || after.MissingLayers != 0 {
		t.Fatalf("after %+v", after)
	}

	hit := c.Wrap(img)
	hl, err := hit.Layers()
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range hl {
		d, err := l.Digest()
		if err != nil {
			t.Fatal(err)
		}
		got, err := c.Get(d)
		if err != nil {
			t.Fatal(err)
		}
		sz, err := l.Size()
		if err != nil {
			t.Fatal(err)
		}
		gs, err := got.Size()
		if err != nil {
			t.Fatal(err)
		}
		if gs != sz {
			t.Fatalf("size %d != %d", gs, sz)
		}
	}
}

func TestIncompleteFileIsNotAHit(t *testing.T) {
	img, err := random.Image(256, 1)
	if err != nil {
		t.Fatal(err)
	}
	c := New(t.TempDir())
	layers, err := img.Layers()
	if err != nil {
		t.Fatal(err)
	}
	d, err := layers[0].Digest()
	if err != nil {
		t.Fatal(err)
	}
	sz, err := layers[0].Size()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(c.file(d), []byte("partial"), 0o600); err != nil {
		t.Fatal(err)
	}
	if c.Has(d, sz) {
		t.Fatal("truncated file should not hit")
	}
	st, err := Inspect(img, c)
	if err != nil {
		t.Fatal(err)
	}
	if st.CachedLayers != 0 {
		t.Fatalf("inspect %+v", st)
	}
	if _, err := os.Stat(c.file(d)); !os.IsNotExist(err) {
		t.Fatal("mismatched blob should be removed")
	}
}

func TestPartialCloseDoesNotCommit(t *testing.T) {
	img, err := random.Image(2048, 1)
	if err != nil {
		t.Fatal(err)
	}
	c := New(t.TempDir())
	wrapped := c.Wrap(img)
	layers, err := wrapped.Layers()
	if err != nil {
		t.Fatal(err)
	}
	rc, err := layers[0].Compressed()
	if err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 16)
	if _, err := rc.Read(buf); err != nil {
		t.Fatal(err)
	}
	_ = rc.Close()
	st, err := Inspect(img, c)
	if err != nil {
		t.Fatal(err)
	}
	if st.CachedLayers != 0 {
		t.Fatalf("partial read committed: %+v", st)
	}
}

func TestUsageAndClear(t *testing.T) {
	img, err := random.Image(256, 2)
	if err != nil {
		t.Fatal(err)
	}
	c := New(t.TempDir())
	wrapped := c.Wrap(img)
	layers, err := wrapped.Layers()
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range layers {
		rc, err := l.Compressed()
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.Copy(io.Discard, rc); err != nil {
			t.Fatal(err)
		}
		if err := rc.Close(); err != nil {
			t.Fatal(err)
		}
	}
	used := c.Usage()
	if used.Files != 2 || used.Bytes <= 0 {
		t.Fatalf("usage %+v", used)
	}
	removed, err := c.Clear()
	if err != nil {
		t.Fatal(err)
	}
	if removed.Files != 2 {
		t.Fatalf("removed %+v", removed)
	}
	after := c.Usage()
	if after.Files != 0 || after.Bytes != 0 {
		t.Fatalf("after %+v", after)
	}
	st, err := Inspect(img, c)
	if err != nil {
		t.Fatal(err)
	}
	if st.CachedLayers != 0 {
		t.Fatalf("cache still hot %+v", st)
	}
}

func TestRememberAndPruneUnreferenced(t *testing.T) {
	imgKeep, err := random.Image(256, 2)
	if err != nil {
		t.Fatal(err)
	}
	imgDrop, err := random.Image(256, 2)
	if err != nil {
		t.Fatal(err)
	}
	c := New(t.TempDir())
	populate := func(img v1.Image) {
		t.Helper()
		wrapped := c.Wrap(img)
		layers, err := wrapped.Layers()
		if err != nil {
			t.Fatal(err)
		}
		for _, l := range layers {
			rc, err := l.Compressed()
			if err != nil {
				t.Fatal(err)
			}
			if _, err := io.Copy(io.Discard, rc); err != nil {
				t.Fatal(err)
			}
			if err := rc.Close(); err != nil {
				t.Fatal(err)
			}
		}
	}
	populate(imgKeep)
	populate(imgDrop)
	if err := c.Remember("harbor.example/app:keep", "sha256:"+strings.Repeat("a", 64), "linux/amd64", imgKeep); err != nil {
		t.Fatal(err)
	}
	if err := c.Remember("harbor.example/app:drop", "sha256:"+strings.Repeat("b", 64), "linux/amd64", imgDrop); err != nil {
		t.Fatal(err)
	}
	used := c.Usage()
	if used.Files != 4 || used.Images != 2 || used.UnreferencedFiles != 0 {
		t.Fatalf("indexed %+v", used)
	}

	if err := c.Remember("harbor.example/app:drop", "sha256:"+strings.Repeat("c", 64), "linux/amd64", imgKeep); err != nil {
		t.Fatal(err)
	}
	used = c.Usage()
	if used.Images != 2 {
		t.Fatalf("still two image keys %+v", used)
	}
	if used.UnreferencedFiles != 2 {
		t.Fatalf("drop-only layers should be unreferenced %+v", used)
	}
	removed, err := c.ClearUnreferenced()
	if err != nil {
		t.Fatal(err)
	}
	if removed.Files != 2 {
		t.Fatalf("removed %+v", removed)
	}
	after := c.Usage()
	if after.Files != 2 || after.UnreferencedFiles != 0 || after.Images != 2 {
		t.Fatalf("after prune %+v", after)
	}
	st, err := Inspect(imgKeep, c)
	if err != nil {
		t.Fatal(err)
	}
	if st.CachedLayers != 2 {
		t.Fatalf("keep layers gone %+v", st)
	}
	st, err = Inspect(imgDrop, c)
	if err != nil {
		t.Fatal(err)
	}
	if st.CachedLayers != 0 {
		t.Fatalf("drop layers still cached %+v", st)
	}
}

func TestInventoryAndDeleteLayer(t *testing.T) {
	img, err := random.Image(256, 2)
	if err != nil {
		t.Fatal(err)
	}
	c := New(t.TempDir())
	wrapped := c.Wrap(img)
	layers, err := wrapped.Layers()
	if err != nil {
		t.Fatal(err)
	}
	var digests []string
	for _, l := range layers {
		d, err := l.Digest()
		if err != nil {
			t.Fatal(err)
		}
		digests = append(digests, d.String())
		rc, err := l.Compressed()
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.Copy(io.Discard, rc); err != nil {
			t.Fatal(err)
		}
		if err := rc.Close(); err != nil {
			t.Fatal(err)
		}
	}
	if err := c.Remember("harbor.example/app:1", "sha256:"+strings.Repeat("d", 64), "linux/amd64", img); err != nil {
		t.Fatal(err)
	}
	inv := c.Inventory()
	if len(inv.Layers) != 2 || len(inv.Images) != 1 {
		t.Fatalf("inventory layers=%d images=%d", len(inv.Layers), len(inv.Images))
	}
	if !inv.Layers[0].Referenced && !inv.Layers[1].Referenced {
		t.Fatal("expected referenced layers")
	}
	if len(inv.Layers[0].Images) != 1 {
		t.Fatalf("images on layer %+v", inv.Layers[0].Images)
	}
	if err := c.DeleteLayer(digests[0]); err != nil {
		t.Fatal(err)
	}
	inv = c.Inventory()
	if len(inv.Layers) != 1 {
		t.Fatalf("after delete layers=%d", len(inv.Layers))
	}
	if inv.Layers[0].Digest != digests[1] {
		t.Fatalf("remaining %s", inv.Layers[0].Digest)
	}
	if inv.Images[0].LayerCount != 1 || inv.Images[0].CachedCount != 1 {
		t.Fatalf("image record %+v", inv.Images[0])
	}
	if err := c.DeleteLayer(digests[1]); err != nil {
		t.Fatal(err)
	}
	inv = c.Inventory()
	if len(inv.Layers) != 0 || len(inv.Images) != 0 {
		t.Fatalf("expected empty after last layer %+v %+v", inv.Layers, inv.Images)
	}
}

func TestGetMissing(t *testing.T) {
	c := New(t.TempDir())
	h, err := v1.NewHash("sha256:" + strings.Repeat("0", 64))
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.Get(h)
	if err != cache.ErrNotFound {
		t.Fatalf("got %v", err)
	}
}

func TestStatsMessage(t *testing.T) {
	if (Stats{}).Message() != "正在下载并导出 TAR" {
		t.Fatal("empty")
	}
	if (Stats{TotalLayers: 3, CachedLayers: 3}).Message() != "全部 3 层已在缓存，正在导出 TAR" {
		t.Fatal("full")
	}
	got := (Stats{TotalLayers: 5, CachedLayers: 4, MissingLayers: 1}).Message()
	if got != "正在下载缺失层并导出 TAR（缓存 4/5 层）" {
		t.Fatal(got)
	}
}
