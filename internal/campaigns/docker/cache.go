package docker

import "sync"

// ImageCache is a cache of metadata about Docker images, indexed by name.
type ImageCache struct {
	images   map[string]Image
	imagesMu sync.Mutex
}

// NewImageCache creates a new image cache.
func NewImageCache() *ImageCache {
	return &ImageCache{
		images: make(map[string]Image),
	}
}

// Get returns the image cache entry for the given Docker image. The name may be
// anything the Docker command line will accept as an image name: this will
// generally be IMAGE or IMAGE:TAG.
func (ic *ImageCache) Get(name string) Image {
	ic.imagesMu.Lock()
	defer ic.imagesMu.Unlock()

	if image, ok := ic.images[name]; ok {
		return image
	}

	image := &image{name: name}
	ic.images[name] = image
	return image
}
