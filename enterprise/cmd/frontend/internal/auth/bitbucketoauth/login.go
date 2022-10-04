package bitbucketoauth

import (
	"net/http"
	"net/url"

	"github.com/dghubble/gologin"
	oauth2Login "github.com/dghubble/gologin/oauth2"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketcloud"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func LoginHandler(config *oauth2.Config, failure http.Handler) http.Handler {
	return oauth2Login.LoginHandler(config, failure)
}

func CallbackHandler(config *oauth2.Config, success, failure http.Handler) http.Handler {
	success = bitbucketHandler(config, success, failure)
	return oauth2Login.CallbackHandler(config, success, failure)
}

func bitbucketHandler(config *oauth2.Config, success, failure http.Handler) http.Handler {
	logger := log.Scoped("BitbucketOAuthHandler", "Bitbucket OAuth Handler")

	if failure == nil {
		failure = gologin.DefaultFailureHandler
	}
	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		token, err := oauth2Login.TokenFromContext(ctx)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}

		bitbucketClient, err := bitbucketClientFromAuthURL(config.Endpoint.AuthURL, token.AccessToken)
		if err != nil {
			ctx = gologin.WithError(ctx, errors.Errorf("could not parse AuthURL %s", config.Endpoint.AuthURL))
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		user, err := bitbucketClient.GetUser(ctx, "")
		err = validateResponse(user, err)
		if err != nil {
			// TODO: Prefer a more general purpose fix, potentially
			// https://github.com/sourcegraph/sourcegraph/pull/20000
			logger.Warn("invalid response", log.Error(err))
		}
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		ctx = WithUser(ctx, user)
		success.ServeHTTP(w, req.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

// validateResponse returns an error if the given GitLab user or error are unexpected. Returns nil
// if they are valid.
func validateResponse(user *bitbucketcloud.User, err error) error {
	if err != nil {
		return errors.Wrap(err, "unable to get Bitbucket user")
	}
	if user == nil || user.UUID == "" {
		return errors.Errorf("unable to get Bitbucket user: bad user info %#+v", user)
	}
	return nil
}

func bitbucketClientFromAuthURL(authURL, oauthToken string) (*bitbucketcloud.Client, error) {
	baseURL, err := url.Parse(authURL)
	if err != nil {
		return nil, err
	}
	baseURL.Path = ""
	baseURL.RawQuery = ""
	baseURL.Fragment = ""
	return bitbucketcloud.NewClientProvider(extsvc.URNBitbucketOAuth, baseURL, nil).GetOAuthClient(oauthToken), nil
}
