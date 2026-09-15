package disk

import (
	"fmt"
	"os"
	"path/filepath"
)

const SafetyMargin = 512 * 1024 * 1024

func existingDir(path string) string {
	dir := path
	if st, err := os.Stat(path); err == nil && !st.IsDir() {
		dir = filepath.Dir(path)
	}
	for {
		if _, err := os.Stat(dir); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return dir
		}
		dir = parent
	}
}

func RequiredBytes(tarSize, missingCache int64, sameVolume bool) (outputNeed, cacheNeed int64) {
	if tarSize < 0 {
		tarSize = 0
	}
	if missingCache < 0 {
		missingCache = 0
	}
	if sameVolume {
		return tarSize + missingCache, 0
	}
	return tarSize, missingCache
}

func EnsureDownloadSpace(outputPath, cacheDir string, tarSize, missingCache int64) error {
	outDir := existingDir(outputPath)
	if cacheDir == "" || missingCache <= 0 {
		return EnsureFreeSpace(outDir, tarSize)
	}
	cacheExisting := existingDir(cacheDir)
	same, err := SameVolume(outDir, cacheExisting)
	outNeed, cacheNeed := RequiredBytes(tarSize, missingCache, err == nil && same)
	if err := EnsureFreeSpace(outDir, outNeed); err != nil {
		return err
	}
	if cacheNeed > 0 {
		return EnsureFreeSpace(cacheExisting, cacheNeed)
	}
	return nil
}

func EnsureFreeSpace(path string, estimated int64) error {
	dir := existingDir(path)
	avail, err := AvailableBytes(dir)
	if err != nil {
		return fmt.Errorf("无法检查磁盘空间: %w", err)
	}
	need := estimated + SafetyMargin
	if need < 0 {
		need = SafetyMargin
	}
	if uint64(need) > avail {
		return fmt.Errorf("磁盘空间不足。\n预计下载大小: %s\n需要预留: %s\n可用空间: %s",
			FormatBytes(estimated), FormatBytes(need), FormatBytes(int64(avail)))
	}
	return nil
}

func FormatBytes(n int64) string {
	if n < 0 {
		n = 0
	}
	const (
		kb = 1024
		mb = 1024 * kb
		gb = 1024 * mb
	)
	switch {
	case n >= gb:
		return fmt.Sprintf("%.2f GB", float64(n)/float64(gb))
	case n >= mb:
		return fmt.Sprintf("%.2f MB", float64(n)/float64(mb))
	case n >= kb:
		return fmt.Sprintf("%.2f KB", float64(n)/float64(kb))
	default:
		return fmt.Sprintf("%d B", n)
	}
}
