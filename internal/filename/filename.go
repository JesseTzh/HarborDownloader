package filename

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"HarborDownloader/internal/model"
)

var illegalChars = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)

var reserved = map[string]struct{}{
	"CON": {}, "PRN": {}, "AUX": {}, "NUL": {},
	"COM1": {}, "COM2": {}, "COM3": {}, "COM4": {}, "COM5": {},
	"COM6": {}, "COM7": {}, "COM8": {}, "COM9": {},
	"LPT1": {}, "LPT2": {}, "LPT3": {}, "LPT4": {}, "LPT5": {},
	"LPT6": {}, "LPT7": {}, "LPT8": {}, "LPT9": {},
}

func Sanitize(name string) string {
	name = strings.TrimSpace(name)
	name = illegalChars.ReplaceAllString(name, "-")
	name = strings.Trim(name, " .")
	if name == "" || name == "." || name == ".." {
		return "image"
	}
	base := name
	if i := strings.LastIndex(name, "."); i > 0 {
		base = name[:i]
	}
	if _, ok := reserved[strings.ToUpper(base)]; ok {
		name = "_" + name
	}
	var b strings.Builder
	for _, r := range name {
		if unicode.IsControl(r) {
			b.WriteByte('-')
			continue
		}
		b.WriteRune(r)
	}
	out := b.String()
	if len(out) > 180 {
		out = out[:180]
	}
	return out
}

func DefaultTarNameFor(ref model.ImageReference, platform model.Platform, targetTag string) (string, error) {
	if strings.TrimSpace(targetTag) == "" {
		return DefaultTarName(ref, platform), nil
	}
	tagged, err := model.ParseTargetTag(targetTag)
	if err != nil {
		return "", err
	}
	return DefaultTarName(tagged, platform), nil
}

func DefaultTarName(ref model.ImageReference, platform model.Platform) string {
	repo := ref.Repository
	if repo == "" {
		repo = "image"
	}
	base := repo
	if i := strings.LastIndex(repo, "/"); i >= 0 && i+1 < len(repo) {
		base = repo[i+1:]
	}
	tag := ref.Tag
	if tag == "" {
		if ref.Digest != "" {
			d := strings.TrimPrefix(ref.Digest, "sha256:")
			if len(d) > 12 {
				d = d[:12]
			}
			tag = d
		} else {
			tag = "latest"
		}
	}
	osName := platform.OS
	arch := platform.Architecture
	if osName == "" {
		osName = "linux"
	}
	if arch == "" {
		arch = "amd64"
	}
	plat := osName + "-" + arch
	if platform.Variant != "" {
		plat += "-" + Sanitize(platform.Variant)
	}
	return fmt.Sprintf("%s_%s_%s.tar", Sanitize(base), Sanitize(tag), Sanitize(plat))
}

func JobOutputDir(root, jobName string) (string, error) {
	name := Sanitize(jobName)
	if name == "" || name == "." || name == ".." {
		name = "job"
	}
	return ResolveOutputPath(root, name)
}

func PrepareJobOutputDir(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return ClearDownloadArtifacts(dir)
}

func ClearDownloadArtifacts(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !isDownloadArtifact(e.Name()) {
			continue
		}
		if err := os.Remove(filepath.Join(dir, e.Name())); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func isDownloadArtifact(name string) bool {
	lower := strings.ToLower(name)
	return strings.HasSuffix(lower, ".tar") ||
		strings.HasSuffix(lower, ".tar.zst") ||
		strings.HasSuffix(lower, ".downloading") ||
		strings.HasSuffix(lower, ".tar.zst.tmp")
}

func ResolveOutputPath(outputDir, fileName string) (string, error) {
	if fileName == "" {
		return "", fmt.Errorf("文件名为空")
	}
	if fileName != filepath.Base(fileName) || strings.Contains(fileName, "..") {
		return "", fmt.Errorf("非法文件名")
	}
	absDir, err := filepath.Abs(filepath.Clean(outputDir))
	if err != nil {
		return "", err
	}
	full := filepath.Join(absDir, fileName)
	rel, err := filepath.Rel(absDir, full)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("输出路径超出指定目录")
	}
	return full, nil
}
