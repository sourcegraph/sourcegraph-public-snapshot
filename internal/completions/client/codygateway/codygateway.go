package codygateway

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/internal/completions/client/anthropic"
	"github.com/sourcegraph/sourcegraph/internal/completions/client/fireworks"
	"github.com/sourcegraph/sourcegraph/internal/completions/client/google"
	"github.com/sourcegraph/sourcegraph/internal/completions/client/openai"
	"github.com/sourcegraph/sourcegraph/internal/completions/tokenusage"
	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewClient instantiates a completions provider backed by Sourcegraph's managed
// Cody Gateway service.
func NewClient(cli httpcli.Doer, endpoint, accessToken string, tokenManager tokenusage.Manager) (types.CompletionsClient, error) {
	gatewayURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	return &codyGatewayClient{
		upstream:     cli,
		gatewayURL:   gatewayURL,
		accessToken:  accessToken,
		tokenManager: tokenManager,
	}, nil
}

type codyGatewayClient struct {
	upstream     httpcli.Doer
	gatewayURL   *url.URL
	accessToken  string
	tokenManager tokenusage.Manager
}

func (c *codyGatewayClient) Stream(
	ctx context.Context, logger log.Logger, request types.CompletionRequest, sendEvent types.SendCompletionEvent) error {
	cc, err := c.clientForParams(logger, request.Feature, &request)
	if err != nil {
		return err
	}

	err = cc.Stream(ctx, logger, request, sendEvent)
	return overwriteErrSource(err)
}

func (c *codyGatewayClient) Complete(ctx context.Context, logger log.Logger, request types.CompletionRequest) (*types.CompletionResponse, error) {
	cc, err := c.clientForParams(logger, request.Feature, &request)
	if err != nil {
		return nil, err
	}
	resp, err := cc.Complete(ctx, logger, request)
	return resp, overwriteErrSource(err)
}

// overwriteErrSource should be used on all errors returned by an underlying
// types.CompletionsClient to avoid confusing error messages.
func overwriteErrSource(err error) error {
	if err == nil {
		return nil
	}
	if statusErr, ok := types.IsErrStatusNotOK(err); ok {
		statusErr.Source = "Sourcegraph Cody Gateway"
	}
	return err
}

func (c *codyGatewayClient) clientForParams(logger log.Logger, feature types.CompletionsFeature, request *types.CompletionRequest) (types.CompletionsClient, error) {
	model := request.ModelConfigInfo.Model
	logger.Info(
		"getting completions client for Cody Gateway routed request",
		log.String("mref", string(model.ModelRef)),
		log.String("modelName", model.ModelName))

	// Tease out the ProviderID and ModelID from the requested model.
	//
	// BUG: We are requiring that in order to use Cody Gateway, the site config's
	// provider name MUST match a well-known value, but this should not be required.
	//
	// e.g. If the site admin has configured a provider with the name "anthropic",
	// that is using the Sourcegraph API provider, then things will work. But if the
	// site admin named their provider "anthropic-via-cody-gateway", then this code
	// will fail, not understanding which API client should be used to process the
	// request.
	//
	// To fix this, we should require the use of the API Version field of the ModelRef,
	// and map _that_ to CompletionProviders. e.g. witth the API Version "anthropic/2023-01-01"
	// we could then determine which CompletionProvider to use, instead of having it
	// be based on the provider's ID.
	//
	// We will have this same issue for any API provider that serves different kinds of
	// models, requiring different API request types. (e.g. AWS Bedrock.)
	providerID := model.ModelRef.ProviderID() // e.g. "anthropic"

	// We then return a Completions Client specific to the API provider. Except, it is using the
	// HTTP endpoint of Cody Gateway. So our completions client will build the API-provider specific
	// HTTP request, which Cody Gateway will then proxy to the actual provider.
	//
	// IMPORTANT: We set the endpoint and access token for the API provider to "". The trick is that
	// the `httpcli.Doer` returned from `tokenManager` will route this to Cody Gateway, and use the
	// the codyGatewayClient's access token and endpoint.
	//
	// For Cody Pro / Sourcegraph.com this is even tricker, since that uses a different auth mechanism.
	// Hence why we update the access token used below based on the request we are resolving.
	// (Which assumes that this codyGatewayClient will NOT be reused across different requests.)
	if request.ModelConfigInfo.CodyProUserAccessToken != nil {
		c.accessToken = *request.ModelConfigInfo.CodyProUserAccessToken
	}

	switch conftypes.CompletionsProviderName(providerID) {
	case conftypes.CompletionsProviderNameAnthropic:
		doer := gatewayDoer(
			c.upstream, feature, c.gatewayURL, c.accessToken,
			"/v1/completions/anthropic-messages")
		client := anthropic.NewClient(doer, "", "", true, c.tokenManager)
		return client, nil

	case "mistral":
		// Annoying legacy hack: We expose Mistral model (e.g. "mixtral-8x22b-instruct") but have only
		// effer offered them via the Fireworks API provider. So when switching to the newer modelconfig
		// format, this is a situation where there wasn't a "mistral API Provider" for these models.
		// Instead, we just send these to fireworks.
		fallthrough
	case conftypes.CompletionsProviderNameFireworks:
		doer := gatewayDoer(
			c.upstream, feature, c.gatewayURL, c.accessToken,
			"/v1/completions/fireworks")
		client := fireworks.NewClient(doer, "", "")
		return client, nil

	case conftypes.CompletionsProviderNameGoogle:
		doer := gatewayDoer(
			c.upstream, feature, c.gatewayURL, c.accessToken,
			"/v1/completions/google")
		return google.NewClient(doer, "", "", true)

	case conftypes.CompletionsProviderNameOpenAI:
		doer := gatewayDoer(
			c.upstream, feature, c.gatewayURL, c.accessToken,
			"/v1/completions/openai")
		client := openai.NewClient(doer, "", "", c.tokenManager)
		return client, nil

	default:
		validProviderIDs := []conftypes.CompletionsProviderName{
			conftypes.CompletionsProviderNameAnthropic,
			conftypes.CompletionsProviderNameFireworks,
			conftypes.CompletionsProviderNameGoogle,
			conftypes.CompletionsProviderNameOpenAI,
		}
		return nil, errors.Newf(
			"to use Cody Gateway, the provider ID (%q) must match one of %v",
			providerID, validProviderIDs)
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (rt roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt(req)
}

// gatewayDoer redirects requests to Cody Gateway with all prerequisite headers.
func gatewayDoer(upstream httpcli.Doer, feature types.CompletionsFeature, gatewayURL *url.URL, accessToken, path string) httpcli.Doer {
	return httpcli.DoerFunc(func(req *http.Request) (*http.Response, error) {
		req.Host = gatewayURL.Host
		req.URL = gatewayURL
		req.URL.Path = path
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
		req.Header.Set(codygateway.FeatureHeaderName, string(feature))

		// HACK: Add actor transport directly. We tried adding the actor transport
		// in https://github.com/sourcegraph/sourcegraph/commit/6b058221ca87f5558759d92c0d72436cede70dc4
		// but it doesn't seem to work.
		resp, err := (&actor.HTTPTransport{
			RoundTripper: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
				return upstream.Do(r)
			}),
		}).RoundTrip(req)

		// If we get a repsonse, record Cody Gateway's x-trace response header,
		// so that we can link up to an event on our end if needed.
		if resp != nil && resp.Header != nil {
			if span := trace.SpanFromContext(req.Context()); span.SpanContext().IsValid() {
				// Would be cool if we can make an OTEL trace link instead, but
				// adding a link after a span has started is not supported yet:
				// https://github.com/open-telemetry/opentelemetry-specification/issues/454
				span.SetAttributes(attribute.String("cody-gateway.x-trace", resp.Header.Get("X-Trace")))
				span.SetAttributes(attribute.String("cody-gateway.x-trace-span", resp.Header.Get("X-Trace-Span")))
			}
		}

		return resp, err
	})
}
