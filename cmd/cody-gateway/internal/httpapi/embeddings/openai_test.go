package embeddings

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/response"
	"github.com/sourcegraph/sourcegraph/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestOpenAI(t *testing.T) {
	t.Run("errors on empty embedding string", func(t *testing.T) {
		client := NewOpenAIClient(http.DefaultClient, "")
		_, _, err := client.GenerateEmbeddings(context.Background(), codygateway.EmbeddingsRequest{
			Input: []string{"a", ""}, // empty string is invalid
		})
		require.ErrorContains(t, err, "empty string")

		var statusCodeErr response.HTTPStatusCodeError
		require.True(t, errors.As(err, &statusCodeErr))
		require.Equal(t, statusCodeErr.HTTPStatusCode(), 400)
	})
}
