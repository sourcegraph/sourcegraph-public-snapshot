package llmproxy

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/client/anthropic"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/client/openai"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/types"
	llmproxy "github.com/sourcegraph/sourcegraph/enterprise/internal/llm-proxy"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const (
	ProviderName         = "llmproxy"
	DefaultEndpoint      = "https://completions.sourcegraph.com"
	openAIModelPrefix    = "openai/"
	anthropicModelPrefix = "anthropic/"
)

func NewClient(cli httpcli.Doer, endpoint, accessToken string) (types.CompletionsClient, error) {
	// TODO: Backcompat with older configs: We can remove this once S2 and k8s are migrated.
	endpoint = strings.TrimSuffix(endpoint, "/v1/completions/anthropic")
	llmProxyURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	return &llmProxyClient{
		upstream:    cli,
		llmProxyURL: llmProxyURL,
		accessToken: accessToken,
	}, nil
}

type llmProxyClient struct {
	upstream    httpcli.Doer
	llmProxyURL *url.URL
	accessToken string
}

func (c *llmProxyClient) Stream(ctx context.Context, feature types.CompletionsFeature, requestParams types.CompletionRequestParameters, sendEvent types.SendCompletionEvent) error {
	cc, err := c.clientForParams(feature, &requestParams)
	if err != nil {
		return err
	}
	return cc.Stream(ctx, feature, requestParams, sendEvent)
}

func (c *llmProxyClient) Complete(ctx context.Context, feature types.CompletionsFeature, requestParams types.CompletionRequestParameters) (*types.CompletionResponse, error) {
	cc, err := c.clientForParams(feature, &requestParams)
	if err != nil {
		return nil, err
	}
	return cc.Complete(ctx, feature, requestParams)
}

func (c *llmProxyClient) clientForParams(feature types.CompletionsFeature, requestParams *types.CompletionRequestParameters) (types.CompletionsClient, error) {
	model := strings.ToLower(requestParams.Model)

	if strings.HasPrefix(model, openAIModelPrefix) {
		requestParams.Model = strings.TrimPrefix(model, openAIModelPrefix)
		return openai.NewClient(llmProxyDoer(c.upstream, feature, c.llmProxyURL, c.accessToken, "/v1/completions/openai"), ""), nil
	}
	if strings.HasPrefix(model, anthropicModelPrefix) {
		requestParams.Model = strings.TrimPrefix(model, anthropicModelPrefix)
		return anthropic.NewClient(llmProxyDoer(c.upstream, feature, c.llmProxyURL, c.accessToken, "/v1/completions/anthropic"), "", ""), nil
	}
	return nil, errors.Newf("no client known for upstream model %s", model)
}

func llmProxyDoer(upstream httpcli.Doer, feature types.CompletionsFeature, llmProxyURL *url.URL, accessToken, path string) httpcli.Doer {
	return httpcli.DoerFunc(func(req *http.Request) (*http.Response, error) {
		req.Host = llmProxyURL.Host
		req.URL = llmProxyURL
		req.URL.Path = path
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
		req.Header.Set(llmproxy.LLMProxyFeatureHeaderName, string(feature))
		return upstream.Do(req)
	})
}
