package bitbucketcloudoauth

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/dghubble/gologin/v2"
	"github.com/dghubble/gologin/v2/bitbucket"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/external/session"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/oauth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/schema"
)

const (
	sessionKey        = "bitbucketcloudoauth@0"
	defaultBBCloudURL = "https://bitbucket.org"
)

func parseProvider(logger log.Logger, p *schema.BitbucketCloudAuthProvider, db database.DB, sourceCfg schema.AuthProviders) (provider *oauth.Provider, messages []string) {
	rawURL := p.Url
	if rawURL == "" {
		rawURL = defaultBBCloudURL
	}
	parsedURL, err := url.Parse(rawURL)
	parsedURL = extsvc.NormalizeBaseURL(parsedURL)
	if err != nil {
		messages = append(messages, fmt.Sprintf("Could not parse Bitbucket Cloud URL %q. You will not be able to login via Bitbucket Cloud.", rawURL))
		return nil, messages
	}

	if !validateClientKeyOrSecret(p.ClientKey) {
		messages = append(messages, "Bitbucket Cloud key contains unexpected characters, possibly hidden")
	}
	if !validateClientKeyOrSecret(p.ClientSecret) {
		messages = append(messages, "Bitbucket Cloud secret contains unexpected characters, possibly hidden")
	}

	return oauth.NewProvider(oauth.ProviderOp{
		AuthPrefix: authPrefix,
		OAuth2Config: func() oauth2.Config {
			return oauth2.Config{
				ClientID:     p.ClientKey,
				ClientSecret: p.ClientSecret,
				Scopes:       requestedScopes(p.ApiScope),
				Endpoint: oauth2.Endpoint{
					AuthURL:  parsedURL.ResolveReference(&url.URL{Path: "/site/oauth2/authorize"}).String(),
					TokenURL: parsedURL.ResolveReference(&url.URL{Path: "/site/oauth2/access_token"}).String(),
				},
			}
		},
		SourceConfig: sourceCfg,
		ServiceID:    parsedURL.String(),
		ServiceType:  extsvc.TypeBitbucketCloud,
		Login: func(oauth2Cfg oauth2.Config) http.Handler {
			return bitbucket.LoginHandler(&oauth2Cfg, nil)
		},
		Callback: func(oauth2Cfg oauth2.Config) http.Handler {
			return bitbucket.CallbackHandler(
				&oauth2Cfg,
				oauth.SessionIssuer(logger, db, &sessionIssuerHelper{
					baseURL:     parsedURL,
					db:          db,
					clientKey:   p.ClientKey,
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

var clientKeySecretValidator = lazyregexp.New("^[a-zA-Z0-9.]*$")

func validateClientKeyOrSecret(clientKeyOrSecret string) (valid bool) {
	return clientKeySecretValidator.MatchString(clientKeyOrSecret)
}

func requestedScopes(apiScopes string) []string {
	return strings.Split(apiScopes, ",")
}
