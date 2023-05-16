package llmproxy

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/client/anthropic"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/client/openai"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const (
	ProviderName    = "llmproxy"
	DefaultEndpoint = "https://completions.sourcegraph.com"
)

func NewClient(cli httpcli.Doer, endpoint, accessToken string, model string) (types.CompletionsClient, error) {
	// TODO: Backcompat with older configs: We can remove this once S2 and k8s are migrated.
	endpoint = strings.TrimSuffix(endpoint, "/v1/completions/anthropic")
	llmProxyURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	if strings.ToLower(model) == "gpt-4" {
		return openai.NewClient(llmProxyDoer(cli, llmProxyURL, accessToken, "/v1/completions/openai"), "", model), nil
	}
	if strings.HasPrefix(strings.ToLower(model), "claude-") {
		return anthropic.NewClient(llmProxyDoer(cli, llmProxyURL, accessToken, "/v1/completions/anthropic"), "", model), nil
	}

	return nil, errors.Newf("no client known for upstream model %s", model)
}

func llmProxyDoer(upstream httpcli.Doer, llmProxyURL *url.URL, accessToken, path string) httpcli.Doer {
	return httpcli.DoerFunc(func(req *http.Request) (*http.Response, error) {
		req.Host = llmProxyURL.Host
		req.URL = llmProxyURL
		req.URL.Path = path
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
		return upstream.Do(req)
	})
}
