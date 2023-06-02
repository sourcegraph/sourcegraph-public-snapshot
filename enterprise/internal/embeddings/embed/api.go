package embed

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

type EmbeddingsClient interface {
	GetEmbeddings(ctx context.Context, texts []string) ([]float32, error)
	GetDimensions() (int, error)
}

func NewEmbeddingsClient(maxRetries int) (EmbeddingsClient, error) {
	c := conf.Get().Embeddings
	if c == nil || !c.Enabled {
		return nil, errors.New("embeddings are not configured or disabled")
	}

	switch c.Provider {
	case "sourcegraph":
		return &sourcegraphEmbeddingsClient{config: c, maxRetries: maxRetries}, nil
	case "openai":
		// TODO: Extract retrying logic into wrapper client.
		return &openaiEmbeddingsClient{config: c, maxRetries: maxRetries}, nil
	default:
		return nil, errors.Newf("invalid provider %q", c.Provider)
	}
}

type sourcegraphEmbeddingsClient struct {
	config     *schema.Embeddings
	maxRetries int
}

func (c *sourcegraphEmbeddingsClient) GetDimensions() (int, error) {
	// Use some good default for the only model we supported so far.
	if c.config.Dimensions == 0 && strings.EqualFold(c.config.Model, "openai/text-embedding-ada-002") {
		return 1536, nil
	}
	if c.config.Dimensions <= 0 {
		return 0, errors.New("invalid config for embeddings.dimensions, must be > 0")
	}
	return c.config.Dimensions, nil
}

// GetEmbeddingsWithRetries tries to embed the given texts using the external service specified in the config.
// In case of failure, it retries the embedding procedure up to maxRetries. This due to the OpenAI API which
// often hangs up when downloading large embedding responses.
func (c *sourcegraphEmbeddingsClient) GetEmbeddings(ctx context.Context, texts []string) ([]float32, error) {
	embeddings, err := c.getEmbeddings(ctx, texts, c.config)
	if err == nil {
		return embeddings, nil
	}

	return nil, err
}

func (c *sourcegraphEmbeddingsClient) getEmbeddings(ctx context.Context, texts []string, config *schema.Embeddings) ([]float32, error) {
	dimensions, err := c.GetDimensions()
	if err != nil {
		return nil, err
	}
	// TODO: This should not be done in the client layer IMO.
	_, replaceNewlines := modelsWithoutNewlines[config.Model]
	augmentedTexts := texts
	if replaceNewlines {
		augmentedTexts = make([]string, len(texts))
		// Replace newlines for certain (OpenAI) models, because they can negatively affect performance.
		for idx, text := range texts {
			augmentedTexts[idx] = strings.ReplaceAll(text, "\n", " ")
		}
	}

	request := codygateway.EmbeddingsRequest{Model: config.Model, Input: augmentedTexts}

	bodyBytes, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	url := config.Endpoint
	if url == "" {
		url = config.Url
	}
	if url == "" {
		url = "https://cody-gateway.sourcegraph.com/v1/embeddings"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.AccessToken)

	resp, err := httpcli.ExternalDoer.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		deadline, hasDeadline := ctx.Deadline()
		if hasDeadline {
			timeToWait := 1
			if deadline.Before(time.Now().Add(time.Duration(timeToWait) * time.Second)) {
				return nil, context.DeadlineExceeded
			}
		}
		// TODO: Handle 429 errors here. We should select on time.After, ctx.Done in here and wait until we should be doing the next request.
		// If we apply a context deadline further up, this should ideally not even attempt sleeping (instead, we can reschedule the job or so).
		return nil, errors.Errorf("embeddings: %s %q: failed with status %d: %s", req.Method, req.URL.String(), resp.StatusCode, string(respBody))
	}

	var response codygateway.EmbeddingsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	// Ensure embedding responses are sorted in the original order.
	sort.Slice(response.Embeddings, func(i, j int) bool {
		return response.Embeddings[i].Index < response.Embeddings[j].Index
	})

	embeddings := make([]float32, 0, len(response.Embeddings)*dimensions)
	for _, embedding := range response.Embeddings {
		embeddings = append(embeddings, embedding.Data...)
	}
	return embeddings, nil
}

type openaiEmbeddingsClient struct {
	config     *schema.Embeddings
	maxRetries int
}

func (c *openaiEmbeddingsClient) GetDimensions() (int, error) {
	// if strings.EqualFold(c.config.Provider, "openai") {
	// Use some good default for the only model we supported so far.
	if c.config.Dimensions == 0 && strings.EqualFold(c.config.Model, "text-embedding-ada-002") {
		return 1536, nil
	}
	if c.config.Dimensions <= 0 {
		return 0, errors.New("invalid config for embeddings.dimensions, must be > 0")
	}
	// }
	return c.config.Dimensions, nil
}

// GetEmbeddingsWithRetries tries to embed the given texts using the external service specified in the config.
// In case of failure, it retries the embedding procedure up to maxRetries. This due to the OpenAI API which
// often hangs up when downloading large embedding responses.
func (c *openaiEmbeddingsClient) GetEmbeddings(ctx context.Context, texts []string) ([]float32, error) {
	embeddings, err := c.getEmbeddings(ctx, texts, c.config)
	if err == nil {
		return embeddings, nil
	}

	// TODO: Don't we already do retries in the HTTP CLI layer?
	for i := 0; i < c.maxRetries; i++ {
		embeddings, err = c.getEmbeddings(ctx, texts, c.config)
		if err == nil {
			return embeddings, nil
		} else {
			// Exponential delay
			delay := time.Duration(int(math.Pow(float64(2), float64(i))))
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay * time.Second):
			}
		}
	}

	return nil, err
}

var modelsWithoutNewlines = map[string]struct{}{
	"text-embedding-ada-002": {},
}

func (c *openaiEmbeddingsClient) getEmbeddings(ctx context.Context, texts []string, config *schema.Embeddings) ([]float32, error) {
	dimensions, err := c.GetDimensions()
	if err != nil {
		return nil, err
	}
	// TODO: This should not be done in the client layer IMO.
	_, replaceNewlines := modelsWithoutNewlines[config.Model]
	augmentedTexts := texts
	if replaceNewlines {
		augmentedTexts = make([]string, len(texts))
		// Replace newlines for certain (OpenAI) models, because they can negatively affect performance.
		for idx, text := range texts {
			augmentedTexts[idx] = strings.ReplaceAll(text, "\n", " ")
		}
	}

	request := openaiEmbeddingAPIRequest{Model: config.Model, Input: augmentedTexts}

	bodyBytes, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	url := config.Endpoint
	if url == "" {
		url = config.Url
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.AccessToken)

	resp, err := httpcli.ExternalDoer.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		deadline, hasDeadline := ctx.Deadline()
		if hasDeadline {
			timeToWait := 1
			if deadline.Before(time.Now().Add(time.Duration(timeToWait) * time.Second)) {
				return nil, context.DeadlineExceeded
			}
		}
		// TODO: Handle 429 errors here. We should select on time.After, ctx.Done in here and wait until we should be doing the next request.
		// If we apply a context deadline further up, this should ideally not even attempt sleeping (instead, we can reschedule the job or so).
		return nil, errors.Errorf("embeddings: %s %q: failed with status %d: %s", req.Method, req.URL.String(), resp.StatusCode, string(respBody))
	}

	var response openaiEmbeddingAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	// Ensure embedding responses are sorted in the original order.
	sort.Slice(response.Data, func(i, j int) bool {
		return response.Data[i].Index < response.Data[j].Index
	})

	embeddings := make([]float32, 0, len(response.Data)*dimensions)
	for _, embedding := range response.Data {
		embeddings = append(embeddings, embedding.Embedding...)
	}
	return embeddings, nil
}

type openaiEmbeddingAPIRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type openaiEmbeddingAPIResponse struct {
	Data []struct {
		Index     int       `json:"index"`
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
}
