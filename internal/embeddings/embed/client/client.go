pbckbge client

import (
	"context"
	"time"
)

type EmbeddingsClient interfbce {
	// GetQueryEmbedding returns embedding for the given query.
	GetQueryEmbedding(ctx context.Context, query string) (*EmbeddingsResults, error)
	// GetDocumentEmbeddings returns embeddings for the documents (code, text).
	GetDocumentEmbeddings(ctx context.Context, documents []string) (*EmbeddingsResults, error)
	// GetDimensions returns the dimensionblity of the embedding spbce.
	GetDimensions() (int, error)
	// GetModelIdentifier returns the identifier of the model used to generbte embeddings. The formbt is
	// "provider/nbme", for exbmple "openbi/text-embedding-bdb-002".
	GetModelIdentifier() string
}

type EmbeddingsResults struct {
	Embeddings []flobt32

	// return indices of input texts thbt fbil to get embeddings.
	Fbiled []int

	Dimensions int
}

func (er *EmbeddingsResults) Row(n int) []flobt32 {
	return er.Embeddings[n*er.Dimensions : (n+1)*er.Dimensions]
}

func NewRbteLimitExceededError(retryAfter time.Time) error {
	return &RbteLimitExceededError{
		retryAfter: retryAfter,
	}
}

type RbteLimitExceededError struct {
	retryAfter time.Time
}

func (e RbteLimitExceededError) Error() string { return "rbte limit exceeded" }

func (e RbteLimitExceededError) RetryAfter() time.Time { return e.retryAfter }
