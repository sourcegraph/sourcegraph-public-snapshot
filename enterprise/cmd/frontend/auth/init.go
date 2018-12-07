// Package auth is imported for side-effects to enable enterprise-only SSO.
package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/external/app"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/auth/githuboauth"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/auth/gitlaboauth"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/auth/httpheader"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/auth/openidconnect"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/auth/saml"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/licensing"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	log15 "gopkg.in/inconshreveable/log15.v2"
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
	for _, p := range conf.Get().AuthProviders {
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

func init() {
	// Warn about usage of user auth providers (those that integrate with external services' ACLs,
	// such as GitHub or GitLab repository permissions) that are not enabled by the license.
	graphqlbackend.AlertFuncs = append(graphqlbackend.AlertFuncs, func(args graphqlbackend.AlertFuncArgs) []*graphqlbackend.Alert {
		// Only site admins can act on this alert, so only show it to site admins.
		if !args.IsSiteAdmin {
			return nil
		}

		if licensing.IsFeatureEnabledLenient(licensing.FeatureACLs) {
			return nil
		}

		var disallowedAuthProviderTypes []string
		for _, p := range conf.Get().AuthProviders {
			if p.Github != nil || p.Gitlab != nil {
				disallowedAuthProviderTypes = append(disallowedAuthProviderTypes, conf.AuthProviderType(p))
			}
		}
		if len(disallowedAuthProviderTypes) > 0 {
			return []*graphqlbackend.Alert{
				{
					TypeValue:    graphqlbackend.AlertTypeError,
					MessageValue: fmt.Sprintf("A Sourcegraph license is required for the following user authentication providers (which integrate with ACLs/permissions on external services): %s. [**Get a license.**](/site-admin/license)", strings.Join(disallowedAuthProviderTypes, ", ")),
				},
			}
		}
		return nil
	})
}
