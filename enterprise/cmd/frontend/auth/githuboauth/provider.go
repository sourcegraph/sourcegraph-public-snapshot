package githuboauth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/dghubble/gologin/github"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/external/auth"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/auth/oauth"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc"
	"github.com/sourcegraph/sourcegraph/schema"
	"golang.org/x/oauth2"
)

const sessionKey = "githuboauth@0"

func parseProvider(p *schema.GitHubAuthProvider, sourceCfg schema.AuthProviders) (provider *oauth.Provider, problems []string) {
	rawURL := p.Url
	if rawURL == "" {
		rawURL = "https://github.com/"
	}
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		problems = append(problems, fmt.Sprintf("Could not parse GitHub URL %q. You will not be able to login via this GitHub instance.", rawURL))
		return nil, problems
	}
	baseURL := extsvc.NormalizeBaseURL(parsedURL).String()
	oauth2Cfg := oauth2.Config{
		ClientID:     p.ClientID,
		ClientSecret: p.ClientSecret,
		Scopes:       []string{"repo"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  strings.TrimSuffix(baseURL, "/") + "/login/oauth/authorize",
			TokenURL: strings.TrimSuffix(baseURL, "/") + "/login/oauth/access_token",
		},
	}
	return oauth.NewProvider(oauth.ProviderOp{
		AuthPrefix:   authPrefix,
		OAuth2Config: oauth2Cfg,
		SourceConfig: sourceCfg,
		StateConfig:  getStateConfig(),
		ServiceID:    baseURL,
		ServiceType:  serviceType,
		Login:        github.LoginHandler(&oauth2Cfg, nil),
		Callback: github.CallbackHandler(&oauth2Cfg, oauth.SessionIssuer(sessionKey, serviceType, baseURL, p.ClientID, getOrCreateUser, func(w http.ResponseWriter) {
			stateConfig := getStateConfig()
			stateConfig.MaxAge = -1
			http.SetCookie(w, oauth.NewCookie(stateConfig, ""))
		}), nil),
	}), nil
}

func getOrCreateUser(ctx context.Context, serviceType, serviceID, clientID string, token *oauth2.Token) (actr *actor.Actor, safeErrMsg string, err error) {
	ghUser, err := github.UserFromContext(ctx)
	if err != nil {
		return nil, "Could not read GitHub user from callback request.", errors.Wrap(err, "could not read user from context")
	}

	login, err := auth.NormalizeUsername(deref(ghUser.Login))
	if err != nil {
		return nil, fmt.Sprintf("Error normalizing the username %q. See https://docs.sourcegraph.com/admin/auth/#username-normalization.", login), err
	}

	var data extsvc.ExternalAccountData
	data.SetAccountData(ghUser)
	data.SetAuthData(token)
	userID, safeErrMsg, err := auth.CreateOrUpdateUser(ctx, db.NewUser{
		Username:        login,
		Email:           deref(ghUser.Email),
		EmailIsVerified: deref(ghUser.Email) != "",
		DisplayName:     deref(ghUser.Name),
		AvatarURL:       deref(ghUser.AvatarURL),
	}, extsvc.ExternalAccountSpec{
		ServiceType: serviceType,
		ServiceID:   serviceID,
		ClientID:    clientID,
		AccountID:   strconv.FormatInt(derefInt64(ghUser.ID), 10),
	}, data)
	if err != nil {
		return nil, safeErrMsg, err
	}
	return actor.FromUser(userID), "", nil
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func derefInt64(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}
