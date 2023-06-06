package client

import (
	"context"
	"time"
)

type EmbeddingsClient interface {
	// GetEmbeddings returns embeddings for the given texts.
	GetEmbeddings(ctx context.Context, texts []string) ([]float32, error)
	// GetDimensions returns the dimensionality of the embedding space.
	GetDimensions() (int, error)
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
