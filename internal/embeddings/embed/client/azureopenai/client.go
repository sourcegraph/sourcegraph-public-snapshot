package azureopenai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/embeddings/embed/client"
	"github.com/sourcegraph/sourcegraph/internal/embeddings/embed/client/modeltransformations"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func NewClient(httpClient httpcli.Doer, config *conftypes.EmbeddingsConfig) *azureOpenaiEmbeddingsClient {
	return &azureOpenaiEmbeddingsClient{
		httpClient:  httpClient,
		dimensions:  config.Dimensions,
		accessToken: config.AccessToken,
		model:       config.Model,
		endpoint:    config.Endpoint,
	}
}

type azureOpenaiEmbeddingsClient struct {
	httpClient  httpcli.Doer
	model       string
	dimensions  int
	endpoint    string
	accessToken string
}

func (c *azureOpenaiEmbeddingsClient) GetDimensions() (int, error) {
	if c.dimensions <= 0 {
		return 0, errors.New("invalid config for embeddings.dimensions, must be > 0")
	}
	return c.dimensions, nil
}

func (c *azureOpenaiEmbeddingsClient) GetModelIdentifier() string {
	return fmt.Sprintf("azure-openai/%s", c.model)
}

func (c *azureOpenaiEmbeddingsClient) GetQueryEmbedding(ctx context.Context, query string) (*client.EmbeddingsResults, error) {
	return c.getEmbeddings(ctx, []string{modeltransformations.ApplyToQuery(query, c.GetModelIdentifier())})
}

func (c *azureOpenaiEmbeddingsClient) GetDocumentEmbeddings(ctx context.Context, documents []string) (*client.EmbeddingsResults, error) {
	return c.getEmbeddings(ctx, modeltransformations.ApplyToDocuments(documents, c.GetModelIdentifier()))
}

// getEmbeddings tries to embed the given texts using the external service specified in the config.
func (c *azureOpenaiEmbeddingsClient) getEmbeddings(ctx context.Context, texts []string) (*client.EmbeddingsResults, error) {
	for _, text := range texts {
		if text == "" {
			// The Azure OpenAI API will return an error if any of the strings in texts is an empty string,
			// so fail fast to avoid making tons of retryable requests.
			return nil, errors.New("cannot generate embeddings for an empty string")
		}
	}

	// For now, we assume all Azure OpenAI models will benefit from stripping out newlines.
	augmentedTexts := make([]string, len(texts))
	// Replace newlines for certain (OpenAI) models, because they can negatively affect performance.
	for idx, text := range texts {
		augmentedTexts[idx] = strings.ReplaceAll(text, "\n", " ")
	}

	embeddings := make([]float32, 0, len(augmentedTexts)*c.dimensions)
	failed := make([]int, 0)
	for i, input := range augmentedTexts {
		// This is a difference to the OpenAI implementation: Azure OpenAI currently
		// only supports a single input at a time, so we will need to fire off a request
		// for each of the texts individually.
		resp, err := c.requestSingleEmbeddingWithRetryOnNull(ctx, input, 3)
		if err != nil {
			failed = append(failed, i)

			// reslice to provide zero value embedding for failed chunk
			embeddings = embeddings[:len(embeddings)+c.dimensions]
			continue
		}
		embeddings = append(embeddings, resp.Data[0].Embedding...)
	}

	return &client.EmbeddingsResults{Embeddings: embeddings, Failed: failed, Dimensions: c.dimensions}, nil
}

func (c *azureOpenaiEmbeddingsClient) requestSingleEmbeddingWithRetryOnNull(ctx context.Context, input string, retries int) (*openaiEmbeddingAPIResponse, error) {
	for i := 0; i < retries; i++ {
		response, err := c.do(ctx, c.model, openaiEmbeddingAPIRequest{Input: input})
		if err != nil {
			return nil, err
		}
		if len(response.Data) != 1 || len(response.Data[0].Embedding) == 0 {
			continue
		}
		return response, nil
	}
	return nil, errors.Newf("null response for embedding after %d retries", retries)
}

func (c *azureOpenaiEmbeddingsClient) do(ctx context.Context, model string, request openaiEmbeddingAPIRequest) (*openaiEmbeddingAPIResponse, error) {
	bodyBytes, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	url, err := url.Parse(c.endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse configured endpoint")
	}
	q := url.Query()
	q.Add("api-version", "2023-05-15")
	url.RawQuery = q.Encode()
	url.Path = fmt.Sprintf("/openai/deployments/%s/embeddings", model)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", c.accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, errors.Errorf("embeddings: %s %q: failed with status %d: %s", req.Method, req.URL.String(), resp.StatusCode, string(respBody))
	}

	var response openaiEmbeddingAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return &response, nil
}

type openaiEmbeddingAPIRequest struct {
	Input string `json:"input"`
}

type openaiEmbeddingAPIResponse struct {
	Data []openaiEmbeddingAPIResponseData `json:"data"`
}

type openaiEmbeddingAPIResponseData struct {
	Index     int       `json:"index"`
	Embedding []float32 `json:"embedding"`
}
