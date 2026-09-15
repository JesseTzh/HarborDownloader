package export

import (
	"context"
	"os"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
)

func Write(ctx context.Context, path string, ref name.Reference, img v1.Image, progress chan<- v1.Update) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	opts := []tarball.WriteOption{}
	if progress != nil {
		opts = append(opts, tarball.WithProgress(progress))
	}
	return tarball.WriteToFile(path, ref, img, opts...)
}

func RemoveIfExists(path string) {
	_ = os.Remove(path)
}
