package gitlaboauth

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/dghubble/gologin"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/auth/oauth"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/schema"
)

const sessionKey = "gitlaboauth@0"

func parseProvider(logger log.Logger, db database.DB, callbackURL string, p *schema.GitLabAuthProvider, sourceCfg schema.AuthProviders) (provider *oauth.Provider, messages []string) {
	rawURL := p.Url
	if rawURL == "" {
		rawURL = "https://gitlab.com/"
	}
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		messages = append(messages, fmt.Sprintf("Could not parse GitLab URL %q. You will not be able to login via this GitLab instance.", rawURL))
		return nil, messages
	}
	codeHost := extsvc.NewCodeHost(parsedURL, extsvc.TypeGitLab)

	return oauth.NewProvider(oauth.ProviderOp{
		AuthPrefix: authPrefix,
		OAuth2Config: func(extraScopes ...string) oauth2.Config {
			return oauth2.Config{
				RedirectURL:  callbackURL,
				ClientID:     p.ClientID,
				ClientSecret: p.ClientSecret,
				Scopes:       gitlab.RequestedOAuthScopes(p.ApiScope, extraScopes),
				Endpoint: oauth2.Endpoint{
					AuthURL:  codeHost.BaseURL.ResolveReference(&url.URL{Path: "/oauth/authorize"}).String(),
					TokenURL: codeHost.BaseURL.ResolveReference(&url.URL{Path: "/oauth/token"}).String(),
				},
			}
		},
		SourceConfig: sourceCfg,
		StateConfig:  getStateConfig(),
		ServiceID:    codeHost.ServiceID,
		ServiceType:  codeHost.ServiceType,
		Login: func(oauth2Cfg oauth2.Config) http.Handler {
			return LoginHandler(&oauth2Cfg, nil)
		},
		Callback: func(oauth2Cfg oauth2.Config) http.Handler {
			return CallbackHandler(
				&oauth2Cfg,
				oauth.SessionIssuer(logger, db, &sessionIssuerHelper{
					db:          db,
					CodeHost:    codeHost,
					clientID:    p.ClientID,
					allowSignup: p.AllowSignup,
					allowGroups: p.AllowGroups,
				}, sessionKey),
				nil,
			)
		},
	}), messages
}

func getStateConfig() gologin.CookieConfig {
	cfg := gologin.CookieConfig{
		Name:     "gitlab-state-cookie",
		Path:     "/",
		MaxAge:   900, // 15 minutes
		HTTPOnly: true,
		Secure:   conf.IsExternalURLSecure(),
	}
	return cfg
}
