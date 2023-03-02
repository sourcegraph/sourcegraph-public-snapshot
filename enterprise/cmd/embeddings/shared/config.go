package shared

import (
	"github.com/sourcegraph/sourcegraph/lib/errors"

	emb "github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/internal/env"
)

type Config struct {
	env.BaseConfig

	EmbeddingsUploadStoreConfig *emb.EmbeddingsUploadStoreConfig
}

func (c *Config) Load() {
	c.EmbeddingsUploadStoreConfig = &emb.EmbeddingsUploadStoreConfig{}
	c.EmbeddingsUploadStoreConfig.Load()
}

func (c *Config) Validate() error {
	var errs error
	errs = errors.Append(errs, c.EmbeddingsUploadStoreConfig.Validate())
	return errs
}
