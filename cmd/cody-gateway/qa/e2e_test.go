package qa

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/internal/completions/client/fireworks"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func Test_Completions(t *testing.T) {
	gatewayURL, gatewayToken := parseBackendData(t)
	for _, f := range []codygateway.Feature{codygateway.FeatureCodeCompletions, codygateway.FeatureChatCompletions} {
		t.Run(string(f), func(t *testing.T) {
			for name, p := range map[string]GatewayFeatureClient{
				"anthropic": AnthropicGatewayFeatureClient{},
				"fireworks": FireworksGatewayFeatureClient{},
				"openai":    OpenAIGatewayFeatureClient{},
				"google":    GoogleGatewayFeatureClient{},
			} {
				t.Run(name, func(t *testing.T) {
					for _, stream := range []bool{false, true} {
						t.Run(fmt.Sprintf("stream %v", stream), func(t *testing.T) {
							stream := stream
							// avoid mutating the same URL
							u := *gatewayURL
							t.Parallel()
							req := &http.Request{URL: &u, Header: make(http.Header)}
							req.Header.Set("X-Sourcegraph-Feature", string(f))
							req.Header.Set("Authorization", "Bearer "+gatewayToken)
							req, err := p.GetRequest(f, req, stream)
							if errors.Is(err, errNotImplemented) {
								t.Skip(string(f), err)
							}
							require.NoError(t, err)
							resp, err := http.DefaultClient.Do(req)
							require.NoError(t, err)
							body, err := io.ReadAll(resp.Body)
							require.NoError(t, err)
							assert.Equal(t, http.StatusOK, resp.StatusCode, string(body))
							if stream {
								assert.Contains(t, resp.Header.Get("Content-Type"), "text/event-stream")
							}
						})
					}
				})
			}
		})
	}
}

func Test_Embeddings_OpenAI(t *testing.T) {
	t.Parallel()
	gatewayURL, gatewayToken := parseBackendData(t)
	gatewayURL.Path = "/v1/embeddings"
	for _, model := range []struct {
		name       string
		dimensions int
		// first float of a vector representing the input "Pls embed"
		firstValue float32
	}{
		{"openai/text-embedding-ada-002", 1536, -0.036106355},
		{"sourcegraph/st-multi-qa-mpnet-base-dot-v1", 768, -0.009880066},
	} {
		req, err := http.NewRequest("POST", gatewayURL.String(), strings.NewReader(fmt.Sprintf(`{"input": ["Pls embed"],"model": "%s"}`, model.name)))
		assert.NoError(t, err)
		req.Header.Set("X-Sourcegraph-Feature", string(codygateway.FeatureEmbeddings))
		req.Header.Set("Authorization", "Bearer "+gatewayToken)
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, resp.StatusCode, http.StatusOK, string(body))
		var response struct {
			Embeddings []struct {
				Data []float32 `json:"data"`
			} `json:"embeddings"`
		}
		err = json.Unmarshal(body, &response)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(response.Embeddings))
		assert.Equal(t, model.dimensions, len(response.Embeddings[0].Data))
		// This can drift somewhat, round the comparison to a few decimal places
		// to avoid diffs like:
		//
		//     expected: -0.036106355
		//     actual  : -0.03610423
		assert.Equal(t,
			fmt.Sprintf("%.4f", model.firstValue),
			fmt.Sprintf("%.4f", response.Embeddings[0].Data[0]))
	}
}

func Test_Finetuned_Fireworks_Completions(t *testing.T) {
	gatewayURL, gatewayToken := parseBackendData(t)
	t.Parallel()
	testConfig := map[string][]struct{ language, model string }{
		fireworks.FineTunedFIMVariant1: {{"", fireworks.FineTunedMixtralAll}},
		fireworks.FineTunedFIMVariant2: {{"typescript", fireworks.FineTunedMixtralTypescript}, {"typescriptreact",
			fireworks.FineTunedMixtralTypescript}, {"javascript", fireworks.FineTunedMixtralJavascript}, {"php", fireworks.FineTunedMixtralPhp}, {"python", fireworks.FineTunedMixtralPython}, {"badlanguage", fireworks.FineTunedMixtralAll}, {"", fireworks.FineTunedMixtralAll}},
		fireworks.FineTunedFIMVariant3: {{"", fireworks.FineTunedLlamaAll}},
		fireworks.FineTunedFIMVariant4: {{"typescript", fireworks.FineTunedLlamaTypescript}, {"typescriptreact",
			fireworks.FineTunedLlamaTypescript}, {"javascript", fireworks.FineTunedLlamaJavascript}, {"php", fireworks.FineTunedLlamaPhp}, {"python", fireworks.FineTunedLlamaPython}, {"badlanguage", fireworks.FineTunedLlamaAll}, {"", fireworks.FineTunedLlamaAll}},
	}
	for model, languageModel := range testConfig {
		t.Run(model, func(t *testing.T) {
			for _, l := range languageModel {
				t.Run(l.language, func(t *testing.T) {
					u := *gatewayURL
					req := &http.Request{URL: &u, Header: make(http.Header)}
					req.Header.Set("X-Sourcegraph-Feature", string(codygateway.FeatureCodeCompletions))
					req.Header.Set("Authorization", "Bearer "+gatewayToken)
					reqBody := fmt.Sprintf(`{
			"prompt":"def bubble_sort(arr):\n>",
			"maxTokensToSample":30,
			"model": "%s",
			"temperature":0.2,
			"topP":0.95,
			"stream":true,
			"languageId": "%s"
			}`, model, l.language)
					req.Method = "POST"
					req.URL.Path = "/v1/completions/fireworks"
					req.Body = io.NopCloser(strings.NewReader(reqBody))

					resp, err := http.DefaultClient.Do(req)

					assert.NoError(t, err)
					body, err := io.ReadAll(resp.Body)
					assert.NoError(t, err)
					assert.Equal(t, resp.StatusCode, http.StatusOK, string(body))
					assert.Contains(t, resp.Header.Get("Content-Type"), "text/event-stream")
					assert.Equal(t, resp.Header.Get("X-Cody-Resolved-Model"), "fireworks/"+l.model)
				})
			}
		})
	}
}

type GatewayFeatureClient interface {
	GetRequest(codygateway.Feature, *http.Request, bool) (*http.Request, error)
}

func parseBackendData(t *testing.T) (*url.URL, string) {
	if _, ok := os.LookupEnv("E2E_GATEWAY_ENDPOINT"); !ok {
		t.Skip("E2E_GATEWAY_ENDPOINT must be set, skipping")
	}
	gatewayEndpoint := os.Getenv("E2E_GATEWAY_ENDPOINT")
	gatewayURL, err := url.Parse(gatewayEndpoint)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := os.LookupEnv("E2E_GATEWAY_TOKEN"); !ok {
		t.Skip("E2E_GATEWAY_TOKEN must be set, skipping")
	}
	return gatewayURL, os.Getenv("E2E_GATEWAY_TOKEN")
}
