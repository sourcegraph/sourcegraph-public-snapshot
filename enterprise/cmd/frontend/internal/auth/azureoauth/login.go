package azureoauth

import (
	"net/http"

	"github.com/dghubble/gologin"
	oauth2gologin "github.com/dghubble/gologin/oauth2"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

// Adapted from "github.com/dghubble/gologin/oauth2"

// CallbackHandler handles OAuth2 redirection URI requests by parsing the auth
// code and state, comparing with the state value from the ctx, and obtaining
// an OAuth2 Token.
func CallbackHandler(config *oauth2.Config, success, failure http.Handler) http.Handler {
	if failure == nil {
		failure = gologin.DefaultFailureHandler
	}

	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		authCode, state, err := parseCallback(req)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		ownerState, err := oauth2gologin.StateFromContext(ctx)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		if state != ownerState || state == "" {
			ctx = gologin.WithError(ctx, oauth2gologin.ErrInvalidState)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}

		// https://learn.microsoft.com/en-us/azure/devops/integrate/get-started/authentication/oauth?view=azure-devops#http-request-body---authorize-app
		clientAssertionType := oauth2.SetAuthURLParam("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
		clientAssertion := oauth2.SetAuthURLParam("client_assertion", config.ClientSecret)
		grantType := oauth2.SetAuthURLParam("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
		assertion := oauth2.SetAuthURLParam("assertion", authCode)

		// use the authorization code to get a Token
		token, err := config.Exchange(ctx, authCode, clientAssertionType, clientAssertion, grantType, assertion, oauth2.ApprovalForce)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		ctx = oauth2gologin.WithToken(ctx, token)
		success.ServeHTTP(w, req.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

// parseCallback parses the "code" and "state" parameters from the http.Request
// and returns them.
func parseCallback(req *http.Request) (authCode, state string, err error) {
	err = req.ParseForm()
	if err != nil {
		return "", "", err
	}
	authCode = req.Form.Get("code")
	state = req.Form.Get("state")
	if authCode == "" || state == "" {
		return "", "", errors.New("oauth2: Request missing code or state")
	}
	return authCode, state, nil
}
