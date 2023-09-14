package dotcom

import (
	"context"
	"net/http"

	"github.com/Khan/genqlient/graphql"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// Client is a type alias to graphql.Client that should be used to communicate
// that this graphql.Client is for dotcom.
type Client graphql.Client

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
func NewClient(externalHTTPClient httpcli.Doer, endpoint, token string) Client {
	// TODO(keegancsmith) we allow unauthed requests for now but should
	// require it when promoting guardrails for use.
	httpClient := externalHTTPClient
	if token != "" {
		authedHeader := "token " + token
		httpClient = httpcli.DoerFunc(func(req *http.Request) (*http.Response, error) {
			req.Header.Set("Authorization", authedHeader)
			return externalHTTPClient.Do(req)
		})
	}
	return &tracedClient{graphql.NewClient(endpoint, httpClient)}
}

type tracedClient struct{ c graphql.Client }

func (tc *tracedClient) MakeRequest(
	ctx context.Context,
	req *graphql.Request,
	resp *graphql.Response,
) error {
	span, ctx := trace.New(ctx, "DotComGraphQL."+req.OpName)

	err := tc.c.MakeRequest(ctx, req, resp)

	span.SetError(err)
	span.SetError(resp.Errors)
	span.End()

	return err
}
