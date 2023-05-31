package codygateway

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/client/anthropic"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/client/openai"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const (
	// ProviderName is 'sourcegraph', since this is a Sourcegraph-provided service,
	// backed by Cody Gateway. This is the value accepted in site configuration.
	ProviderName    = "sourcegraph"
	DefaultEndpoint = "https://cody-gateway.sourcegraph.com"

	openAIModelPrefix    = "openai/"
	anthropicModelPrefix = "anthropic/"
)

// NewClient instantiates a completions provider backed by Sourcegraph's managed
// Cody Gateway service.
func NewClient(cli httpcli.Doer, endpoint, accessToken string) (types.CompletionsClient, error) {
	// TODO: Backcompat with older configs: We can remove this once S2 and k8s are migrated.
	endpoint = strings.TrimSuffix(endpoint, "/v1/completions/anthropic")
	gatewayURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	return &codyGatewayClient{
		upstream:    cli,
		gatewayURL:  gatewayURL,
		accessToken: accessToken,
	}, nil
}

type codyGatewayClient struct {
	upstream    httpcli.Doer
	gatewayURL  *url.URL
	accessToken string
}

func (c *codyGatewayClient) Stream(ctx context.Context, feature types.CompletionsFeature, requestParams types.CompletionRequestParameters, sendEvent types.SendCompletionEvent) error {
	cc, err := c.clientForParams(feature, &requestParams)
	if err != nil {
		return err
	}
	return cc.Stream(ctx, feature, requestParams, sendEvent)
}

func (c *codyGatewayClient) Complete(ctx context.Context, feature types.CompletionsFeature, requestParams types.CompletionRequestParameters) (*types.CompletionResponse, error) {
	cc, err := c.clientForParams(feature, &requestParams)
	if err != nil {
		return nil, err
	}
	return cc.Complete(ctx, feature, requestParams)
}

func (c *codyGatewayClient) clientForParams(feature types.CompletionsFeature, requestParams *types.CompletionRequestParameters) (types.CompletionsClient, error) {
	model := strings.ToLower(requestParams.Model)

	if strings.HasPrefix(model, openAIModelPrefix) {
		requestParams.Model = strings.TrimPrefix(model, openAIModelPrefix)
		return openai.NewClient(gatewayDoer(c.upstream, feature, c.gatewayURL, c.accessToken, "/v1/completions/openai"), "", ""), nil
	}
	if strings.HasPrefix(model, anthropicModelPrefix) {
		requestParams.Model = strings.TrimPrefix(model, anthropicModelPrefix)
		return anthropic.NewClient(gatewayDoer(c.upstream, feature, c.gatewayURL, c.accessToken, "/v1/completions/anthropic"), "", ""), nil
	}
	return nil, errors.Newf("no client known for upstream model %s", model)
}

func gatewayDoer(upstream httpcli.Doer, feature types.CompletionsFeature, gatewayURL *url.URL, accessToken, path string) httpcli.Doer {
	return httpcli.DoerFunc(func(req *http.Request) (*http.Response, error) {
		req.Host = gatewayURL.Host
		req.URL = gatewayURL
		req.URL.Path = path
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
		req.Header.Set(codygateway.FeatureHeaderName, string(feature))
		return upstream.Do(req)
	})
}
