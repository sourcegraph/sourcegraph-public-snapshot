package codygateway

import (
	"github.com/sourcegraph/sourcegraph/internal/completions/types"
)

type Feature string

var AllFeatures = []Feature{
	FeatureCodeCompletions,
	FeatureChatCompletions,
	FeatureEmbeddings,
}

// NOTE: When you add a new feature here, make sure to add it to the slice above as well.
const (
	FeatureCodeCompletions         = Feature(types.CompletionsFeatureCode)
	FeatureChatCompletions         = Feature(types.CompletionsFeatureChat)
	FeatureEmbeddings      Feature = "embeddings"
)

func (f Feature) IsValid() bool {
	switch f {
	case FeatureCodeCompletions,
		FeatureChatCompletions,
		FeatureEmbeddings:
		return true
	}
	return false
}

var featureDisplayNames map[Feature]string = map[Feature]string{FeatureChatCompletions: "Chat", FeatureCodeCompletions: "Autocomplete", FeatureEmbeddings: "Embeddings"}

func (f Feature) DisplayName() string {
	display, ok := featureDisplayNames[f]
	if !ok {
		return string(f)
	}
	return display
}

type EmbeddingsRequest struct {
	// Model is the name of the embeddings model to use.
	Model string `json:"model"`
	// Input is the list of strings to generate embeddings for.
	Input []string `json:"input"`
	// IsQuery is true if the request is used for querying, false if it used for indexing.
	// TODO: Refactor this to use an enum to be more descriptive. This will require updating callers in bfg/embeddings.
	IsQuery bool `json:"isQuery"`
}

type Embedding struct {
	// Index is the index of the input string this embedding corresponds to.
	Index int `json:"index"`
	// Data is the embedding vector for the input string.
	Data []float32 `json:"data"`
}

type EmbeddingsResponse struct {
	// Embeddings is a list of generated embeddings, one for each input string.
	Embeddings []Embedding `json:"embeddings"`
	// Model is the name of the model used to generate the embeddings.
	Model string `json:"model"`
	// ModelDimensions is the dimensionality of the embeddings model used.
	ModelDimensions int `json:"dimensions"`
}

// AttributionRequest is request for attribution search.
// Expected in JSON form as the body of POST request.
type AttributionRequest struct {
	// Snippet is the text to search attribution of.
	Snippet string `json:"snippet"`
	// Limit is the upper bound of number of responses we want to get.
	Limit int `json:"limit"`
}

// AttributionResponse is response of attribution search.
// Contains some repositories to which the snippet can be attributed to.
type AttributionResponse struct {
	// Repositories which contain code matching search snippet.
	Repositories []AttributionRepository
	// TotalCount denotes how many total matches there were (including listed repositories).
	TotalCount int `json:"totalCount,omitempty"`
	// LimitHit is true if the number of search hits goes beyond limit specified in request.
	LimitHit bool `json:"limitHit,omitempty"`
}

// AttributionRepository represents matching of search content against a repository.
type AttributionRepository struct {
	// Name of the repo on dotcom. Like github.com/sourcegraph/sourcegraph.
	Name string `json:"name"`
}
