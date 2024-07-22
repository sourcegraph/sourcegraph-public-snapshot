package shared

import (
	"github.com/sourcegraph/sourcegraph/lib/errors"

	emb "github.com/sourcegraph/sourcegraph/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/internal/env"
)

const defaultEmbeddingsCacheSize = "6GiB"

type Config struct {
	env.BaseConfig

	EmbeddingsUploadStoreConfig *emb.ObjectStorageConfig

	EmbeddingsCacheSize uint64
}

func (c *Config) Load() {
	c.EmbeddingsUploadStoreConfig = &emb.ObjectStorageConfig{}
	c.EmbeddingsUploadStoreConfig.Load()

	c.EmbeddingsCacheSize = env.MustGetBytes("EMBEDDINGS_CACHE_SIZE", defaultEmbeddingsCacheSize, "The size of the in-memory cache for embeddings indexes")
}

func (c *Config) Validate() error {
	var errs error
	errs = errors.Append(errs, c.BaseConfig.Validate())
	errs = errors.Append(errs, c.EmbeddingsUploadStoreConfig.Validate())
	return errs
}
