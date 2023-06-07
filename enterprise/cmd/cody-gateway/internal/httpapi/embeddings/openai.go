package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sort"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/response"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func NewOpenAIClient(accessToken string) EmbeddingsClient {
	return &openaiClient{
		accessToken: accessToken,
	}
}

type openaiClient struct {
	accessToken string
}

const apiURL = "https://api.openai.com/v1/embeddings"

func (c *openaiClient) GenerateEmbeddings(ctx context.Context, input codygateway.EmbeddingsRequest) (*codygateway.EmbeddingsResponse, int, error) {
	openAIModel, ok := openAIModelMappings[input.Model]
	if !ok {
		return nil, 0, response.NewHTTPStatusCodeError(http.StatusBadRequest, errors.Newf("no OpenAI model found for %q", input.Model))
	}

	request := openaiEmbeddingsRequest{
		Model: openAIModel.upstreamName,
		Input: input.Input,
		// TODO: Maybe set user.
	}

	bodyBytes, err := json.Marshal(request)
	if err != nil {
		return nil, 0, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := httpcli.ExternalDoer.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 || resp.StatusCode < 200 {
		// If we are being rate limited by OpenAI, we don't want to forward that error and instead
		// return a 503 to the client. It's not them being limited, it's us and that an operations
		// error on our side.
		if resp.StatusCode == http.StatusTooManyRequests {
			return nil, 0, response.NewHTTPStatusCodeError(http.StatusServiceUnavailable, errors.Newf("we're facing too much load at the moment, please retry"))
		}
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		// We don't forward the status code here, everything but 429 should turn into a 500 error.
		return nil, 0, errors.Errorf("embeddings: %s %q: failed with status %d: %s", req.Method, req.URL.String(), resp.StatusCode, string(respBody))
	}

	var response openaiEmbeddingsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		// Although we might've incurred cost at this point, we don't want to count
		// that towards the rate limit of the requester, so return 0 for the consumed
		// token count.
		return nil, 0, err
	}

	// Ensure embedding responses are sorted in the original order.
	sort.Slice(response.Data, func(i, j int) bool {
		return response.Data[i].Index < response.Data[j].Index
	})

	embeddings := make([]codygateway.Embedding, len(response.Data))
	for i, d := range response.Data {
		embeddings[i] = codygateway.Embedding{
			Index: d.Index,
			Data:  d.Embedding,
		}
	}

	return &codygateway.EmbeddingsResponse{
		Embeddings:      embeddings,
		Model:           response.Model,
		ModelDimensions: openAIModel.dimensions,
	}, response.Usage.TotalTokens, nil
}

type openaiEmbeddingsRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
	User  string   `json:"user"`
}

type openaiEmbeddingsUsage struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

type openaiEmbeddingsData struct {
	Index     int       `json:"index"`
	Embedding []float32 `json:"embedding"`
}

type openaiEmbeddingsResponse struct {
	Model string                 `json:"model"`
	Usage openaiEmbeddingsUsage  `json:"usage"`
	Data  []openaiEmbeddingsData `json:"data"`
}

type openAIModel struct {
	upstreamName string
	dimensions   int
}

var openAIModelMappings = map[string]openAIModel{
	string(ModelNameOpenAIAda): {
		upstreamName: "text-embedding-ada-002",
		dimensions:   1536,
	},
}
