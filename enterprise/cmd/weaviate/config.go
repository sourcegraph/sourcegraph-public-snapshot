package weaviate

import "github.com/sourcegraph/sourcegraph/internal/env"

type Config struct {
	Path string
}

func (c *Config) Load() {
	c.Path = env.Get("WEAVIATE_PATH", "asdf", "Path to the weaviate binary")
}

func (c *Config) Validate() error {
	return nil
}
