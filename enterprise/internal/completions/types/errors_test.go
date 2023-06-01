package types

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrStatusNotOK(t *testing.T) {
	rec := httptest.NewRecorder()
	rec.WriteHeader(http.StatusTooManyRequests)
	rec.Header().Set("X-RateLimit-Limit", "100")
	rec.Write([]byte("oh no, please slow down!"))
	resp := rec.Result()

	var err error
	err = NewErrStatusNotOK(t.Name(), resp)

	// Check that it's safe to close the body multiple times, since callsites
	// might have a defer resp.Body.Close() and also call NewErrStatusNotOK.
	assert.NoError(t, resp.Body.Close())

	t.Run("NewErrStatusNotOK", func(t *testing.T) {
		assert.Error(t, err)
		autogold.Expect("TestErrStatusNotOK: unexpected status code 429: oh no, please slow down!").Equal(t, err.Error())
	})

	t.Run("IsErrStatusNotOK", func(t *testing.T) {
		errNotOK, ok := IsErrStatusNotOK(err)
		require.True(t, ok)
		assert.Equal(t, resp.StatusCode, errNotOK.statusCode)
		assert.Equal(t, resp.Header, errNotOK.responseHeader)
		assert.Equal(t, "oh no, please slow down!", errNotOK.responseBody)

		t.Run("WriteResponseHeaders", func(t *testing.T) {
			rec := httptest.NewRecorder()
			errNotOK.WriteResponseHeaders(rec)

			// Should have written status code and headers.
			writtenResp := rec.Result()
			assert.Equal(t, resp.StatusCode, writtenResp.StatusCode)
			assert.Equal(t, resp.Header, writtenResp.Header)

			// Should not have written the response body.
			writtenBody, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.Empty(t, writtenBody)
		})
	})
}
