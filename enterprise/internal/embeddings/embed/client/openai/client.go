package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func NewClient(config *schema.Embeddings) *openaiEmbeddingsClient {
	return &openaiEmbeddingsClient{
		dimensions:  config.Dimensions,
		accessToken: config.AccessToken,
		model:       getModel(config),
		url:         getURL(config),
	}
}

func getModel(config *schema.Embeddings) string {
	if config.Model == "" {
		return "text-embedding-ada-002"
	}
	return strings.ToLower(config.Model)
}

const defaultAPIURL = "https://api.openai.com/v1/embeddings"

func getURL(config *schema.Embeddings) string {
	url := config.Endpoint
	// Fallback to URL, it's the previous name of the setting.
	if url == "" {
		url = config.Url
	}
	// If that is also not set, use a sensible default.
	if url == "" {
		url = defaultAPIURL
	}
	return url
}

type openaiEmbeddingsClient struct {
	model       string
	dimensions  int
	url         string
	accessToken string
}

func (c *openaiEmbeddingsClient) GetDimensions() (int, error) {
	// Use some good default for the only model we supported so far.
	if c.dimensions == 0 && strings.EqualFold(c.model, "text-embedding-ada-002") {
		return 1536, nil
	}
	if c.dimensions <= 0 {
		return 0, errors.New("invalid config for embeddings.dimensions, must be > 0")
	}
	return c.dimensions, nil
}

func (c *openaiEmbeddingsClient) GetModelIdentifier() string {
	return fmt.Sprintf("openai/%s", c.model)
}

// GetEmbeddingsWithRetries tries to embed the given texts using the external service specified in the config.
// In case of failure, it retries the embedding procedure up to maxRetries. This due to the OpenAI API which
// often hangs up when downloading large embedding responses.
func (c *openaiEmbeddingsClient) GetEmbeddingsWithRetries(ctx context.Context, texts []string, maxRetries int) ([]float32, error) {
	embeddings, err := c.getEmbeddings(ctx, texts)
	if err == nil {
		return embeddings, nil
	}

	for i := 0; i < maxRetries; i++ {
		embeddings, err = c.getEmbeddings(ctx, texts)
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

func (c *openaiEmbeddingsClient) getEmbeddings(ctx context.Context, texts []string) ([]float32, error) {
	_, replaceNewlines := modelsWithoutNewlines[c.model]
	augmentedTexts := texts
	if replaceNewlines {
		augmentedTexts = make([]string, len(texts))
		// Replace newlines for certain (OpenAI) models, because they can negatively affect performance.
		for idx, text := range texts {
			augmentedTexts[idx] = strings.ReplaceAll(text, "\n", " ")
		}
	}

	request := openaiEmbeddingAPIRequest{Model: c.model, Input: augmentedTexts}

	bodyBytes, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := httpcli.ExternalDoer.Do(req)
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

	if len(response.Data) == 0 {
		return nil, nil
	}

	// Ensure embedding responses are sorted in the original order.
	sort.Slice(response.Data, func(i, j int) bool {
		return response.Data[i].Index < response.Data[j].Index
	})

	dimensionality := len(response.Data[0].Embedding)
	embeddings := make([]float32, 0, len(response.Data)*dimensionality)
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
