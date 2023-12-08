package usagestats

import (
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func GetCodyProviders() (*types.CodyProviders, error) {
	c := conf.SiteConfig()
	providers := types.CodyProviders{
		Completions: &types.CodyCompletionProvider{
			ChatModel:       c.Completions.ChatModel,
			CompletionModel: c.Completions.CompletionModel,
			FastChatModel:   c.Completions.FastChatModel,
			Provider:        c.Completions.Provider,
		},
		Embeddings: &types.CodyEmbeddingsProvider{
			Model:    c.Embeddings.Model,
			Provider: c.Embeddings.Provider,
		},
	}
	return &providers, nil
}
