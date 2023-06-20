package openai

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/stretchr/testify/require"
)

func TestOpenAI(t *testing.T) {
	t.Run("errors on empty embedding string", func(t *testing.T) {
		client := NewClient(&conftypes.EmbeddingsConfig{})
		invalidTexts := []string{"a", ""} // empty string is invalid
		_, err := client.GetEmbeddingsWithRetries(context.Background(), invalidTexts, 10)
		require.ErrorContains(t, err, "empty string")
	})
}
