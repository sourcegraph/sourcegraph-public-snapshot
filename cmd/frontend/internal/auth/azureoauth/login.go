package azureoauth

import (
	"fmt"
	"net/http"

	"github.com/dghubble/gologin/v2"
	oauth2Login "github.com/dghubble/gologin/v2/oauth2"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/azuredevops"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"golang.org/x/oauth2"
)

func loginHandler(c oauth2.Config) http.Handler {
	return oauth2Login.LoginHandler(&c, gologin.DefaultFailureHandler)
}

func azureDevOpsHandler(logger log.Logger, config *oauth2.Config, success, failure http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		token, err := oauth2Login.TokenFromContext(ctx)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}

		azureClient, err := azuredevops.NewClient(
			urnAzureDevOpsOAuth,
			azuredevops.VisualStudioAppURL,
			&auth.OAuthBearerToken{Token: token.AccessToken},
			nil,
		)
		if err != nil {
			logger.Error("failed to create azuredevops.Client", log.String("error", err.Error()))
			ctx = gologin.WithError(ctx, errors.Errorf("failed to create HTTP client for azuredevops with AuthURL %q", config.Endpoint.AuthURL))
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}

		profile, err := azureClient.GetAuthorizedProfile(ctx)
		if err != nil {
			msg := "failed to get Azure profile after oauth2 callback"
			logger.Error(msg, log.String("error", err.Error()))
			ctx = gologin.WithError(ctx, errors.Wrap(err, msg))
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}

		if profile.ID == "" || profile.EmailAddress == "" {
			msg := "bad Azure profile in API response"
			logger.Error(msg, log.String("profile", fmt.Sprintf("%#v", profile)))

			ctx = gologin.WithError(
				ctx,
				errors.Errorf("%s: %#v", msg, profile),
			)

			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}

		ctx = withUser(ctx, profile)
		success.ServeHTTP(w, req.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

// Adapted from "github.com/dghubble/gologin/oauth2"
//
// AzureDevOps expects some extra parameters in the POST body of the request to get the access
// token. Custom implementation is needed to be able to pass those as AuthCodeOption args to the
// config.Exchange method call.

// CallbackHandler handles OAuth2 redirection URI requests by parsing the auth
// code and state, comparing with the state value from the ctx, and obtaining
// an OAuth2 Token.
func callbackHandler(config *oauth2.Config, success http.Handler) http.Handler {
	failure := gologin.DefaultFailureHandler

	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		authCode, state, err := parseCallback(req)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		ownerState, err := oauth2Login.StateFromContext(ctx)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		if state != ownerState || state == "" {
			ctx = gologin.WithError(ctx, oauth2Login.ErrInvalidState)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}

		// Custom values in the POST body required by the API to get an access token. See:
		// https://learn.microsoft.com/en-us/azure/devops/integrate/get-started/authentication/oauth?view=azure-devops#http-request-body---authorize-app
		clientAssertionType := oauth2.SetAuthURLParam("client_assertion_type", azuredevops.ClientAssertionType)
		clientAssertion := oauth2.SetAuthURLParam("client_assertion", config.ClientSecret)
		grantType := oauth2.SetAuthURLParam("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
		assertion := oauth2.SetAuthURLParam("assertion", authCode)

		// Use the authorization code to get a Token.
		//
		// This will set the default value of "grant_type" to "authorization_code". But since we
		// pass a custom AuthCodeOption, it will overwrite that value.
		//
		// DEBUGGING NOTE: This also sets the authCode in the "code" URL arg, but we need to set the
		// auth code against the "assertion" URL arg. This means an extra arg in the form of
		// code=<auth-code> is also sent in the POST request. But it works. However if fetching the
		// access token breaks in the future by any chance without us having changed any code of our
		// own, this is a good place to start and writing a custom Exchange method to not send any
		// extra args.
		//
		// For now it works.
		token, err := config.Exchange(ctx, authCode, clientAssertionType, clientAssertion, grantType, assertion, oauth2.ApprovalForce)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		ctx = oauth2Login.WithToken(ctx, token)
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
