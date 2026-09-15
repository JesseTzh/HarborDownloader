package model

import (
	"strings"

	"github.com/google/go-containerregistry/pkg/name"

	"HarborDownloader/internal/apperr"
)

func ParseTargetTag(raw string) (ImageReference, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ImageReference{}, nil
	}
	if strings.Contains(raw, "..") {
		return ImageReference{}, apperr.New(apperr.CodeInvalid, "保存为的镜像名包含非法路径。", nil)
	}
	if strings.Contains(raw, "@") {
		return ImageReference{}, apperr.New(apperr.CodeInvalid, "保存为的镜像名不能使用 digest，请使用 name:tag。", nil)
	}
	lower := strings.ToLower(raw)
	switch {
	case strings.HasPrefix(lower, "https://"):
		raw = raw[len("https://"):]
	case strings.HasPrefix(lower, "http://"):
		raw = raw[len("http://"):]
	}
	if raw == "" {
		return ImageReference{}, apperr.New(apperr.CodeInvalid, "保存为的镜像名不能为空。", nil)
	}

	tag, err := name.NewTag(raw, name.WeakValidation)
	if err != nil {
		return ImageReference{}, apperr.New(apperr.CodeInvalid, "无法解析保存为的镜像名。\n请使用例如 app-gateway:1.0.0 或 company/app:prod 的格式。", err)
	}
	return ImageReference{
		Registry:   tag.RegistryStr(),
		Repository: tag.RepositoryStr(),
		Tag:        tag.TagStr(),
		Raw:        tag.String(),
	}, nil
}
