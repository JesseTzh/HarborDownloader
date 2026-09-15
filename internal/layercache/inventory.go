package layercache

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

type LayerInfo struct {
	Digest     string
	Size       int64
	Partial    bool
	Referenced bool
	Images     []LayerImage
}

type LayerImage struct {
	Image    string
	Platform string
	Digest   string
}

type ImageInfo struct {
	Image       string
	Digest      string
	Platform    string
	UpdatedAt   time.Time
	LayerCount  int
	CachedCount int
}

type Inventory struct {
	Usage  Usage
	Layers []LayerInfo
	Images []ImageInfo
}

func (c *Cache) Inventory() Inventory {
	inv := Inventory{
		Layers: []LayerInfo{},
		Images: []ImageInfo{},
	}
	if c == nil {
		return inv
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	idx, err := c.loadIndexLocked()
	if err != nil {
		idx = &indexFile{}
	}
	byDigest := map[string][]LayerImage{}
	for _, rec := range idx.Images {
		info := ImageInfo{
			Image:      rec.Image,
			Digest:     rec.Digest,
			Platform:   rec.Platform,
			UpdatedAt:  rec.UpdatedAt,
			LayerCount: len(rec.Layers),
		}
		for _, d := range rec.Layers {
			byDigest[d] = append(byDigest[d], LayerImage{
				Image:    rec.Image,
				Platform: rec.Platform,
				Digest:   rec.Digest,
			})
			if c.Has(mustHash(d), 0) {
				info.CachedCount++
			}
		}
		inv.Images = append(inv.Images, info)
	}

	u := Usage{Path: c.dir, Images: len(idx.Images)}
	if c.dir != "" {
		entries, err := os.ReadDir(c.dir)
		if err == nil {
			for _, e := range entries {
				if e.IsDir() || skipMetaFile(e.Name()) {
					continue
				}
				info, err := e.Info()
				if err != nil {
					continue
				}
				u.Files++
				u.Bytes += info.Size()
				layer := LayerInfo{Size: info.Size(), Images: []LayerImage{}}
				if strings.Contains(e.Name(), ".partial") {
					layer.Partial = true
					layer.Digest = e.Name()
					layer.Referenced = false
				} else if h, ok := hashFromFileName(e.Name()); ok {
					layer.Digest = h.String()
					layer.Images = byDigest[h.String()]
					if layer.Images == nil {
						layer.Images = []LayerImage{}
					}
					layer.Referenced = len(layer.Images) > 0
				} else {
					layer.Digest = e.Name()
				}
				if !layer.Referenced {
					u.UnreferencedFiles++
					u.UnreferencedBytes += info.Size()
				}
				inv.Layers = append(inv.Layers, layer)
			}
		}
	}
	inv.Usage = u

	sort.Slice(inv.Layers, func(i, j int) bool {
		a, b := inv.Layers[i], inv.Layers[j]
		if a.Referenced != b.Referenced {
			return !a.Referenced
		}
		if a.Size != b.Size {
			return a.Size > b.Size
		}
		return a.Digest < b.Digest
	})
	sort.Slice(inv.Images, func(i, j int) bool {
		if !inv.Images[i].UpdatedAt.Equal(inv.Images[j].UpdatedAt) {
			return inv.Images[i].UpdatedAt.After(inv.Images[j].UpdatedAt)
		}
		return inv.Images[i].Image < inv.Images[j].Image
	})
	return inv
}

func (c *Cache) DeleteLayer(digest string) error {
	if c == nil || c.dir == "" {
		return fmt.Errorf("缓存不可用")
	}
	target := strings.TrimSpace(digest)
	if target == "" || strings.ContainsAny(target, `/\`) || strings.Contains(target, "..") {
		return fmt.Errorf("无效的 layer。")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	var path string
	var hashKey string
	if h, err := v1.NewHash(target); err == nil {
		path = c.file(h)
		hashKey = h.String()
	} else if h, ok := hashFromFileName(filepath.Base(target)); ok {
		path = c.file(h)
		hashKey = h.String()
	} else if strings.Contains(target, ".partial") {
		path = filepath.Join(c.dir, filepath.Base(target))
	} else {
		return fmt.Errorf("无效的 layer。")
	}

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	if hashKey == "" {
		return nil
	}
	return c.stripLayerLocked(hashKey)
}

func (c *Cache) stripLayerLocked(digest string) error {
	idx, err := c.loadIndexLocked()
	if err != nil {
		return err
	}
	changed := false
	kept := make([]ImageRecord, 0, len(idx.Images))
	for _, rec := range idx.Images {
		filtered := make([]string, 0, len(rec.Layers))
		for _, l := range rec.Layers {
			if l == digest {
				changed = true
				continue
			}
			filtered = append(filtered, l)
		}
		if len(filtered) == 0 {
			changed = true
			continue
		}
		rec.Layers = filtered
		kept = append(kept, rec)
	}
	if !changed {
		return nil
	}
	idx.Images = kept
	return c.saveIndexLocked(idx)
}

func mustHash(s string) v1.Hash {
	h, err := v1.NewHash(s)
	if err != nil {
		return v1.Hash{}
	}
	return h
}
