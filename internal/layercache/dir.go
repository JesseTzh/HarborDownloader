package layercache

import (
	"os"
	"path/filepath"
)

func Dir() (string, error) {
	if d := os.Getenv("LOCALAPPDATA"); d != "" {
		return filepath.Join(d, "HarborDownloader", "cache", "layers"), nil
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		home, herr := os.UserHomeDir()
		if herr != nil {
			return "", err
		}
		return filepath.Join(home, ".harbor-downloader", "cache", "layers"), nil
	}
	return filepath.Join(cache, "HarborDownloader", "layers"), nil
}

func Open() (*Cache, error) {
	dir, err := Dir()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	c := New(dir)
	c.cleanupPartials()
	return c, nil
}
