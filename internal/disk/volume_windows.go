//go:build windows

package disk

import (
	"path/filepath"
	"strings"
)

func volumeKey(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return strings.ToLower(filepath.VolumeName(abs)), nil
}
