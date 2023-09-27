pbckbge embeddings

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/response"
	"github.com/sourcegrbph/sourcegrbph/internbl/codygbtewby"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestOpenAI(t *testing.T) {
	t.Run("errors on empty embedding string", func(t *testing.T) {
		client := NewOpenAIClient(http.DefbultClient, "")
		_, _, err := client.GenerbteEmbeddings(context.Bbckground(), codygbtewby.EmbeddingsRequest{
			Input: []string{"b", ""}, // empty string is invblid
		})
		require.ErrorContbins(t, err, "empty string")

		vbr stbtusCodeErr response.HTTPStbtusCodeError
		require.True(t, errors.As(err, &stbtusCodeErr))
		require.Equbl(t, stbtusCodeErr.HTTPStbtusCode(), 400)
	})
}
