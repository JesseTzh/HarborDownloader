package export

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/klauspost/compress/zstd"

	"HarborDownloader/internal/disk"
)

const PackedName = "all.tar.zst"

type PackProgress func(current, total int64)

func PackTars(ctx context.Context, dir string, progress PackProgress) (string, int64, error) {
	if err := ctx.Err(); err != nil {
		return "", 0, err
	}
	files, total, err := listTarFiles(dir)
	if err != nil {
		return "", 0, err
	}
	if len(files) == 0 {
		return "", 0, fmt.Errorf("目录中没有可打包的 TAR 文件。")
	}
	if err := disk.EnsureFreeSpace(dir, total); err != nil {
		return "", 0, err
	}

	dest := filepath.Join(dir, PackedName)
	tmp := dest + ".tmp"
	RemoveIfExists(tmp)
	defer RemoveIfExists(tmp)

	out, err := os.Create(tmp)
	if err != nil {
		return "", 0, err
	}
	enc, err := zstd.NewWriter(out, zstd.WithEncoderLevel(zstd.EncoderLevelFromZstd(19)))
	if err != nil {
		_ = out.Close()
		return "", 0, err
	}
	tw := tar.NewWriter(enc)

	closeAll := func() error {
		var first error
		if err := tw.Close(); err != nil && first == nil {
			first = err
		}
		if err := enc.Close(); err != nil && first == nil {
			first = err
		}
		if err := out.Close(); err != nil && first == nil {
			first = err
		}
		return first
	}

	var copied int64
	lastEmit := time.Time{}
	emit := func(force bool) {
		if progress == nil {
			return
		}
		now := time.Now()
		if !force && !lastEmit.IsZero() && now.Sub(lastEmit) < 200*time.Millisecond {
			return
		}
		progress(copied, total)
		lastEmit = now
	}
	emit(true)

	for _, path := range files {
		if err := ctx.Err(); err != nil {
			_ = closeAll()
			return "", 0, err
		}
		st, err := os.Stat(path)
		if err != nil {
			_ = closeAll()
			return "", 0, err
		}
		hdr, err := tar.FileInfoHeader(st, "")
		if err != nil {
			_ = closeAll()
			return "", 0, err
		}
		hdr.Name = filepath.Base(path)
		hdr.Format = tar.FormatPAX
		if err := tw.WriteHeader(hdr); err != nil {
			_ = closeAll()
			return "", 0, err
		}
		src, err := os.Open(path)
		if err != nil {
			_ = closeAll()
			return "", 0, err
		}
		n, err := copyCtx(ctx, tw, src, func(n int64) {
			copied += n
			emit(false)
		})
		_ = src.Close()
		if err != nil {
			_ = closeAll()
			return "", 0, err
		}
		if n != st.Size() {
			_ = closeAll()
			return "", 0, fmt.Errorf("打包 %s 不完整", hdr.Name)
		}
	}
	if err := closeAll(); err != nil {
		return "", 0, err
	}
	if err := ctx.Err(); err != nil {
		return "", 0, err
	}
	if err := os.Rename(tmp, dest); err != nil {
		return "", 0, err
	}
	st, err := os.Stat(dest)
	if err != nil {
		return "", 0, err
	}
	if st.Size() <= 0 {
		RemoveIfExists(dest)
		return "", 0, fmt.Errorf("打包校验失败：文件不存在或大小为 0。")
	}
	emit(true)
	return dest, st.Size(), nil
}

func listTarFiles(dir string) ([]string, int64, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, 0, err
	}
	var files []string
	var total int64
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		if !strings.HasSuffix(strings.ToLower(name), ".tar") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			return nil, 0, err
		}
		if !info.Mode().IsRegular() {
			continue
		}
		files = append(files, filepath.Join(dir, name))
		total += info.Size()
	}
	sort.Strings(files)
	return files, total, nil
}

func copyCtx(ctx context.Context, dst io.Writer, src io.Reader, onBytes func(int64)) (int64, error) {
	buf := make([]byte, 32*1024)
	var written int64
	for {
		if err := ctx.Err(); err != nil {
			return written, err
		}
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[:nr])
			if nw > 0 {
				written += int64(nw)
				if onBytes != nil {
					onBytes(int64(nw))
				}
			}
			if ew != nil {
				return written, ew
			}
			if nr != nw {
				return written, io.ErrShortWrite
			}
		}
		if er != nil {
			if er == io.EOF {
				return written, nil
			}
			return written, er
		}
	}
}
