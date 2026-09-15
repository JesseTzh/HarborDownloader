package layercache

import (
	"fmt"
	"os"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

type Stats struct {
	TotalLayers   int
	CachedLayers  int
	MissingLayers int
	TotalBytes    int64
	CachedBytes   int64
	MissingBytes  int64
}

func Inspect(img v1.Image, c *Cache) (Stats, error) {
	var st Stats
	if img == nil {
		return st, nil
	}
	layers, err := img.Layers()
	if err != nil {
		return st, err
	}
	st.TotalLayers = len(layers)
	for _, l := range layers {
		sz, err := l.Size()
		if err != nil {
			return st, err
		}
		st.TotalBytes += sz
		digest, err := l.Digest()
		if err != nil {
			return st, err
		}
		if c != nil && c.Has(digest, sz) {
			st.CachedLayers++
			st.CachedBytes += sz
			continue
		}
		if c != nil {
			info, err := os.Stat(c.file(digest))
			if err == nil && info.Size() > 0 && sz > 0 && info.Size() != sz {
				_ = c.Delete(digest)
			}
		}
		st.MissingLayers++
		st.MissingBytes += sz
	}
	return st, nil
}

func (s Stats) Message() string {
	if s.TotalLayers == 0 || s.CachedLayers == 0 {
		return "正在下载并导出 TAR"
	}
	if s.MissingLayers == 0 {
		return fmt.Sprintf("全部 %d 层已在缓存，正在导出 TAR", s.TotalLayers)
	}
	return fmt.Sprintf("正在下载缺失层并导出 TAR（缓存 %d/%d 层）", s.CachedLayers, s.TotalLayers)
}
