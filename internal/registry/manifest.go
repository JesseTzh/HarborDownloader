package registry

import (
	"context"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

func (c *Client) Digest(ctx context.Context, img v1.Image) (string, error) {
	_ = ctx
	d, err := img.Digest()
	if err != nil {
		return "", err
	}
	return d.String(), nil
}
