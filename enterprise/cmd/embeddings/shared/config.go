package shared

import (
	"net/url"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	emb "github.com/sourcegraph/sourcegraph/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/internal/env"
)

const defaultEmbeddingsCacheSize = "6GiB"

type Config struct {
	env.BaseConfig

	EmbeddingsUploadStoreConfig *emb.EmbeddingsUploadStoreConfig

	EmbeddingsCacheSize uint64

	WeaviateURL *url.URL
}

func (c *Config) Load() {
	c.EmbeddingsUploadStoreConfig = &emb.EmbeddingsUploadStoreConfig{}
	c.EmbeddingsUploadStoreConfig.Load()

	if u := c.GetOptional("WEAVIATE_URL", "The URL of the optional weaviate instance."); u != "" {
		var err error
		c.WeaviateURL, err = url.Parse(u)
		if err != nil {
			c.AddError(errors.Wrap(err, "failed to parse WEAVIATE_URL"))
		}
	}

	c.EmbeddingsCacheSize = env.MustGetBytes("EMBEDDINGS_CACHE_SIZE", defaultEmbeddingsCacheSize, "The size of the in-memory cache for embeddings indexes")
}

func (c *Config) Validate() error {
	var errs error
	errs = errors.Append(errs, c.BaseConfig.Validate())
	errs = errors.Append(errs, c.EmbeddingsUploadStoreConfig.Validate())
	return errs
}
