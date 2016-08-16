package authutil

import (
	"net/url"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth/idkey"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth/sharedsecret"
)

// AddSystemAuthToURL adds credentials to urlStr (which is assumed to
// be an HTTP(S) URL) that authenticate requests as the system (i.e.,
// not as any particular user).
func AddSystemAuthToURL(ctx context.Context, scope, urlStr string) (string, error) {
	k := idkey.FromContext(ctx)
	if k == nil {
		panic("no ID key in context")
	}
	return AddAuthToURL(sourcegraph.WithCredentials(ctx, sharedsecret.TokenSource(k, scope)), urlStr)
}

// AddAuthToURL adds credentials to urlStr (which is assumed to
// be an HTTP(S) URL) from the provided context's auth.
func AddAuthToURL(ctx context.Context, urlStr string) (string, error) {
	tok, err := sourcegraph.CredentialsFromContext(ctx).Token()
	if err != nil {
		return "", err
	}
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}
	u.User = url.UserPassword("x-oauth-basic", string(tok.AccessToken))
	return u.String(), nil
}
