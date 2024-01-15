package qa

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func Test_Completions(t *testing.T) {
	gatewayURL, gatewayToken := parseBackendData(t)
	for _, f := range []codygateway.Feature{codygateway.FeatureCodeCompletions, codygateway.FeatureChatCompletions} {
		for name, p := range map[string]GatewayFeatureClient{"anthropic": AnthropicGatewayFeatureClient{}, "fireworks": FireworksGatewayFeatureClient{}, "openai": OpenAIGatewayFeatureClient{}} {
			for _, stream := range []bool{false, true} {
				t.Run(string(f)+" "+name+" stream "+strconv.FormatBool(stream), func(t *testing.T) {
					stream := stream
					// avoid mutating the same URL
					url := *gatewayURL
					t.Parallel()
					req := &http.Request{URL: &url, Header: make(http.Header)}
					req.Header.Set("X-Sourcegraph-Feature", string(f))
					req.Header.Set("Authorization", "Bearer "+gatewayToken)
					req, err := p.GetRequest(f, req, stream)
					if err != nil && errors.Is(err, errNotImplemented) {
						t.Skip(string(f), err)
					}
					assert.NoError(t, err)
					resp, err := http.DefaultClient.Do(req)
					assert.NoError(t, err)
					body, err := io.ReadAll(resp.Body)
					assert.NoError(t, err)
					assert.Equal(t, resp.StatusCode, http.StatusOK, string(body))
					if stream {
						assert.Contains(t, resp.Header.Get("Content-Type"), "text/event-stream")
					}
				})
			}
		}
	}
}

func parseBackendData(t *testing.T) (*url.URL, string) {
	if _, ok := os.LookupEnv("E2E_GATEWAY_ENDPOINT"); !ok {
		t.Fatal("E2E_GATEWAY_ENDPOINT must be set")
	}
	gatewayEndpoint := os.Getenv("E2E_GATEWAY_ENDPOINT")
	gatewayURL, err := url.Parse(gatewayEndpoint)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := os.LookupEnv("E2E_GATEWAY_TOKEN"); !ok {
		t.Fatal("E2E_GATEWAY_TOKEN must be set")
	}
	return gatewayURL, os.Getenv("E2E_GATEWAY_TOKEN")
}

func Test_Embeddings(t *testing.T) {
	t.Parallel()
	gatewayURL, gatewayToken := parseBackendData(t)
	gatewayURL.Path = "/v1/embeddings"
	req, err := http.NewRequest("POST", gatewayURL.String(), strings.NewReader(`{"input": ["Pls embed"],"model": "openai/text-embedding-ada-002"}`))
	if err != nil {
		t.Fail()
	}
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
	assert.Equal(t, 1536, len(response.Embeddings[0].Data))
}

type GatewayFeatureClient interface {
	GetRequest(codygateway.Feature, *http.Request, bool) (*http.Request, error)
}

type AnthropicGatewayFeatureClient struct {
}

type FireworksGatewayFeatureClient struct {
}

func (fc FireworksGatewayFeatureClient) GetRequest(f codygateway.Feature, req *http.Request, stream bool) (*http.Request, error) {
	if f == codygateway.FeatureCodeCompletions {
		body := `{"prompt":"def bubble_sort(arr):\n>","maxTokensToSample":30,"model":"accounts/fireworks/models/starcoder-16b-w8a16","temperature":0.2,"topP":0.95, "stream":` + strconv.FormatBool(stream) + `}`
		req.Method = "POST"
		req.URL.Path = "/v1/completions/fireworks"
		req.Body = io.NopCloser(strings.NewReader(body))
		return req, nil
	}
	if f == codygateway.FeatureChatCompletions {
		body := `{"model":"accounts/fireworks/models/mixtral-8x7b-instruct","messages":[{"role":"user","content":"You are Cody"},{"role":"assistant","content":"Ok, I am Cody"},{"role":"user","content":"What is your real name name though?"}],"n":1,"max_tokens":30,"temperature":0.2,"top_p":0.95, "stream":` + strconv.FormatBool(stream) + `}`
		req.Method = "POST"
		req.URL.Path = "/v1/completions/fireworks"
		req.Body = io.NopCloser(strings.NewReader(body))
		return req, nil
	}
	return nil, errors.New("unknown feature: " + string(f))
}

type OpenAIGatewayFeatureClient struct {
}

func (o OpenAIGatewayFeatureClient) GetRequest(f codygateway.Feature, req *http.Request, stream bool) (*http.Request, error) {
	if f == codygateway.FeatureCodeCompletions {
		return nil, errNotImplemented
	}
	if f == codygateway.FeatureChatCompletions {
		body := `{"model":"gpt-4-1106-preview","messages":[{"role":"user","content":"You are Cody"},{"role":"assistant","content":"Ok, I am Cody"},{"role":"user","content":"What is your real name name though?"}],"n":1,"max_tokens":30,"temperature":0.2,"top_p":0.95, "stream":` + strconv.FormatBool(stream) + `}`
		req.Method = "POST"
		req.URL.Path = "/v1/completions/openai"
		req.Body = io.NopCloser(strings.NewReader(body))
		return req, nil
	}
	return nil, errors.New("unknown feature: " + string(f))
}

var errNotImplemented = errors.New("not implemented")

var _ GatewayFeatureClient = AnthropicGatewayFeatureClient{}
var _ GatewayFeatureClient = FireworksGatewayFeatureClient{}
var _ GatewayFeatureClient = OpenAIGatewayFeatureClient{}

func (a AnthropicGatewayFeatureClient) GetRequest(f codygateway.Feature, req *http.Request, stream bool) (*http.Request, error) {
	if f == codygateway.FeatureCodeCompletions {
		body := `{"model":"claude-instant-1","prompt":"Human: def bubble_sort(arr):\nAssistant: ","n":1,"max_tokens_to_sample":30,"temperature":0.2,"top_p":0.95, "stream":` + strconv.FormatBool(stream) + `}`
		req.Method = "POST"
		req.URL.Path = "/v1/completions/anthropic"
		req.Body = io.NopCloser(strings.NewReader(body))
		return req, nil
	}
	if f == codygateway.FeatureChatCompletions {
		body := `{"model":"claude-2.1","prompt":"Human: Your are Cody?:\nAssistant: I am Cody\nHuman: What is your real name?\nAssistant:","n":1,"max_tokens_to_sample":30,"temperature":0.2,"top_p":0.95, "stream":` + strconv.FormatBool(stream) + `}`

		req.Method = "POST"
		req.URL.Path = "/v1/completions/anthropic"
		req.Body = io.NopCloser(strings.NewReader(body))
		return req, nil
	}
	return nil, errors.New("unknown feature: " + string(f))
}
