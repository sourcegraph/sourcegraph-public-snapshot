package embeddings

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/response"
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

	t.Run("retry on empty embedding", func(t *testing.T) {
		gotRequest1 := false
		gotRequest2 := false
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			// On the first request, respond with a null embedding
			if !gotRequest1 {
				resp := openaiEmbeddingsResponse{
					Data: []openaiEmbeddingsData{{
						Index:     0,
						Embedding: append(make([]float32, 1535), 1),
					}, {
						Index:     1,
						Embedding: nil,
					}},
				}
				json.NewEncoder(w).Encode(resp)
				gotRequest1 = true
				return
			}

			// The client should try that embedding once more. This time, actually return a value.
			if !gotRequest2 {
				resp := openaiEmbeddingsResponse{
					Data: []openaiEmbeddingsData{{
						Index:     0,
						Embedding: append(make([]float32, 1535), 2),
					}},
				}
				json.NewEncoder(w).Encode(resp)
				gotRequest2 = true
				return
			}

			panic("only expected 2 responses")
		}))
		defer s.Close()

		httpClient := s.Client()
		oldTransport := httpClient.Transport
		httpClient.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
			r.URL, _ = url.Parse(s.URL)
			return oldTransport.RoundTrip(r)
		})

		client := NewOpenAIClient(s.Client(), "")
		resp, _, err := client.GenerateEmbeddings(context.Background(), codygateway.EmbeddingsRequest{
			Model: string(ModelNameOpenAIAda),
			Input: []string{"a", "b"},
		})
		require.NoError(t, err)
		require.Equal(t, &codygateway.EmbeddingsResponse{
			ModelDimensions: 1536,
			Embeddings: []codygateway.Embedding{{
				Index: 0,
				Data:  append(make([]float32, 1535), 1),
			}, {
				Index: 1,
				Data:  append(make([]float32, 1535), 2),
			}},
		}, resp)
		require.True(t, gotRequest1)
		require.True(t, gotRequest2)
	})

	t.Run("retry on empty embedding fails", func(t *testing.T) {
		gotRequest1 := false
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			// On the first request, respond with a null embedding
			if !gotRequest1 {
				resp := openaiEmbeddingsResponse{
					Data: []openaiEmbeddingsData{{
						Index:     0,
						Embedding: append(make([]float32, 1535), 1),
					}, {
						Index:     1,
						Embedding: nil,
					}},
				}
				json.NewEncoder(w).Encode(resp)
				gotRequest1 = true
				return
			}

			// After the first request, always respond with an invalid response
			resp := openaiEmbeddingsResponse{
				Data: []openaiEmbeddingsData{{
					Index:     0,
					Embedding: nil,
				}},
			}
			json.NewEncoder(w).Encode(resp)
			return
		}))
		defer s.Close()

		httpClient := s.Client()

		// HACK: override the URL to always go to the test server
		oldTransport := httpClient.Transport
		httpClient.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
			r.URL, _ = url.Parse(s.URL)
			return oldTransport.RoundTrip(r)
		})

		client := NewOpenAIClient(s.Client(), "")
		_, _, err := client.GenerateEmbeddings(context.Background(), codygateway.EmbeddingsRequest{
			Model: string(ModelNameOpenAIAda),
			Input: []string{"a", "b"},
		})
		require.Error(t, err, "expected request to error on failed retry")
	})
}

type roundTripFunc func(r *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
