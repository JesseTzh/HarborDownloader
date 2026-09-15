//go:build unix

package disk

import (
	"strconv"

	"golang.org/x/sys/unix"
)

func volumeKey(path string) (string, error) {
	var st unix.Stat_t
	if err := unix.Stat(path, &st); err != nil {
		return "", err
	}
	return strconv.FormatUint(uint64(st.Dev), 10), nil
}
