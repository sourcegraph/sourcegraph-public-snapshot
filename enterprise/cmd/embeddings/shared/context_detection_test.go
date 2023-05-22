package shared

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
)

func TestIsContextRequiredForChatQuery(t *testing.T) {
	getContextDetectionEmbeddingIndex := func(ctx context.Context) (*embeddings.ContextDetectionEmbeddingIndex, error) {
		return &embeddings.ContextDetectionEmbeddingIndex{
			MessagesWithAdditionalContextMeanEmbedding:    []float32{0.0, 1.0},
			MessagesWithoutAdditionalContextMeanEmbedding: []float32{1.0, 0.0},
		}, nil
	}

	cases := []struct {
		name         string
		query        string
		embedding    []float32
		embeddingErr error
		want         bool
	}{
		{
			name:      "query matches no context regex",
			query:     "that answer looks incorrect",
			embedding: []float32{0.0, 1.0}, // unused
			want:      false,
		},
		{
			name:      "query matches context regex",
			query:     "where is the cody plugin code?",
			embedding: []float32{1.0, 0.0}, // unused
			want:      true,
		},
		{
			name:      "query that matches context regex 2",
			query:     "is the zoekt package used in my repo",
			embedding: []float32{1.0, 0.0}, // unused
			want:      true,
		},
		{
			name:      "query matches context regex 3",
			query:     "what directory contains the cody plugin",
			embedding: []float32{1.0, 0.0}, // unused
			want:      true,
		},
		{
			name:      "query similar to no context embeddings",
			query:     "hello, testing this works!",
			embedding: []float32{0.9, 0.1},
			want:      false,
		},
		{
			name:      "query not similar enough to no context embeddings",
			query:     "hello, testing this works!",
			embedding: []float32{0.5, 0.5},
			want:      true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			getQueryEmbedding := func(_ context.Context, query string) ([]float32, error) {
				return tt.embedding, tt.embeddingErr
			}

			got, err := isContextRequiredForChatQuery(context.Background(),
				getQueryEmbedding,
				getContextDetectionEmbeddingIndex,
				tt.query)

			if err != nil {
				t.Fatal(err)
			}

			if got != tt.want {
				t.Fatalf("expected context required to be %t but was %t", tt.want, got)
			}
		})
	}
}
