package bitbucketserveroauth

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/dghubble/gologin/v2"
	"github.com/dghubble/gologin/v2/bitbucket"
	oauth2Login "github.com/dghubble/gologin/v2/oauth2"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/oauth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/session"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/schema"
	"golang.org/x/oauth2"
)

const (
	sessionKey    = "bitbucketserveroauth@0"
	authorizePath = "/rest/oauth2/latest/authorize"
	tokenPath     = "/rest/oauth2/latest/token"
	redirectPath  = authPrefix + "/callback"
)

func parseProvider(logger log.Logger, p *schema.BitbucketServerAuthProvider, db database.DB, sourceCfg schema.AuthProviders) (provider *oauth.Provider, messages []string) {
	parsedURL, err := url.Parse(p.Url)
	if err != nil {
		messages = append(messages, fmt.Sprintf("Could not parse Bitbucket Server URL %q. You will not be able to login via Bitbucket Server.", p.Url))
		return nil, messages
	}

	parsedURL = extsvc.NormalizeBaseURL(parsedURL)

	extURL, err := url.Parse(conf.ExternalURL())
	if err != nil {
		messages = append(messages, fmt.Sprintf("Could not parse Sourcegraph external URL %q.", conf.ExternalURL()))
		return nil, messages
	}

	authURL := parsedURL.ResolveReference(&url.URL{Path: authorizePath}).String()
	tokenURL := parsedURL.ResolveReference(&url.URL{Path: tokenPath}).String()
	redirectURL := extURL.ResolveReference(&url.URL{Path: redirectPath}).String()

	return oauth.NewProvider(oauth.ProviderOp{
		AuthPrefix: authPrefix,
		OAuth2Config: func() oauth2.Config {
			return oauth2.Config{
				ClientID:     p.ClientID,
				ClientSecret: p.ClientSecret,
				Endpoint: oauth2.Endpoint{
					AuthURL:  authURL,
					TokenURL: tokenURL,
				},
				Scopes:      []string{"REPO_READ"},
				RedirectURL: redirectURL,
			}
		},
		SourceConfig: sourceCfg,
		ServiceID:    parsedURL.String(),
		ServiceType:  extsvc.TypeBitbucketServer,
		Login: func(oauth2Cfg oauth2.Config) http.Handler {
			return bitbucket.LoginHandler(&oauth2Cfg, nil)
		},
		Callback: func(oauth2Cfg oauth2.Config) http.Handler {
			return oauth2Login.CallbackHandler(
				&oauth2Cfg,
				oauth.SessionIssuer(logger, db, &sessionIssuerHelper{
					logger:      logger.Scoped("sessionIssuerHelper"),
					baseURL:     parsedURL,
					db:          db,
					clientKey:   p.ClientID,
					allowSignup: p.AllowSignup,
				}, sessionKey),
				http.HandlerFunc(failureHandler),
			)
		},
	}), messages
}

func failureHandler(w http.ResponseWriter, r *http.Request) {
	// As a special case we want to handle `access_denied` errors by redirecting
	// back. This case arises when the user decides not to proceed by clicking `cancel`.
	if err := r.URL.Query().Get("error"); err != "access_denied" {
		// Fall back to default failure handler
		gologin.DefaultFailureHandler.ServeHTTP(w, r)
		return
	}

	var encodedState string
	err := session.GetData(r, "oauthState", &encodedState)
	if err != nil {
		http.Error(w, "Authentication failed. Try signing in again (and clearing cookies for the current site). The error was: could not get OAuth state from context.", http.StatusInternalServerError)
		return
	}
	state, err := oauth.DecodeState(encodedState)
	if err != nil {
		http.Error(w, "Authentication failed. Try signing in again (and clearing cookies for the current site). The error was: could not get decode OAuth state.", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, auth.SafeRedirectURL(state.Redirect), http.StatusFound)
}
