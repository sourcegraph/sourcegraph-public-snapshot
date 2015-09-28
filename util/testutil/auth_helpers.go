package testutil

import (
	"net/url"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/auth/idkey"
	"sourcegraph.com/sourcegraph/sourcegraph/auth/sharedsecret"
)

// AddSystemAuthToURL adds credentials to urlStr (which is assumed to
// be an HTTP(S) URL) that authenticate requests as the system (i.e.,
// not as any particular user).
func AddSystemAuthToURL(ctx context.Context, urlStr string) (string, error) {
	src := sharedsecret.TokenSource(idkey.FromContext(ctx), "tmp")
	tok, err := src.Token()
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
