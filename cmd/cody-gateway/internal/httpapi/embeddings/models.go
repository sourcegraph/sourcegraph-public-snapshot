package embeddings

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"

	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/codygateway"
)

type ModelName string

const (
	ModelNameOpenAIAda            ModelName = "openai/text-embedding-ada-002"
	ModelNameSourcegraphSTMultiQA ModelName = "sourcegraph/st-multi-qa-mpnet-base-dot-v1"
)

type EmbeddingsClient interface {
	ProviderName() string
	GenerateEmbeddings(context.Context, codygateway.EmbeddingsRequest) (_ *codygateway.EmbeddingsResponse, consumedTokens int, _ error)
}

type ModelFactory interface {
	ForModel(model string) (_ EmbeddingsClient, ok bool)
}

type ModelFactoryMap map[ModelName]EmbeddingsClient

func (mf ModelFactoryMap) ForModel(model string) (EmbeddingsClient, bool) {
	c, ok := mf[ModelName(model)]
	return c, ok
}

func NewListHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		act := actor.FromContext(r.Context())

		modelEnabled := func(model ModelName) bool {
			rl, ok := act.RateLimits[codygateway.FeatureEmbeddings]
			if !act.AccessEnabled || !ok || !rl.IsValid() {
				return false
			}
			return slices.Contains(rl.AllowedModels, string(model))
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
				Dimensions: 756,
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
