package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"testing"

	jsonv2 "github.com/go-json-experiment/json"
	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
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

var inputSizes = []int{1, 512}

func Benchmark_JsonParsing_Response(b *testing.B) {
	for _, inputSize := range inputSizes {
		dat, err := os.ReadFile(fmt.Sprintf("testdata/%d.json", inputSize))
		if err != nil {
			b.Fatal(err)
		}
		b.ResetTimer()

		b.Run(strconv.Itoa(inputSize), func(b *testing.B) {
			b.Run("std", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					var resp openaiEmbeddingsResponse
					err := json.NewDecoder(bytes.NewReader(dat)).Decode(&resp)
					if err != nil {
						b.Fatal(err)
					}
				}
			})
			b.Run("jsoniter", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					var resp openaiEmbeddingsResponse
					err := jsoniter.NewDecoder(bytes.NewReader(dat)).Decode(&resp)
					if err != nil {
						b.Fatal(err)
					}
				}
			})
			b.Run("v2", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					var resp openaiEmbeddingsResponse
					err := jsonv2.UnmarshalRead(bytes.NewReader(dat), &resp)
					if err != nil {
						b.Fatal(err)
					}
				}
			})
		})
	}
}

func Test_JSON_V2_Matches_Stdlib(t *testing.T) {
	dat, err := os.ReadFile(fmt.Sprintf("testdata/512.json"))
	if err != nil {
		t.Fatal(err)
	}
	var v1, v2 openaiEmbeddingsResponse
	err = json.Unmarshal(dat, &v1)
	assert.NoError(t, err)
	err = jsonv2.Unmarshal(dat, &v2)
	assert.NoError(t, err)
	assert.Equal(t, v1, v2)
	assert.Equal(t, 512, len(v1.Data))
	assert.Equal(t, 512, len(v2.Data))
}
