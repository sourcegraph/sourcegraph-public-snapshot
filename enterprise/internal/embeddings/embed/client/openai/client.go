package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func NewClient(config *schema.Embeddings) *openaiEmbeddingsClient {
	return &openaiEmbeddingsClient{
		config: config,
		url:    getURL(config),
	}
}

func getURL(config *schema.Embeddings) string {
	url := config.Endpoint
	// Fallback to URL, it's the previous name of the setting.
	if url == "" {
		url = config.Url
	}
	// If that is also not set, use a sensible default.
	if url == "" {
		url = "https://api.openai.com/v1/embeddings"
	}
	return url
}

type openaiEmbeddingsClient struct {
	config *schema.Embeddings
	url    string
}

func (c *openaiEmbeddingsClient) GetDimensions() (int, error) {
	// Use some good default for the only model we supported so far.
	if c.config.Dimensions == 0 && strings.EqualFold(c.config.Model, "text-embedding-ada-002") {
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
func (c *openaiEmbeddingsClient) GetEmbeddings(ctx context.Context, texts []string) ([]float32, error) {
	embeddings, err := c.getEmbeddings(ctx, texts, c.config)
	if err == nil {
		return embeddings, nil
	}

	// TODO: Don't we already do retries in the HTTP CLI layer?
	// for i := 0; i < c.maxRetries; i++ {
	// 	embeddings, err = c.getEmbeddings(ctx, texts, c.config)
	// 	if err == nil {
	// 		return embeddings, nil
	// 	} else {
	// 		// Exponential delay
	// 		delay := time.Duration(int(math.Pow(float64(2), float64(i))))
	// 		select {
	// 		case <-ctx.Done():
	// 			return nil, ctx.Err()
	// 		case <-time.After(delay * time.Second):
	// 		}
	// 	}
	// }

	return nil, err
}

var modelsWithoutNewlines = map[string]struct{}{
	"text-embedding-ada-002": {},
}

func (c *openaiEmbeddingsClient) getEmbeddings(ctx context.Context, texts []string, config *schema.Embeddings) ([]float32, error) {
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(bodyBytes))
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
