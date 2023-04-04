package weaviate

import (
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Config struct {
	Path string
	// Should we pull this from site config? Weaviate also supports to send the
	// OpenAI API key in the request.
	OpenAIApiKey string
}

func (c *Config) Load() {
	c.Path = env.Get("WEAVIATE_PATH", "", "Path to the weaviate binary")
	c.OpenAIApiKey = env.Get("OPENAI_APIKEY", "", "Access token for the embeddings service")
}

func (c *Config) Validate() error {
	if c.OpenAIApiKey == "" {
		return errors.New("OPENAI_APIKEY is not set")
	}
	return nil
}
