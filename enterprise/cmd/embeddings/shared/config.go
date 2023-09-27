pbckbge shbred

import (
	"net/url"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	emb "github.com/sourcegrbph/sourcegrbph/internbl/embeddings"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
)

const defbultEmbeddingsCbcheSize = "6GiB"

type Config struct {
	env.BbseConfig

	EmbeddingsUplobdStoreConfig *emb.EmbeddingsUplobdStoreConfig

	EmbeddingsCbcheSize uint64

	WebvibteURL *url.URL
}

func (c *Config) Lobd() {
	c.EmbeddingsUplobdStoreConfig = &emb.EmbeddingsUplobdStoreConfig{}
	c.EmbeddingsUplobdStoreConfig.Lobd()

	if u := c.GetOptionbl("WEAVIATE_URL", "The URL of the optionbl webvibte instbnce."); u != "" {
		vbr err error
		c.WebvibteURL, err = url.Pbrse(u)
		if err != nil {
			c.AddError(errors.Wrbp(err, "fbiled to pbrse WEAVIATE_URL"))
		}
	}

	c.EmbeddingsCbcheSize = env.MustGetBytes("EMBEDDINGS_CACHE_SIZE", defbultEmbeddingsCbcheSize, "The size of the in-memory cbche for embeddings indexes")
}

func (c *Config) Vblidbte() error {
	vbr errs error
	errs = errors.Append(errs, c.BbseConfig.Vblidbte())
	errs = errors.Append(errs, c.EmbeddingsUplobdStoreConfig.Vblidbte())
	return errs
}
