package githuboauth

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/dghubble/gologin/v2"
	"github.com/dghubble/gologin/v2/github"
	goauth2 "github.com/dghubble/gologin/v2/oauth2"
	"github.com/inconshreveable/log15" //nolint:logging // TODO move all logging to sourcegraph/log
	"golang.org/x/oauth2"

	"github.com/sourcegraph/sourcegraph/internal/dotcom"
	gh "github.com/sourcegraph/sourcegraph/internal/extsvc/github"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/oauth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/schema"
)

const sessionKey = "githuboauth@0"

func parseProvider(logger log.Logger, p *schema.GitHubAuthProvider, db database.DB, sourceCfg schema.AuthProviders) (provider *oauth.Provider, messages []string) {
	rawURL := p.GetURL()
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		messages = append(messages, fmt.Sprintf("Could not parse GitHub URL %q. You will not be able to login via this GitHub instance.", rawURL))
		return nil, messages
	}
	if !validateClientIDAndSecret(p.ClientID) {
		messages = append(messages, "GitHub client ID contains unexpected characters, possibly hidden")
	}
	if !validateClientIDAndSecret(p.ClientSecret) {
		messages = append(messages, "GitHub client secret contains unexpected characters, possibly hidden")
	}
	codeHost := extsvc.NewCodeHost(parsedURL, extsvc.TypeGitHub)

	callbackHandler := github.CallbackHandler
	if !gh.URLIsGitHubDotCom(parsedURL) {
		callbackHandler = github.EnterpriseCallbackHandler
	}

	return oauth.NewProvider(oauth.ProviderOp{
		AuthPrefix: authPrefix,
		OAuth2Config: func() oauth2.Config {
			return oauth2.Config{
				ClientID:     p.ClientID,
				ClientSecret: p.ClientSecret,
				Scopes:       requestedScopes(p),
				Endpoint: oauth2.Endpoint{
					AuthURL:  codeHost.BaseURL.ResolveReference(&url.URL{Path: "/login/oauth/authorize"}).String(),
					TokenURL: codeHost.BaseURL.ResolveReference(&url.URL{Path: "/login/oauth/access_token"}).String(),
				},
			}
		},
		SourceConfig: sourceCfg,
		ServiceID:    codeHost.ServiceID,
		ServiceType:  codeHost.ServiceType,
		Login: func(oauth2Cfg oauth2.Config) http.Handler {
			return github.LoginHandler(&oauth2Cfg, nil)
		},
		Callback: func(oauth2Cfg oauth2.Config) http.Handler {
			return callbackHandler(
				&oauth2Cfg,
				oauth.SessionIssuer(logger, db, &sessionIssuerHelper{
					CodeHost:     codeHost,
					logger:       log.Scoped("sessionIssuerHelper"),
					db:           db,
					clientID:     p.ClientID,
					allowSignup:  p.AllowSignup,
					allowOrgs:    p.AllowOrgs,
					allowOrgsMap: p.AllowOrgsMap,
				}, sessionKey),
				http.HandlerFunc(failureHandler),
			)
		},
	}), messages
}

func failureHandler(w http.ResponseWriter, r *http.Request) {
	// As a special case wa want to handle `access_denied` errors by redirecting
	// back. This case arises when the user decides not to proceed by clicking `cancel`.
	if err := r.URL.Query().Get("error"); err != "access_denied" {
		// Fall back to default failure handler
		gologin.DefaultFailureHandler.ServeHTTP(w, r)
		return
	}

	ctx := r.Context()
	encodedState, err := goauth2.StateFromContext(ctx)
	if err != nil {
		log15.Error("OAuth failed: could not get state from context.", "error", err)
		http.Error(w, "Authentication failed. Try signing in again (and clearing cookies for the current site). The error was: could not get OAuth state from context.", http.StatusInternalServerError)
		return
	}
	state, err := oauth.DecodeState(encodedState)
	if err != nil {
		log15.Error("OAuth failed: could not decode state.", "error", err)
		http.Error(w, "Authentication failed. Try signing in again (and clearing cookies for the current site). The error was: could not get decode OAuth state.", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, auth.SafeRedirectURL(state.Redirect), http.StatusFound)
}

var clientIDSecretValidator = lazyregexp.New("^[a-zA-Z0-9.]*$")

func validateClientIDAndSecret(clientIDOrSecret string) (valid bool) {
	return clientIDSecretValidator.MatchString(clientIDOrSecret)
}

func requestedScopes(p *schema.GitHubAuthProvider) []string {
	scopes := []string{"user:email"}
	if !dotcom.SourcegraphDotComMode() {
		scopes = append(scopes, "repo")
	}

	// Needs extra scope to check organization membership
	if len(p.AllowOrgs) > 0 || p.AllowGroupsPermissionsSync {
		scopes = append(scopes, "read:org")
	}

	return scopes
}
