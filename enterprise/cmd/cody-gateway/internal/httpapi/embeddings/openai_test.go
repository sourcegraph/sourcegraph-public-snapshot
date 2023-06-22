package embeddings

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/codygateway"
	"github.com/stretchr/testify/require"
)

func TestOpenAI(t *testing.T) {
	t.Run("errors on empty embedding string", func(t *testing.T) {
		client := NewOpenAIClient("")
		_, _, err := client.GenerateEmbeddings(context.Background(), codygateway.EmbeddingsRequest{
			Input: []string{"a", ""}, // empty string is invalid
		})
		require.ErrorContains(t, err, "empty string")
	})
}
