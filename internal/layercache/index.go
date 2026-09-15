package layercache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

type ImageRecord struct {
	Image     string    `json:"image"`
	Digest    string    `json:"digest,omitempty"`
	Platform  string    `json:"platform"`
	Layers    []string  `json:"layers"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type indexFile struct {
	Images []ImageRecord `json:"images"`
}

func (c *Cache) indexPath() string {
	if c == nil || c.dir == "" {
		return ""
	}
	return filepath.Join(c.dir, "index.json")
}

func (c *Cache) Remember(image, digest, platform string, img v1.Image) error {
	if c == nil || image == "" || img == nil {
		return nil
	}
	layers, err := img.Layers()
	if err != nil {
		return err
	}
	digests := make([]string, 0, len(layers))
	seen := map[string]struct{}{}
	for _, l := range layers {
		h, err := l.Digest()
		if err != nil {
			return err
		}
		s := h.String()
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		digests = append(digests, s)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	idx, err := c.loadIndexLocked()
	if err != nil {
		return err
	}
	rec := ImageRecord{
		Image:     image,
		Digest:    digest,
		Platform:  platform,
		Layers:    digests,
		UpdatedAt: time.Now().UTC(),
	}
	replaced := false
	for i, existing := range idx.Images {
		if existing.Image == image && existing.Platform == platform {
			idx.Images[i] = rec
			replaced = true
			break
		}
	}
	if !replaced {
		idx.Images = append(idx.Images, rec)
	}
	return c.saveIndexLocked(idx)
}

func (c *Cache) referencedLocked() (map[string]struct{}, int, error) {
	idx, err := c.loadIndexLocked()
	if err != nil {
		return nil, 0, err
	}
	refs := make(map[string]struct{})
	for _, rec := range idx.Images {
		for _, layer := range rec.Layers {
			refs[layer] = struct{}{}
		}
	}
	return refs, len(idx.Images), nil
}

func (c *Cache) loadIndexLocked() (*indexFile, error) {
	path := c.indexPath()
	if path == "" {
		return &indexFile{}, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &indexFile{}, nil
		}
		return nil, err
	}
	if len(b) == 0 {
		return &indexFile{}, nil
	}
	var idx indexFile
	if err := json.Unmarshal(b, &idx); err != nil {
		return nil, err
	}
	if idx.Images == nil {
		idx.Images = []ImageRecord{}
	}
	return &idx, nil
}

func (c *Cache) saveIndexLocked(idx *indexFile) error {
	path := c.indexPath()
	if path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), "index.*.json.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	_ = os.Remove(path)
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return err
	}
	return nil
}

func (c *Cache) removeIndexLocked() error {
	path := c.indexPath()
	if path == "" {
		return nil
	}
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func hashFromFileName(name string) (v1.Hash, bool) {
	if name == "" || name == "index.json" || strings.Contains(name, ".partial") {
		return v1.Hash{}, false
	}
	key := name
	if strings.Count(name, "-") == 1 && !strings.Contains(name, ":") {
		parts := strings.SplitN(name, "-", 2)
		if len(parts) == 2 {
			key = parts[0] + ":" + parts[1]
		}
	}
	h, err := v1.NewHash(key)
	if err != nil {
		return v1.Hash{}, false
	}
	return h, true
}
