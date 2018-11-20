package gitlaboauth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/dghubble/gologin"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/external/auth"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/auth/oauth"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/schema"
	"golang.org/x/oauth2"
)

const sessionKey = "gitlaboauth@0"

func parseProvider(callbackURL string, p *schema.GitLabAuthProvider, sourceCfg schema.AuthProviders) (provider *oauth.Provider, problems []string) {
	rawURL := p.Url
	if rawURL == "" {
		rawURL = "https://gitlab.com/"
	}
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		problems = append(problems, fmt.Sprintf("Could not parse GitLab URL %q. You will not be able to login via this GitLab instance.", rawURL))
		return nil, problems
	}
	codeHost := gitlab.NewCodeHost(parsedURL)
	oauth2Cfg := oauth2.Config{
		RedirectURL:  callbackURL,
		ClientID:     p.ClientID,
		ClientSecret: p.ClientSecret,
		Scopes:       []string{"api", "read_user"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  codeHost.BaseURL().ResolveReference(&url.URL{Path: "/oauth/authorize"}).String(),
			TokenURL: codeHost.BaseURL().ResolveReference(&url.URL{Path: "/oauth/token"}).String(),
		},
	}
	return oauth.NewProvider(oauth.ProviderOp{
		AuthPrefix:   authPrefix,
		OAuth2Config: oauth2Cfg,
		SourceConfig: sourceCfg,
		StateConfig:  getStateConfig(),
		ServiceID:    codeHost.ServiceID(),
		ServiceType:  codeHost.ServiceType(),
		Login:        LoginHandler(&oauth2Cfg, nil),
		Callback: CallbackHandler(&oauth2Cfg, oauth.SessionIssuer(
			sessionKey, codeHost.ServiceType(), codeHost.ServiceID(), p.ClientID, getOrCreateUser,
			func(w http.ResponseWriter) {
				stateConfig := getStateConfig()
				stateConfig.MaxAge = -1
				http.SetCookie(w, oauth.NewCookie(stateConfig, ""))
			},
		), nil),
	}), nil
}

func getOrCreateUser(ctx context.Context, serviceType, serviceID, clientID string, token *oauth2.Token) (actr *actor.Actor, safeErrMsg string, err error) {
	gUser, err := UserFromContext(ctx)
	if err != nil {
		return nil, "Could not read GitLab user from callback request.", errors.Wrap(err, "could not read user from context")
	}

	login, err := auth.NormalizeUsername(gUser.Username)
	if err != nil {
		return nil, fmt.Sprintf("Error normalizing the username %q. See https://docs.sourcegraph.com/admin/auth/#username-normalization.", login), err
	}

	var data extsvc.ExternalAccountData
	data.SetAccountData(gUser)
	data.SetAuthData(token)

	userID, safeErrMsg, err := auth.CreateOrUpdateUser(ctx, db.NewUser{
		Username:        login,
		Email:           gUser.Email,
		EmailIsVerified: gUser.Email != "",
		DisplayName:     gUser.Name,
		AvatarURL:       gUser.AvatarURL,
	}, extsvc.ExternalAccountSpec{
		ServiceType: serviceType,
		ServiceID:   serviceID,
		ClientID:    clientID,
		AccountID:   strconv.FormatInt(int64(gUser.ID), 10),
	}, data)
	if err != nil {
		return nil, safeErrMsg, err
	}
	return actor.FromUser(userID), "", nil
}

func getStateConfig() gologin.CookieConfig {
	cfg := gologin.CookieConfig{
		Name:     "gitlab-state-cookie",
		Path:     "/",
		MaxAge:   120, // 120 seconds
		HTTPOnly: true,
	}
	if conf.Get().TlsCert != "" {
		cfg.Secure = true
	}
	return cfg
}
