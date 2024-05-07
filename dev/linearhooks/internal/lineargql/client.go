package lineargql

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Khan/genqlient/graphql"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type tokenAuthTransport struct {
	token string

	wrapped http.RoundTripper
}

func (t *tokenAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", t.token)
	return t.wrapped.RoundTrip(req)
}

func NewGraphQLClient(endpoint string, token string) graphql.Client {
	return newTracedClient(endpoint, &http.Client{
		Transport: &tokenAuthTransport{
			token:   token,
			wrapped: http.DefaultTransport,
		},
	})
}

func newTracedClient(endpoint string, httpClient graphql.Doer) graphql.Client {
	client := graphql.NewClient(endpoint, httpClient)
	return &tracedClient{endpoint: endpoint, client: client}
}

type tracedClient struct {
	endpoint string
	client   graphql.Client
}

var tracer = otel.Tracer("internal/lineargql")

func (tc *tracedClient) MakeRequest(
	ctx context.Context,
	req *graphql.Request,
	resp *graphql.Response,
) error {
	// Start a span
	ctx, span := tracer.Start(ctx, fmt.Sprintf("GraphQL: %s", req.OpName),
		trace.WithAttributes(
			attribute.String("endpoint", tc.endpoint),
			attribute.String("query", req.Query),
		))

	// Do the request
	err := tc.client.MakeRequest(ctx, req, resp)

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
