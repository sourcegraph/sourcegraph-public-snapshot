package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sort"

	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/actor"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/response"
	"github.com/sourcegraph/sourcegraph/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func NewOpenAIClient(httpClient httpcli.Doer, accessToken string) EmbeddingsClient {
	return &openaiClient{
		httpClient:  httpClient,
		accessToken: accessToken,
	}
}

type openaiClient struct {
	httpClient  httpcli.Doer
	accessToken string
}

func (c *openaiClient) ProviderName() string { return "OpenAI" }

const apiURL = "https://api.openai.com/v1/embeddings"

func (c *openaiClient) GenerateEmbeddings(ctx context.Context, input codygateway.EmbeddingsRequest) (_ *codygateway.EmbeddingsResponse, _ int, err error) {
	for _, s := range input.Input {
		if s == "" {
			// The OpenAI API will return an error if any of the strings in texts is an empty string,
			// so fail fast to avoid making tons of retryable requests.
			return nil, 0, response.NewCustomHTTPStatusCodeError(http.StatusBadRequest, errors.New("cannot generate embeddings for an empty string"), -1)
		}
	}

	model, ok := openAIModelMappings[input.Model]
	if !ok {
		return nil, 0, response.NewCustomHTTPStatusCodeError(http.StatusBadRequest, errors.Newf("no OpenAI model found for %q", input.Model), -1)
	}

	response, err := c.requestEmbeddings(ctx, model, input.Input)
	if err != nil {
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
		ModelDimensions: model.dimensions,
	}, response.Usage.TotalTokens, nil
}

func (c *openaiClient) requestEmbeddings(ctx context.Context, model openAIModel, input []string) (*openaiEmbeddingsResponse, error) {
	act := actor.FromContext(ctx)

	request := openaiEmbeddingsRequest{
		Model: model.upstreamName,
		Input: input,
		// Set the actor ID for upstream tracking.
		User: act.ID,
	}

	bodyBytes, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 || resp.StatusCode < 200 {
		// If we are being rate limited by OpenAI, we don't want to forward that error and instead
		// return a 503 to the client. It's not them being limited, it's us and that an operations
		// error on our side.
		if resp.StatusCode == http.StatusTooManyRequests {
			return nil, response.NewCustomHTTPStatusCodeError(http.StatusServiceUnavailable,
				errors.New("we're facing too much load at the moment, please retry later"), resp.StatusCode)
		}

		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))

		// If OpenAI tells us we gave them a bad request, blame the client and
		// tell them.
		if resp.StatusCode == http.StatusBadRequest {
			return nil, response.NewHTTPStatusCodeError(http.StatusBadRequest,
				errors.Newf("bad request: %s", string(respBody)))
		}

		// We don't forward other status codes, we just return a generic error
		// instead.
		return nil, errors.Errorf("embeddings: %s %q: failed with status %d: %s",
			req.Method, req.URL.String(), resp.StatusCode, string(respBody))
	}

	var response openaiEmbeddingsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		// Although we might've incurred cost at this point, we don't want to count
		// that towards the rate limit of the requester, so return 0 for the consumed
		// token count.
		return nil, err
	}

	return &response, nil
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
