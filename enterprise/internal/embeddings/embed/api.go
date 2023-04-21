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
	"sync"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

type EmbeddingAPIRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type EmbeddingAPIResponse struct {
	Data []struct {
		Index     int       `json:"index"`
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
}

type EmbeddingsClient interface {
	GetEmbeddingsWithRetries(ctx context.Context, texts []string, maxRetries int) ([]float32, error)
	GetDimensions() (int, error)
}

func NewEmbeddingsClient() EmbeddingsClient {
	client := &embeddingsClient{config: conf.Get().Embeddings}

	mu := sync.Mutex{}
	conf.Watch(func() {
		mu.Lock()
		defer mu.Unlock()
		client.setConfig(conf.Get().Embeddings)
	})

	return client
}

type embeddingsClient struct {
	config *schema.Embeddings
}

func (c *embeddingsClient) isDisabled() bool {
	return c.config == nil || !c.config.Enabled
}

func (c *embeddingsClient) setConfig(config *schema.Embeddings) {
	c.config = config
}

func (c *embeddingsClient) GetDimensions() (int, error) {
	if c.isDisabled() {
		return -1, errors.New("embeddings are not configured or disabled")
	}
	return c.config.Dimensions, nil
}

// GetEmbeddingsWithRetries tries to embed the given texts using the external service specified in the config.
// In case of failure, it retries the embedding procedure up to maxRetries. This due to the OpenAI API which
// often hangs up when downloading large embedding responses.
func (c *embeddingsClient) GetEmbeddingsWithRetries(ctx context.Context, texts []string, maxRetries int) ([]float32, error) {
	if c.isDisabled() {
		return nil, errors.New("embeddings are not configured or disabled")
	}

	embeddings, err := getEmbeddings(ctx, texts, c.config)
	if err == nil {
		return embeddings, nil
	}

	for i := 0; i < maxRetries; i++ {
		embeddings, err = getEmbeddings(ctx, texts, c.config)
		if err == nil {
			return embeddings, nil
		} else {
			// Exponential delay
			delay := time.Duration(int(math.Pow(float64(2), float64(i))))
			time.Sleep(delay * time.Second)
		}
	}

	return nil, err
}

func getEmbeddings(ctx context.Context, texts []string, config *schema.Embeddings) ([]float32, error) {
	// Replace newlines, which can negatively affect performance.
	augmentedTexts := make([]string, len(texts))
	for idx, text := range texts {
		augmentedTexts[idx] = strings.ReplaceAll(text, "\n", " ")
	}

	request := EmbeddingAPIRequest{Model: config.Model, Input: augmentedTexts}

	bodyBytes, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", config.Url, bytes.NewReader(bodyBytes))
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
		return nil, errors.Errorf("embeddings: %s %q: failed with status %d: %s", req.Method, req.URL.String(), resp.StatusCode, string(respBody))
	}

	var response EmbeddingAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	// Ensure embedding responses are sorted in the original order.
	sort.Slice(response.Data, func(i, j int) bool {
		return response.Data[i].Index < response.Data[j].Index
	})

	embeddings := make([]float32, 0, len(response.Data)*config.Dimensions)
	for _, embedding := range response.Data {
		embeddings = append(embeddings, embedding.Embedding...)
	}
	return embeddings, nil
}
