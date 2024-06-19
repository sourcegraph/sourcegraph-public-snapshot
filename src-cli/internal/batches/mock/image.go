package mock

import (
	"context"

	"github.com/sourcegraph/src-cli/internal/batches/docker"
)

type Image struct {
	RawDigest string
	DigestErr error
	EnsureErr error
	UidGid    docker.UIDGID
	UidGidErr error
}

var _ docker.Image = &Image{}

func (image *Image) Digest(ctx context.Context) (string, error) {
	return image.RawDigest, image.DigestErr
}

func (image *Image) Ensure(ctx context.Context) error {
	return image.EnsureErr
}

func (image *Image) UIDGID(ctx context.Context) (docker.UIDGID, error) {
	return image.UidGid, image.UidGidErr
}
