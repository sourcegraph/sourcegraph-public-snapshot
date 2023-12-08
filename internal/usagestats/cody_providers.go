package usagestats

import (
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func GetCodyProviders() (*types.CodyProviders, error) {
	c := conf.SiteConfig()
	providers := types.CodyProviders{
		Completions: &types.CodyCompletionProvider{
			Provider: c.Completions.Provider,
		},
		Embeddings: &types.CodyEmbeddingsProvider{
			Provider: c.Embeddings.Provider,
		},
	}
	if c.Completions.Provider == string(conftypes.CompletionsProviderNameSourcegraph) {
		providers.Completions.ChatModel = c.Completions.ChatModel
		providers.Completions.CompletionModel = c.Completions.CompletionModel
	}
	if c.Embeddings.Provider == string(conftypes.EmbeddingsProviderNameSourcegraph) {
		providers.Embeddings.Model = c.Embeddings.Model
	}
	return &providers, nil
}
