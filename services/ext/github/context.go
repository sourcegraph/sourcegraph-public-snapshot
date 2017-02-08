package github

import (
	"context"

	"github.com/sourcegraph/go-github/github"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/githubutil"
)

type contextKey int

const (
	clientKey contextKey = iota
	reposKey
)

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
	return newContext(ctx, userClient)
}

func newContext(ctx context.Context, client *github.Client) context.Context {
	return context.WithValue(ctx, clientKey, client)
}

// client returns the context's GitHub API client.
func client(ctx context.Context) *github.Client {
	client, _ := ctx.Value(clientKey).(*github.Client)
	if client == nil {
		panic("no GitHub API client set in context")
	}
	return client
}

func OrgsFromContext(ctx context.Context) *github.OrganizationsService {
	return client(ctx).Organizations
}
