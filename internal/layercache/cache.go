package layercache

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/cache"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

type Cache struct {
	dir string
	mu  sync.Mutex
}

func New(dir string) *Cache {
	return &Cache{dir: dir}
}

func (c *Cache) Path() string {
	if c == nil {
		return ""
	}
	return c.dir
}

func (c *Cache) Wrap(img v1.Image) v1.Image {
	if c == nil || img == nil {
		return img
	}
	return cache.Image(img, c)
}

func (c *Cache) file(h v1.Hash) string {
	var name string
	if runtime.GOOS == "windows" {
		name = fmt.Sprintf("%s-%s", h.Algorithm, h.Hex)
	} else {
		name = h.String()
	}
	return filepath.Join(c.dir, name)
}

func (c *Cache) Has(h v1.Hash, size int64) bool {
	if c == nil || h.Hex == "" {
		return false
	}
	st, err := os.Stat(c.file(h))
	if err != nil || st.Size() <= 0 {
		return false
	}
	if size > 0 && st.Size() != size {
		return false
	}
	return true
}

func (c *Cache) Get(h v1.Hash) (v1.Layer, error) {
	if c == nil {
		return nil, cache.ErrNotFound
	}
	path := c.file(h)
	st, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, cache.ErrNotFound
		}
		return nil, err
	}
	if st.Size() <= 0 {
		_ = os.Remove(path)
		return nil, cache.ErrNotFound
	}
	return &fileLayer{path: path, digest: h, size: st.Size()}, nil
}

func (c *Cache) Put(l v1.Layer) (v1.Layer, error) {
	if c == nil || l == nil {
		return l, nil
	}
	digest, err := l.Digest()
	if err != nil {
		return nil, err
	}
	size, err := l.Size()
	if err != nil {
		size = 0
	}
	mt, err := l.MediaType()
	if err != nil {
		mt = types.DockerLayer
	}
	return &cachingLayer{inner: l, c: c, digest: digest, size: size, mediaType: mt}, nil
}

func (c *Cache) Delete(h v1.Hash) error {
	if c == nil {
		return nil
	}
	err := os.Remove(c.file(h))
	if os.IsNotExist(err) {
		return cache.ErrNotFound
	}
	return err
}

func (c *Cache) cleanupPartials() {
	if c == nil {
		return
	}
	entries, err := os.ReadDir(c.dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.Contains(name, ".partial") {
			_ = os.Remove(filepath.Join(c.dir, name))
		}
	}
}

type fileLayer struct {
	path      string
	digest    v1.Hash
	size      int64
	mediaType types.MediaType
}

func (l *fileLayer) Digest() (v1.Hash, error) { return l.digest, nil }
func (l *fileLayer) Size() (int64, error)     { return l.size, nil }
func (l *fileLayer) MediaType() (types.MediaType, error) {
	if l.mediaType == "" {
		return types.DockerLayer, nil
	}
	return l.mediaType, nil
}

func (l *fileLayer) Compressed() (io.ReadCloser, error) {
	return os.Open(l.path)
}

func (l *fileLayer) Uncompressed() (io.ReadCloser, error) {
	f, err := os.Open(l.path)
	if err != nil {
		return nil, err
	}
	gz, err := gzip.NewReader(f)
	if err != nil {
		f.Close()
		return nil, err
	}
	return &gzipCloser{gz: gz, f: f}, nil
}

func (l *fileLayer) DiffID() (v1.Hash, error) {
	rc, err := l.Uncompressed()
	if err != nil {
		return v1.Hash{}, err
	}
	defer rc.Close()
	h, _, err := v1.SHA256(rc)
	return h, err
}

type gzipCloser struct {
	gz *gzip.Reader
	f  *os.File
}

func (g *gzipCloser) Read(p []byte) (int, error) { return g.gz.Read(p) }

func (g *gzipCloser) Close() error {
	err := g.gz.Close()
	if cerr := g.f.Close(); err == nil {
		err = cerr
	}
	return err
}

type cachingLayer struct {
	inner     v1.Layer
	c         *Cache
	digest    v1.Hash
	size      int64
	mediaType types.MediaType
}

func (l *cachingLayer) Digest() (v1.Hash, error)             { return l.digest, nil }
func (l *cachingLayer) Size() (int64, error)                 { return l.inner.Size() }
func (l *cachingLayer) DiffID() (v1.Hash, error)             { return l.inner.DiffID() }
func (l *cachingLayer) MediaType() (types.MediaType, error)  { return l.inner.MediaType() }
func (l *cachingLayer) Uncompressed() (io.ReadCloser, error) { return l.inner.Uncompressed() }

func (l *cachingLayer) Compressed() (io.ReadCloser, error) {
	if l.c.Has(l.digest, l.size) {
		fl, err := l.c.Get(l.digest)
		if err == nil {
			return fl.Compressed()
		}
		if !errors.Is(err, cache.ErrNotFound) {
			return nil, err
		}
	}
	rc, err := l.inner.Compressed()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(l.c.dir, 0o700); err != nil {
		rc.Close()
		return nil, err
	}
	tmp, err := os.CreateTemp(l.c.dir, l.digest.Hex+".*.partial")
	if err != nil {
		rc.Close()
		return nil, err
	}
	cw := &countWriter{w: tmp}
	return &populateCloser{
		Reader:   io.TeeReader(rc, cw),
		inner:    rc,
		file:     tmp,
		final:    l.c.file(l.digest),
		expected: l.size,
		counter:  cw,
	}, nil
}

type countWriter struct {
	w io.Writer
	n int64
}

func (c *countWriter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	c.n += int64(n)
	return n, err
}

type populateCloser struct {
	io.Reader
	inner    io.Closer
	file     *os.File
	final    string
	expected int64
	counter  *countWriter
}

func (p *populateCloser) Close() error {
	innerErr := p.inner.Close()
	name := p.file.Name()
	closeErr := p.file.Close()
	n := p.counter.n
	complete := innerErr == nil && closeErr == nil && n > 0 && (p.expected <= 0 || n == p.expected)
	if !complete {
		_ = os.Remove(name)
		if innerErr != nil {
			return innerErr
		}
		return closeErr
	}
	_ = os.Remove(p.final)
	if err := os.Rename(name, p.final); err != nil {
		_ = os.Remove(name)
		return err
	}
	return nil
}
