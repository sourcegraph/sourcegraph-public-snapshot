package github

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/dghubble/gologin/v2"
	oauth2Login "github.com/dghubble/gologin/v2/oauth2"
	"github.com/google/go-github/v48/github"
	"golang.org/x/oauth2"
)

// Github login errors
var (
	ErrUnableToGetGithubUser = errors.New("github: unable to get Github User")
)

// StateHandler checks for a state cookie. If found, the state value is read
// and added to the ctx. Otherwise, a non-guessable value is added to the ctx
// and to a (short-lived) state cookie issued to the requester.
//
// Implements OAuth 2 RFC 6749 10.12 CSRF Protection. If you wish to issue
// state params differently, write a http.Handler which sets the ctx state,
// using oauth2 WithState(ctx, state) since it is required by LoginHandler
// and CallbackHandler.
func StateHandler(config gologin.CookieConfig, success http.Handler) http.Handler {
	return oauth2Login.StateHandler(config, success)
}

// LoginHandler handles Github login requests by reading the state value from
// the ctx and redirecting requests to the AuthURL with that state value.
func LoginHandler(config *oauth2.Config, failure http.Handler) http.Handler {
	return oauth2Login.LoginHandler(config, failure)
}

// CallbackHandler handles Github redirection URI requests and adds the Github
// access token and User to the ctx. If authentication succeeds, handling
// delegates to the success handler, otherwise to the failure handler.
func CallbackHandler(config *oauth2.Config, success, failure http.Handler) http.Handler {
	success = githubHandler(config, false, success, failure)
	return oauth2Login.CallbackHandler(config, success, failure)
}

// EnterpriseCallbackHandler handles Github Enterprise redirection URI requests
// and adds the Github access token and User to the ctx. If authentication
// succeeds,handling delegates to the success handler, otherwise to the failure
// handler. The Github Enterprise API URL is inferred from the OAuth2 config's
// AuthURL endpoint.
func EnterpriseCallbackHandler(config *oauth2.Config, success, failure http.Handler) http.Handler {
	success = githubHandler(config, true, success, failure)
	return oauth2Login.CallbackHandler(config, success, failure)
}

// githubHandler is a http.Handler that gets the OAuth2 Token from the ctx to
// get the corresponding Github User. If successful, the User is added to the
// ctx and the success handler is called. Otherwise, the failure handler is
// called.
func githubHandler(config *oauth2.Config, isEnterprise bool, success, failure http.Handler) http.Handler {
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

		httpClient := config.Client(ctx, token)
		var githubClient *github.Client
		if isEnterprise {
			githubClient, err = enterpriseGithubClientFromAuthURL(config.Endpoint.AuthURL, httpClient)
			if err != nil {
				ctx = gologin.WithError(ctx, fmt.Errorf("github: error creating Client: %v", err))
				failure.ServeHTTP(w, req.WithContext(ctx))
				return
			}
		} else {
			githubClient = github.NewClient(httpClient)
		}
		user, resp, err := githubClient.Users.Get(ctx, "")
		err = validateResponse(user, resp, err)
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

// validateResponse returns an error if the given Github user, raw
// http.Response, or error are unexpected. Returns nil if they are valid.
func validateResponse(user *github.User, resp *github.Response, err error) error {
	if err != nil || resp.StatusCode != http.StatusOK {
		return ErrUnableToGetGithubUser
	}
	if user == nil || user.ID == nil {
		return ErrUnableToGetGithubUser
	}
	return nil
}

// enterpriseGithubClientFromAuthURL returns a Github client that targets a GHE instance.
func enterpriseGithubClientFromAuthURL(authURL string, httpClient *http.Client) (*github.Client, error) {
	client := github.NewClient(httpClient)

	// convert authURL to GHE baseURL https://<mygithub>.com/api/v3/
	baseURL, err := url.Parse(authURL)
	if err != nil {
		return nil, fmt.Errorf("github: error parsing Endoint.AuthURL: %s", authURL)
	}

	baseURL.Path = "/api/v3/"
	client.BaseURL = baseURL
	client.UploadURL = baseURL

	return client, nil
}
