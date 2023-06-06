package sourcegraph

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/embed/client"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func NewClient(config *schema.SiteConfiguration) *sourcegraphEmbeddingsClient {
	return &sourcegraphEmbeddingsClient{
		model:       config.Embeddings.Model,
		dimensions:  config.Embeddings.Dimensions,
		url:         getURL(config.Embeddings),
		accessToken: getAccessToken(config),
	}
}

func getAccessToken(config *schema.SiteConfiguration) string {
	// If an access token is configured, use it.
	if config.Embeddings.AccessToken != "" {
		return config.Embeddings.AccessToken
	}
	// Otherwise, use the current license key to compute an access token.
	return licensing.GenerateLicenseKeyBasedAccessToken(config.LicenseKey)
}

func getURL(config *schema.Embeddings) string {
	url := config.Endpoint
	// Fallback to URL, it's the previous name of the setting.
	if url == "" {
		url = config.Url
	}
	// If that is also not set, use a sensible default.
	if url == "" {
		url = "https://cody-gateway.sourcegraph.com/v1/embeddings"
	}
	return url
}

type sourcegraphEmbeddingsClient struct {
	model       string
	dimensions  int
	url         string
	accessToken string
}

func (c *sourcegraphEmbeddingsClient) GetDimensions() (int, error) {
	// TODO: Later, we should ideally ask the gateway for the dimensionality of the model
	// so we don't have to hard-code defaults for all the models and can roll out new models
	// to older instances, too.

	// Use some good default for the only model we supported so far.
	if c.dimensions == 0 && strings.EqualFold(c.model, "openai/text-embedding-ada-002") {
		return 1536, nil
	}
	if c.dimensions <= 0 {
		return 0, errors.New("invalid config for embeddings.dimensions, must be > 0")
	}
	return c.dimensions, nil
}

var modelsWithoutNewlines = map[string]struct{}{
	"openai/text-embedding-ada-002": {},
}

// GetEmbeddingsWithRetries tries to embed the given texts using the external service specified in the config.
// In case of failure, it retries the embedding procedure up to maxRetries. This due to the OpenAI API which
// often hangs up when downloading large embedding responses.
func (c *sourcegraphEmbeddingsClient) GetEmbeddings(ctx context.Context, texts []string) ([]float32, error) {
	dimensions, err := c.GetDimensions()
	if err != nil {
		return nil, err
	}
	// TODO: This should not be done in the client layer IMO.
	_, replaceNewlines := modelsWithoutNewlines[c.model]
	augmentedTexts := texts
	if replaceNewlines {
		augmentedTexts = make([]string, len(texts))
		// Replace newlines for certain (OpenAI) models, because they can negatively affect performance.
		for idx, text := range texts {
			augmentedTexts[idx] = strings.ReplaceAll(text, "\n", " ")
		}
	}

	request := codygateway.EmbeddingsRequest{Model: c.model, Input: augmentedTexts}

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
		if resp.StatusCode == http.StatusTooManyRequests {
			retryAfterHeader := resp.Header.Get("retry-after")
			if retryAfterHeader != "" {
				// There are two valid formats for retry-after headers: seconds
				// until retry in int, or a RFC1123 date string.
				// First, see if it is denoted in seconds.
				s, err := strconv.Atoi(retryAfterHeader)
				// If denoted in seconds, only retry if we will get access within
				// the next retryAfterMaxSleepDuration seconds.
				if err == nil {
					return nil, client.NewRateLimitExceededError(time.Now().Add(time.Duration(s) * time.Second))
				}

				// If we weren't able to parse as seconds, try to parse as RFC1123.
				after, err := time.Parse(time.RFC1123, retryAfterHeader)
				if err == nil {
					return nil, client.NewRateLimitExceededError(after)
				}
				// We don't know how to parse this header, so let's just return a generic error.
			}
		}
		respBody, _ := io.ReadAll(resp.Body)
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
