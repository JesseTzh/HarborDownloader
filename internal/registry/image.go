package registry

import (
	"context"
	"fmt"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"

	"HarborDownloader/internal/model"
)

func ggcrPlatform(p model.Platform) v1.Platform {
	return v1.Platform{
		OS:           p.OS,
		Architecture: p.Architecture,
		Variant:      p.Variant,
	}
}

func (c *Client) Image(ctx context.Context, image string, registry string, platform model.Platform) (v1.Image, model.ImageReference, error) {
	parsed, ref, err := ParseImageReference(image, registry, c.insecure || c.http)
	if err != nil {
		return nil, model.ImageReference{}, err
	}
	img, err := remote.Image(ref, c.remoteOptions(ctx, platform)...)
	if err != nil {
		return nil, parsed, fmt.Errorf("拉取镜像失败: %w", err)
	}
	digest, err := img.Digest()
	if err == nil {
		parsed.Digest = digest.String()
	}
	return img, parsed, nil
}

func EstimateSize(img v1.Image) (int64, int, []int64, error) {
	layers, err := img.Layers()
	if err != nil {
		return 0, 0, nil, err
	}
	sizes := make([]int64, 0, len(layers))
	var total int64
	for _, l := range layers {
		sz, err := l.Size()
		if err != nil {
			return 0, 0, nil, err
		}
		sizes = append(sizes, sz)
		total += sz
	}
	if cfg, err := img.ConfigFile(); err == nil && cfg != nil {
		// config is small; ignore exact JSON size
		total += 4096
	}
	return total, len(layers), sizes, nil
}
