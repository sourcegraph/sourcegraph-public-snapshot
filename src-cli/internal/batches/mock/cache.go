package mock

import (
	"context"

	"github.com/sourcegraph/src-cli/internal/batches/docker"
)

type ImageCache struct {
	Images map[string]docker.Image
}

var _ docker.ImageCache = &ImageCache{}

func (c *ImageCache) Get(name string) docker.Image { return c.Images[name] }
func (c *ImageCache) Ensure(ctx context.Context, name string) (docker.Image, error) {
	img := c.Images[name]
	return img, img.Ensure(ctx)
}
