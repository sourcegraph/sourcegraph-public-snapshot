package github

import (
	"strings"

	"github.com/sourcegraph/go-github/github"
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/githubutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/traceutil"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/serverctx"
)

func init() {
	// Make a GitHub API client available in the context that is
	// authenticated as the current user, or just using our
	// application credentials if there's no current user.
	//
	// This appends to LastFuncs, not just Funcs, because it must be
	// run AFTER the actor has been stored in the context, because it
	// depends on the actor.
	serverctx.LastFuncs = append(serverctx.LastFuncs,
		NewContextWithAuthedClient,
	)
}

type contextKey int

const (
	minimalClientKey contextKey = iota
)

// NewContextWithMockClient creates a new mock client for testing purpose.
func NewContextWithMockClient(ctx context.Context, isAuthedUser bool, userClient *github.Client, appClient *github.Client, mockRepos githubRepos) context.Context {
	return newContext(ctx, &minimalClient{
		repos:             mockRepos,
		orgs:              userClient.Organizations,
		appAuthorizations: appClient.Authorizations,
		isAuthedUser:      isAuthedUser,
	})
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
func NewContextWithAuthedClient(ctx context.Context) (context.Context, error) {
	ghConf := *githubutil.Default
	ghConf.AppdashSpanID = traceutil.SpanIDFromContext(ctx)

	a := auth.ActorFromContext(ctx)
	var userClient *github.Client

	isAuthedUser := false
	if a.IsAuthenticated() {
		host := strings.TrimPrefix(githubutil.Default.BaseURL.Host, "api.") // api.github.com -> github.com
		tok, err := store.ExternalAuthTokensFromContext(ctx).GetUserToken(ctx, a.UID, host, githubutil.Default.OAuth.ClientID)
		if err == nil {
			userClient = ghConf.AuthedClient(tok.Token)
			isAuthedUser = true
		}
		if err != nil && err != store.ErrNoExternalAuthToken && err != store.ErrExternalAuthTokenDisabled {
			return nil, err
		}
	}
	if userClient == nil {
		userClient = ghConf.UnauthedClient()
	}
	return NewContextWithClient(ctx, isAuthedUser, userClient, ghConf.ApplicationAuthedClient()), nil
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
