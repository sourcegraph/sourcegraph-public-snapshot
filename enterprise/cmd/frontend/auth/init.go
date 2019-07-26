// Package auth is imported for side-effects to enable enterprise-only SSO.
package auth

import (
	"net/http"

	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/cmd/frontend/auth"
	"sourcegraph.com/cmd/frontend/external/app"
	"sourcegraph.com/enterprise/cmd/frontend/auth/githuboauth"
	"sourcegraph.com/enterprise/cmd/frontend/auth/gitlaboauth"
	"sourcegraph.com/enterprise/cmd/frontend/auth/httpheader"
	"sourcegraph.com/enterprise/cmd/frontend/auth/openidconnect"
	"sourcegraph.com/enterprise/cmd/frontend/auth/saml"
	"sourcegraph.com/pkg/conf"
)

func init() {
	// Register enterprise auth middleware
	auth.RegisterMiddlewares(
		openidconnect.Middleware,
		saml.Middleware,
		httpheader.Middleware,
		githuboauth.Middleware,
		gitlaboauth.Middleware,
	)
	// Register app-level sign-out handler
	app.RegisterSSOSignOutHandler(ssoSignOutHandler)
}

func ssoSignOutHandler(w http.ResponseWriter, r *http.Request) (signOutURLs []app.SignOutURL) {
	for _, p := range conf.Get().Critical.AuthProviders {
		var e app.SignOutURL
		var err error
		switch {
		case p.Openidconnect != nil:
			e.ProviderDisplayName = p.Openidconnect.DisplayName
			e.ProviderServiceType = p.Openidconnect.Type
			e.URL, err = openidconnect.SignOut(w, r)
		case p.Saml != nil:
			e.ProviderDisplayName = p.Saml.DisplayName
			e.ProviderServiceType = p.Saml.Type
			e.URL, err = saml.SignOut(w, r)
		case p.Github != nil:
			e.ProviderDisplayName = p.Github.DisplayName
			e.ProviderServiceType = p.Github.Type
			e.URL, err = githuboauth.SignOutURL(p.Github.Url)
		case p.Gitlab != nil:
			e.ProviderDisplayName = p.Gitlab.DisplayName
			e.ProviderServiceType = p.Gitlab.Type
			e.URL, err = gitlaboauth.SignOutURL(p.Gitlab.Url)
		}
		if e.URL != "" {
			signOutURLs = append(signOutURLs, e)
		}
		if err != nil {
			log15.Error("Error clearing auth provider session data.", "err", err)
		}
	}

	return signOutURLs
}
