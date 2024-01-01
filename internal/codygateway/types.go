package codygateway

import (
	"time"

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
	FeatureAttribution     Feature = "attribution"
)

func (f Feature) IsValid() bool {
	switch f {
	case FeatureCodeCompletions,
		FeatureChatCompletions,
		FeatureEmbeddings,
		FeatureAttribution:
		return true
	}
	return false
}

var featureDisplayNames = map[Feature]string{
	FeatureChatCompletions: "Chat",
	FeatureCodeCompletions: "Autocomplete",
	FeatureEmbeddings:      "Embeddings",
	FeatureAttribution:     "Attribution",
}

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

// ActorConcurrencyLimitConfig is the configuration for the concurrent requests
// limit of an actor.
type ActorConcurrencyLimitConfig struct {
	// Percentage is the percentage of the daily rate limit to be used to compute the
	// concurrency limit.
	Percentage float32
	// Interval is the time interval of the limit bucket.
	Interval time.Duration
}

// ActorRateLimitNotifyConfig is the configuration for the rate limit
// notifications of an actor.
type ActorRateLimitNotifyConfig struct {
	// SlackWebhookURL is the URL of the Slack webhook to send the alerts to.
	SlackWebhookURL string
}
