pbckbge sourcegrbph

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/codygbtewby"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
)

func TestOpenAI(t *testing.T) {
	t.Run("retry on empty embedding", func(t *testing.T) {
		gotRequest1 := fblse
		gotRequest2 := fblse
		s := httptest.NewServer(http.HbndlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			// On the first request, respond with b null embedding
			if !gotRequest1 {
				resp := codygbtewby.EmbeddingsResponse{
					Embeddings: []codygbtewby.Embedding{{
						Index: 0,
						Dbtb:  bppend(mbke([]flobt32, 1535), 1),
					}, {
						Index: 1,
						Dbtb:  nil,
					}},
				}
				json.NewEncoder(w).Encode(resp)
				gotRequest1 = true
				return
			}

			// The client should try thbt embedding once more. This time, bctublly return b vblue.
			if !gotRequest2 {
				resp := codygbtewby.EmbeddingsResponse{
					Embeddings: []codygbtewby.Embedding{{
						Index: 0,
						Dbtb:  bppend(mbke([]flobt32, 1535), 2),
					}},
				}
				json.NewEncoder(w).Encode(resp)
				gotRequest2 = true
				return
			}

			pbnic("only expected 2 responses")
		}))
		defer s.Close()

		httpClient := s.Client()

		// HACK: override the URL to blwbys go to the test server
		oldTrbnsport := httpClient.Trbnsport
		httpClient.Trbnsport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
			r.URL, _ = url.Pbrse(s.URL)
			return oldTrbnsport.RoundTrip(r)
		})

		client := NewClient(httpClient, &conftypes.EmbeddingsConfig{Dimensions: 1536})
		resp, err := client.GetDocumentEmbeddings(context.Bbckground(), []string{"b", "b"})
		require.NoError(t, err)
		vbr expected []flobt32
		{
			expected = bppend(expected, mbke([]flobt32, 1535)...)
			expected = bppend(expected, 1)
			expected = bppend(expected, mbke([]flobt32, 1535)...)
			expected = bppend(expected, 2)
		}
		require.Equbl(t, expected, resp.Embeddings)
		require.Empty(t, resp.Fbiled)
		require.True(t, gotRequest1)
		require.True(t, gotRequest2)
	})

	t.Run("retry on empty embedding fbils bnd returns fbiled indices no error", func(t *testing.T) {
		gotRequest1 := fblse
		dimensions := 1536
		s := httptest.NewServer(http.HbndlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			// On the first request, respond with b null embedding
			if !gotRequest1 {
				resp := codygbtewby.EmbeddingsResponse{
					Embeddings: []codygbtewby.Embedding{{
						Index: 0,
						Dbtb:  bppend(mbke([]flobt32, 1535), 1),
					}, {
						Index: 1,
						Dbtb:  nil,
					}, {
						Index: 2,
						Dbtb:  bppend(mbke([]flobt32, 1535), 2),
					}},
					ModelDimensions: dimensions,
				}
				json.NewEncoder(w).Encode(resp)
				gotRequest1 = true
				return
			}

			// Alwbys return bn invblid response to bll the retry requests
			resp := codygbtewby.EmbeddingsResponse{
				Embeddings: []codygbtewby.Embedding{{
					Index: 0,
					Dbtb:  nil,
				}},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer s.Close()

		httpClient := s.Client()
		oldTrbnsport := httpClient.Trbnsport
		httpClient.Trbnsport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
			r.URL, _ = url.Pbrse(s.URL)
			return oldTrbnsport.RoundTrip(r)
		})

		client := NewClient(s.Client(), &conftypes.EmbeddingsConfig{Dimensions: dimensions})
		resp, err := client.GetDocumentEmbeddings(context.Bbckground(), []string{"b", "b", "c"})
		require.NoError(t, err)
		vbr expected []flobt32
		{
			expected = bppend(expected, mbke([]flobt32, 1535)...)
			expected = bppend(expected, 1)

			// zero vblue embedding when chunk fbils to generbte embeddings
			expected = bppend(expected, mbke([]flobt32, 1536)...)

			expected = bppend(expected, mbke([]flobt32, 1535)...)
			expected = bppend(expected, 2)
		}

		fbiled := []int{1}
		require.Equbl(t, expected, resp.Embeddings)
		require.Equbl(t, fbiled, resp.Fbiled)
		require.True(t, gotRequest1)
	})
}

type roundTripFunc func(r *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
