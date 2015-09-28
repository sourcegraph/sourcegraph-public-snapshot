package github

import (
	"strings"

	"github.com/sourcegraph/go-github/github"
	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/server/serverctx"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/util/githubutil"
)

func init() {
	// Make a GitHub API client available in the context that is
	// authenticated as the current user, or just using our
	// application credentials if there's no current user.
	serverctx.Funcs = append(serverctx.Funcs,
		NewContextWithAuthedClient,
	)
}

type contextKey int

const (
	minimalClientKey contextKey = iota
)

// NewContextWithClient creates a new child context with the specified
// GitHub client.
func NewContextWithClient(ctx context.Context, client *github.Client) context.Context {
	return newContext(ctx, newMinimalClient(client))
}

// NewContextWithUnauthedClient creates a new child context with a
// GitHub client that is authenticated using the application
// credentials but no user credentials.
func NewContextWithUnauthedClient(ctx context.Context) context.Context {
	return NewContextWithClient(ctx, githubutil.Default.UnauthedClient())
}

// NewContextWithAuthedClient creates a new child context with a
// GitHub client that is authenticated using the credentials of the
// context's actor, or unauthenticated if there is no actor (or if the
// actor has no stored GitHub credentials).
func NewContextWithAuthedClient(ctx context.Context) (context.Context, error) {
	a := auth.ActorFromContext(ctx)
	var c *github.Client
	if a.IsAuthenticated() {
		if s := store.ExternalAuthTokensFromContextOrNil(ctx); s != nil {
			host := strings.TrimPrefix(githubutil.Default.BaseURL.Host, "api.") // api.github.com -> github.com
			tok, err := s.GetUserToken(ctx, a.UID, host, githubutil.Default.OAuth.ClientID)
			if err == nil {
				c = githubutil.Default.AuthedClient(tok.Token)
			}
			if err != nil && err != auth.ErrNoExternalAuthToken && err != auth.ErrExternalAuthTokenDisabled {
				return nil, err
			}
		}
	}
	if c == nil {
		c = githubutil.Default.UnauthedClient()
	}
	return NewContextWithClient(ctx, c), nil
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
