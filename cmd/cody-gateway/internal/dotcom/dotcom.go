package dotcom

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Khan/genqlient/graphql"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// NewClient returns a new GraphQL client for the Sourcegraph.com API authenticated
// with the given Sourcegraph access token.
//
// To use, add a query or mutation to operations.graphql and use the generated
// functions and types with the client, for example:
//
//	c := dotcom.NewClient(sourcegraphToken)
//	resp, err := dotcom.CheckAccessToken(ctx, c, licenseToken)
//	if err != nil {
//		log.Fatal(err)
//	}
//	println(resp.GetDotcom().ProductSubscriptionByAccessToken.LlmProxyAccess.Enabled)
//
// The client generator automatically ensures we're up-to-date with the GraphQL schema.
func NewClient(endpoint, token string) graphql.Client {
	return &tracedClient{graphql.NewClient(endpoint, &http.Client{
		Transport: &tokenAuthTransport{
			token:   token,
			wrapped: http.DefaultTransport,
		},
	})}
}

// tracedClient instruments graphql.Client with OpenTelemetry tracing.
type tracedClient struct{ c graphql.Client }

var tracer = otel.Tracer("cody-gateway/internal/dotcom")

func (tc *tracedClient) MakeRequest(
	ctx context.Context,
	req *graphql.Request,
	resp *graphql.Response,
) error {
	// Start a span
	ctx, span := tracer.Start(ctx, fmt.Sprintf("GraphQL: %s", req.OpName),
		trace.WithAttributes(attribute.String("query", req.Query)))

	// Do the request
	err := tc.c.MakeRequest(ctx, req, resp)

	// Assess the result
	if err != nil {
		span.RecordError(err)
	}
	if len(resp.Errors) > 0 {
		span.RecordError(resp.Errors)
	}
	span.End()

	return err
}

// tokenAuthTransport adds token header authentication to requests.
type tokenAuthTransport struct {
	token   string
	wrapped http.RoundTripper
}

func (t *tokenAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("token %s", t.token))
	return t.wrapped.RoundTrip(req)
}
