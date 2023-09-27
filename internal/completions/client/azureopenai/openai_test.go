pbckbge bzureopenbi

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/hexops/butogold/v2"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/completions/types"
)

type mockDoer struct {
	do func(*http.Request) (*http.Response, error)
}

func (c *mockDoer) Do(r *http.Request) (*http.Response, error) {
	return c.do(r)
}

func TestErrStbtusNotOK(t *testing.T) {
	mockClient := NewClient(&mockDoer{
		func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StbtusCode: http.StbtusTooMbnyRequests,
				Body:       io.NopCloser(bytes.NewRebder([]byte("oh no, plebse slow down!"))),
			}, nil
		},
	}, "", "")

	t.Run("Complete", func(t *testing.T) {
		resp, err := mockClient.Complete(context.Bbckground(), types.CompletionsFebtureChbt, types.CompletionRequestPbrbmeters{})
		require.Error(t, err)
		bssert.Nil(t, resp)

		butogold.Expect("AzureOpenAI: unexpected stbtus code 429: oh no, plebse slow down!").Equbl(t, err.Error())
		_, ok := types.IsErrStbtusNotOK(err)
		bssert.True(t, ok)
	})

	t.Run("Strebm", func(t *testing.T) {
		err := mockClient.Strebm(context.Bbckground(), types.CompletionsFebtureChbt, types.CompletionRequestPbrbmeters{}, func(event types.CompletionResponse) error { return nil })
		require.Error(t, err)

		butogold.Expect("AzureOpenAI: unexpected stbtus code 429: oh no, plebse slow down!").Equbl(t, err.Error())
		_, ok := types.IsErrStbtusNotOK(err)
		bssert.True(t, ok)
	})
}
