package registry

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"

	"HarborDownloader/internal/apperr"
	"HarborDownloader/internal/model"
)

type NormalizedRegistry struct {
	Host     string
	Scheme   string
	Insecure bool
	HTTP     bool
}

func NormalizeRegistry(raw string, insecure bool) (NormalizedRegistry, error) {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimRight(raw, "/")
	if raw == "" {
		return NormalizedRegistry{}, apperr.New(apperr.CodeInvalid, "Harbor 地址不能为空。", nil)
	}
	n := NormalizedRegistry{Insecure: insecure, Scheme: "https"}
	lower := strings.ToLower(raw)
	switch {
	case strings.HasPrefix(lower, "http://"):
		if !insecure {
			return NormalizedRegistry{}, apperr.New(apperr.CodeInvalid, "使用 HTTP 必须勾选「允许不安全 Registry」。", nil)
		}
		n.Host = raw[len("http://"):]
		n.Scheme = "http"
		n.HTTP = true
	case strings.HasPrefix(lower, "https://"):
		n.Host = raw[len("https://"):]
		n.Scheme = "https"
	default:
		if strings.Contains(raw, "://") {
			return NormalizedRegistry{}, apperr.New(apperr.CodeInvalid, "不支持的 Registry 协议。", nil)
		}
		n.Host = raw
	}
	n.Host = strings.TrimRight(n.Host, "/")
	if n.Host == "" || strings.Contains(n.Host, " ") {
		return NormalizedRegistry{}, apperr.New(apperr.CodeInvalid, "Harbor 地址无效。", nil)
	}
	if strings.Contains(n.Host, "/") {
		u, err := url.Parse(n.Scheme + "://" + n.Host)
		if err != nil || u.Host == "" {
			return NormalizedRegistry{}, apperr.New(apperr.CodeInvalid, "Harbor 地址无效。", nil)
		}
		n.Host = u.Host
	}
	return n, nil
}

func ParseImageReference(image, defaultRegistry string, insecure bool) (model.ImageReference, name.Reference, error) {
	image = strings.TrimSpace(image)
	if image == "" {
		return model.ImageReference{}, nil, apperr.New(apperr.CodeInvalid, "镜像地址不能为空。", nil)
	}
	if strings.Contains(image, "..") {
		return model.ImageReference{}, nil, apperr.New(apperr.CodeInvalid, "镜像地址包含非法路径。", nil)
	}

	img := image
	lower := strings.ToLower(img)
	switch {
	case strings.HasPrefix(lower, "https://"):
		img = img[len("https://"):]
	case strings.HasPrefix(lower, "http://"):
		img = img[len("http://"):]
	}

	reg, err := NormalizeRegistry(defaultRegistry, insecure)
	if err != nil && !strings.Contains(img, "/") {
		return model.ImageReference{}, nil, err
	}

	if !hasRegistryHost(img) {
		if defaultRegistry == "" {
			return model.ImageReference{}, nil, apperr.New(apperr.CodeInvalid, "请填写 Harbor 地址，或在镜像中包含 Registry 主机名。", nil)
		}
		host := reg.Host
		if host == "" {
			n, nerr := NormalizeRegistry(defaultRegistry, insecure)
			if nerr != nil {
				return model.ImageReference{}, nil, nerr
			}
			host = n.Host
		}
		img = strings.TrimRight(host, "/") + "/" + strings.TrimLeft(img, "/")
	}

	opts := []name.Option{name.WeakValidation}
	if insecure || (err == nil && reg.HTTP) {
		opts = append(opts, name.Insecure)
	}

	ref, err := name.ParseReference(img, opts...)
	if err != nil {
		return model.ImageReference{}, nil, apperr.New(apperr.CodeInvalid, "无法解析镜像地址。\n请使用例如 harbor.company.local/project/app-gateway:1.0.0 的格式。", err)
	}

	out := model.ImageReference{
		Registry:   ref.Context().RegistryStr(),
		Repository: ref.Context().RepositoryStr(),
		Raw:        ref.Name(),
	}
	switch t := ref.(type) {
	case name.Tag:
		out.Tag = t.TagStr()
	case name.Digest:
		out.Digest = t.DigestStr()
	default:
		out.Tag = "latest"
	}
	if out.Repository == "" {
		return model.ImageReference{}, nil, apperr.New(apperr.CodeInvalid, "缺少 Repository 名称。", nil)
	}
	return out, ref, nil
}

func hasRegistryHost(image string) bool {
	// first path segment looks like a hostname if it contains a dot, localhost, or :port
	first := image
	if i := strings.IndexByte(image, '/'); i >= 0 {
		first = image[:i]
	}
	if first == "localhost" || strings.HasPrefix(first, "localhost:") {
		return true
	}
	if strings.Contains(first, ".") || strings.Contains(first, ":") {
		return true
	}
	return false
}

func ParsePlatform(p model.Platform) (model.Platform, error) {
	if p.OS == "" && p.Architecture == "" {
		return model.DefaultPlatform(), nil
	}
	if p.OS == "" || p.Architecture == "" {
		return model.Platform{}, apperr.New(apperr.CodeInvalid, "平台必须包含 OS 和 Architecture，例如 linux/amd64。", nil)
	}
	p.OS = strings.ToLower(strings.TrimSpace(p.OS))
	p.Architecture = strings.ToLower(strings.TrimSpace(p.Architecture))
	p.Variant = strings.ToLower(strings.TrimSpace(p.Variant))
	switch p.OS + "/" + p.Architecture {
	case "linux/amd64", "linux/arm64":
		return p, nil
	default:
		return model.Platform{}, apperr.New(apperr.CodeInvalid, fmt.Sprintf("MVP 仅支持 linux/amd64 与 linux/arm64，当前为 %s/%s。", p.OS, p.Architecture), nil)
	}
}
