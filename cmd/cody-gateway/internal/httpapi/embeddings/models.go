package embeddings

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"

	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/codygateway"
)

type PrefixedModelName string

const (
	ModelNameOpenAIAda              PrefixedModelName = "openai/text-embedding-ada-002"
	ModelNameSourcegraphSTMultiQA   PrefixedModelName = "sourcegraph/st-multi-qa-mpnet-base-dot-v1"
	ModelNameSourcegraphMetadataGen PrefixedModelName = "sourcegraph/st-multi-qa-mpnet-metadata"
)

type EmbeddingsClient interface {
	ProviderName() string
	GenerateEmbeddings(context.Context, codygateway.EmbeddingsRequest) (_ *codygateway.EmbeddingsResponse, consumedTokens int, _ error)
}

type ModelFactory interface {
	ForModel(model string) (_ EmbeddingsClient, ok bool)
}

type ModelFactoryMap map[PrefixedModelName]EmbeddingsClient

func (mf ModelFactoryMap) ForModel(model string) (EmbeddingsClient, bool) {
	c, ok := mf[PrefixedModelName(model)]
	return c, ok
}

func NewListHandler(prefixedAllowedModels []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		act := actor.FromContext(r.Context())

		modelEnabled := func(model PrefixedModelName) bool {
			rl, ok := act.RateLimits[codygateway.FeatureEmbeddings]
			if !act.AccessEnabled || !ok || !rl.IsValid() {
				return false
			}
			return slices.Contains(rl.EvaluateAllowedModels(prefixedAllowedModels), string(model))
		}

		models := modelsResponse{
			// Just a hardcoded list for now.
			{
				Enabled:    modelEnabled(ModelNameOpenAIAda),
				Name:       string(ModelNameOpenAIAda),
				Dimensions: 1536,
				Deprecated: false,
			},
			{
				Enabled:    modelEnabled(ModelNameSourcegraphSTMultiQA),
				Name:       string(ModelNameSourcegraphSTMultiQA),
				Dimensions: 768,
				Deprecated: false,
			},
			{
				Enabled:    modelEnabled(ModelNameSourcegraphMetadataGen),
				Name:       string(ModelNameSourcegraphMetadataGen),
				Dimensions: 768,
				Deprecated: false,
			},
		}
		_ = json.NewEncoder(w).Encode(models)
	})
}

type model struct {
	Name       string `json:"name"`
	Dimensions int    `json:"dimensions"`
	Enabled    bool   `json:"enabled"`
	Deprecated bool   `json:"deprecated"`
}

type modelsResponse []model
