package github

import (
	"context"

	"github.com/sourcegraph/go-github/github"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/githubutil"
)

type contextKey int

const (
	minimalClientKey contextKey = iota
	reposKey
)

// WithMockHasAuthedUser creates a new mock client that is nil and
// will panic on any operation except for HasAuthedUser, which reports
// the value of hasAuthedUser.
func WithMockHasAuthedUser(ctx context.Context, hasAuthedUser bool) context.Context {
	return newContext(ctx, &minimalClient{isAuthedUser: hasAuthedUser})
}

// NewContextWithClient creates a new child context with the specified
// GitHub clients. The userClient is authenticated as the user (or no
// user if there is none), and appClient is authenticated using the
// OAuth2 application's client ID and secret (which is required for
// certain GitHub API endpoints).
func NewContextWithClient(ctx context.Context, isAuthedUser bool, userClient *github.Client, appClient *github.Client) context.Context {
	return newContext(ctx, newMinimalClient(isAuthedUser, userClient, appClient))
}

// NewContextWithAuthedClient creates a new child context with a
// GitHub client that is authenticated using the credentials of the
// context's actor, or unauthenticated if there is no actor (or if the
// actor has no stored GitHub credentials).
func NewContextWithAuthedClient(ctx context.Context) context.Context {
	ghConf := *githubutil.Default
	ghConf.Context = ctx

	a := auth.ActorFromContext(ctx)
	var userClient *github.Client

	isAuthedUser := a.IsAuthenticated() && a.GitHubToken != ""
	if isAuthedUser {
		userClient = ghConf.AuthedClient(a.GitHubToken)
	} else {
		userClient = ghConf.UnauthedClient()
	}
	return NewContextWithClient(ctx, isAuthedUser, userClient, ghConf.ApplicationAuthedClient())
}

func newContext(ctx context.Context, client *minimalClient) context.Context {
	return context.WithValue(ctx, minimalClientKey, client)
}

// client returns the context's GitHub API client.
func client(ctx context.Context) *minimalClient {
	client, _ := ctx.Value(minimalClientKey).(*minimalClient)
	if client == nil {
		panic("no GitHub API client set in context")
	}
	return client
}
