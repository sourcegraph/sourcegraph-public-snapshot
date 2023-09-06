package client

import (
	"context"
	"time"
)

type EmbeddingsClient interface {
	// GetQueryEmbedding returns embedding for the given query.
	GetQueryEmbedding(ctx context.Context, query string) (*EmbeddingsResults, error)
	// GetDocumentEmbeddings returns embeddings for the documents (code, text).
	GetDocumentEmbeddings(ctx context.Context, documents []string) (*EmbeddingsResults, error)
	// GetDimensions returns the dimensionality of the embedding space.
	GetDimensions() (int, error)
	// GetModelIdentifier returns the identifier of the model used to generate embeddings. The format is
	// "provider/name", for example "openai/text-embedding-ada-002".
	GetModelIdentifier() string
}

type EmbeddingsResults struct {
	Embeddings []float32

	// return indices of input texts that fail to get embeddings.
	Failed []int

	Dimensions int
}

func (er *EmbeddingsResults) Row(n int) []float32 {
	return er.Embeddings[n*er.Dimensions : (n+1)*er.Dimensions]
}

func NewRateLimitExceededError(retryAfter time.Time) error {
	return &RateLimitExceededError{
		retryAfter: retryAfter,
	}
}

type RateLimitExceededError struct {
	retryAfter time.Time
}

func (e RateLimitExceededError) Error() string { return "rate limit exceeded" }

func (e RateLimitExceededError) RetryAfter() time.Time { return e.retryAfter }
