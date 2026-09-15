package layercache

import (
	"os"
	"path/filepath"
	"strings"
)

type Usage struct {
	Path              string
	Files             int
	Bytes             int64
	Images            int
	UnreferencedFiles int
	UnreferencedBytes int64
}

func (c *Cache) Usage() Usage {
	u := Usage{}
	if c == nil {
		return u
	}
	u.Path = c.dir
	refs, images, err := c.referencedLockedSafe()
	if err == nil {
		u.Images = images
	}
	if c.dir == "" {
		return u
	}
	entries, err := os.ReadDir(c.dir)
	if err != nil {
		return u
	}
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
		if isUnreferenced(e.Name(), refs) {
			u.UnreferencedFiles++
			u.UnreferencedBytes += info.Size()
		}
	}
	return u
}

func (c *Cache) referencedLockedSafe() (map[string]struct{}, int, error) {
	if c == nil {
		return map[string]struct{}{}, 0, nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.referencedLocked()
}

func skipMetaFile(name string) bool {
	return name == "index.json" || strings.Contains(name, ".json.tmp")
}

func isUnreferenced(name string, refs map[string]struct{}) bool {
	if strings.Contains(name, ".partial") {
		return true
	}
	h, ok := hashFromFileName(name)
	if !ok {
		return false
	}
	if refs == nil {
		return true
	}
	_, referenced := refs[h.String()]
	return !referenced
}

func (c *Cache) ClearUnreferenced() (Usage, error) {
	removed := Usage{}
	if c == nil || c.dir == "" {
		return removed, nil
	}
	removed.Path = c.dir
	c.mu.Lock()
	defer c.mu.Unlock()
	refs, images, err := c.referencedLocked()
	if err != nil {
		return removed, err
	}
	removed.Images = images
	entries, err := os.ReadDir(c.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return removed, nil
		}
		return removed, err
	}
	var first error
	for _, e := range entries {
		if e.IsDir() || skipMetaFile(e.Name()) {
			continue
		}
		if !isUnreferenced(e.Name(), refs) {
			continue
		}
		path := filepath.Join(c.dir, e.Name())
		info, statErr := e.Info()
		if err := os.Remove(path); err != nil {
			if first == nil {
				first = err
			}
			continue
		}
		removed.Files++
		removed.UnreferencedFiles++
		if statErr == nil {
			removed.Bytes += info.Size()
			removed.UnreferencedBytes += info.Size()
		}
	}
	return removed, first
}

func (c *Cache) Clear() (Usage, error) {
	removed := Usage{}
	if c == nil || c.dir == "" {
		return removed, nil
	}
	removed.Path = c.dir
	c.mu.Lock()
	defer c.mu.Unlock()
	entries, err := os.ReadDir(c.dir)
	if err != nil && !os.IsNotExist(err) {
		return removed, err
	}
	var first error
	if err == nil {
		for _, e := range entries {
			if e.IsDir() || skipMetaFile(e.Name()) {
				continue
			}
			path := filepath.Join(c.dir, e.Name())
			info, statErr := e.Info()
			if err := os.Remove(path); err != nil {
				if first == nil {
					first = err
				}
				continue
			}
			removed.Files++
			if statErr == nil {
				removed.Bytes += info.Size()
			}
		}
	}
	if err := c.removeIndexLocked(); err != nil && first == nil {
		first = err
	}
	return removed, first
}
