package llmproxy

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/streaming/anthropic"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

const ProviderName = "llmproxy"

var llmProxyURL = mustParseURL("https://llm-proxy-ytplpvhyiq-uc.a.run.app/v1/completions/anthropic")

type llmProxyAnthropicClient struct {
	cli             httpcli.Doer
	accessToken     string
	model           string
	anthropicClient types.CompletionsClient
}

func NewClient(cli httpcli.Doer, accessToken string, model string) types.CompletionsClient {
	anthropicDoer := httpcli.DoerFunc(func(req *http.Request) (*http.Response, error) {
		req.Host = llmProxyURL.Host
		req.URL = llmProxyURL
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
		return cli.Do(req)
	})
	return &llmProxyAnthropicClient{
		cli:             cli,
		accessToken:     accessToken,
		model:           model,
		anthropicClient: anthropic.NewAnthropicClient(anthropicDoer, "", model),
	}
}

func (a *llmProxyAnthropicClient) Complete(
	ctx context.Context,
	requestParams types.CodeCompletionRequestParameters,
) (*types.CodeCompletionResponse, error) {
	return a.anthropicClient.Complete(ctx, requestParams)
}

func (a *llmProxyAnthropicClient) Stream(
	ctx context.Context,
	requestParams types.ChatCompletionRequestParameters,
	sendEvent types.SendCompletionEvent,
) error {
	return a.anthropicClient.Stream(ctx, requestParams, sendEvent)
}

func mustParseURL(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(fmt.Sprintf("failed to parse URL %q: %s", s, err.Error()))
	}
	return u
}
