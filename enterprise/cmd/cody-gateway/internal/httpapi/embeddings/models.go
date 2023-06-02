package embeddings

import (
	"context"
	"encoding/json"
	"net/http"

	"golang.org/x/exp/slices"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/actor"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codygateway"
)

type EmbeddingsClient interface {
	GenerateEmbeddings(context.Context, codygateway.EmbeddingsRequest) (_ *codygateway.EmbeddingsResponse, consumedTokens int, _ error)
}

type ModelFactory interface {
	ForModel(model string) (_ EmbeddingsClient, ok bool)
}

type ModelFactoryMap map[string]EmbeddingsClient

func (mf ModelFactoryMap) ForModel(model string) (EmbeddingsClient, bool) {
	c, ok := mf[model]
	return c, ok
}

// var Models = map[string]EmbeddingsClient{
// 	"openai/text-embedding-ada-002": nil,
// }

type openAIModel struct {
	upstreamName string
	dimensions   int
}

var openAIModelMappings = map[string]openAIModel{
	"openai/text-embedding-ada-002": {
		upstreamName: "text-embedding-ada-002",
		dimensions:   1536,
	},
}

func NewListHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		act := actor.FromContext(r.Context())

		modelEnabled := func(model string) bool {
			if !act.AccessEnabled || !act.EmbeddingsRateLimit.IsValid() {
				return false
			}
			return slices.Contains(act.EmbeddingsRateLimit.AllowedModels, model)
		}

		models := modelsResponse{
			{
				Enabled:    modelEnabled("openai/text-embedding-ada-002"),
				Name:       "openai/text-embedding-ada-002",
				Dimensions: 1536,
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
