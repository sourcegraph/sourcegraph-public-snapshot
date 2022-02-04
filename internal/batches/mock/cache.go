package mock

import "github.com/sourcegraph/src-cli/internal/batches/docker"

type ImageCache struct {
	Images map[string]docker.Image
}

var _ docker.ImageCache = &ImageCache{}

func (c *ImageCache) Get(name string) docker.Image { return c.Images[name] }
