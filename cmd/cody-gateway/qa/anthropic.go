package qa

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type AnthropicGatewayFeatureClient struct{}

func (a AnthropicGatewayFeatureClient) GetRequest(f codygateway.Feature, req *http.Request, stream bool) (*http.Request, error) {
	if f == codygateway.FeatureCodeCompletions {
		body := fmt.Sprintf(`{
			"model":"claude-instant-1",
			"prompt":"Human: def bubble_sort(arr):\nAssistant: ",
			"n":1,
			"max_tokens_to_sample":30,
			"temperature":0.2,
			"top_p":0.95,
			"stream":%t
		}`, stream)
		req.Method = "POST"
		req.URL.Path = "/v1/completions/anthropic"
		req.Body = io.NopCloser(strings.NewReader(body))
		return req, nil
	}
	if f == codygateway.FeatureChatCompletions {
		body := fmt.Sprintf(`{
			"model":"claude-2.1",
			"prompt":"Human: Your are Cody?:\nAssistant: I am Cody\nHuman: What is your real name?\nAssistant:",
			"n":1,
			"max_tokens_to_sample":30,
			"temperature":0.2,
			"top_p":0.95,
			"stream":%t
		}`, stream)

		req.Method = "POST"
		req.URL.Path = "/v1/completions/anthropic"
		req.Body = io.NopCloser(strings.NewReader(body))
		return req, nil
	}
	return nil, errors.New("unknown feature: " + string(f))
}
